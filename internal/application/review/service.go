package review

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wyw14/cry-092/internal/application"
	"github.com/wyw14/cry-092/internal/domain/proposal"
	domain "github.com/wyw14/cry-092/internal/domain/review"
	"github.com/wyw14/cry-092/internal/domain/shared"
	"github.com/wyw14/cry-092/internal/domain/supervision"
)

type Service struct {
	Reviews     application.ReviewRepository
	Replies     application.ReplyRepository
	Assignments application.AssignmentRepository
	Proposals   application.ProposalRepository
	Supervision application.SupervisionRepository
	Audits      application.AuditRepository
	Outbox      application.Outbox
	Tx          application.TransactionManager
	Clock       application.Clock
	IDs         application.IDGenerator
}

func (s Service) Evaluate(ctx context.Context, proposalID, replyID, representativeID string, rating domain.Rating, comment string) (*domain.Evaluation, error) {
	evaluation, err := s.prepareEvaluation(ctx, proposalID, replyID, representativeID, rating, comment)
	if err != nil {
		return nil, fmt.Errorf("prepare evaluation: %w", err)
	}
	if err := s.Reviews.SaveEvaluation(ctx, evaluation); err != nil {
		return nil, fmt.Errorf("save evaluation: %w", err)
	}
	err = s.Tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		p, err := s.Proposals.Get(txCtx, proposalID)
		if err != nil {
			return err
		}
		if p.RepresentativeID != representativeID {
			return shared.ErrForbidden
		}
		reply, err := s.Replies.GetReply(txCtx, replyID)
		if err != nil {
			return err
		}
		if reply.ProposalID != proposalID {
			return shared.ErrConflict
		}
		version := p.Version
		if err := p.Advance(proposal.StatusAnswered, proposal.StatusEvaluated, s.Clock.Now()); err != nil {
			return err
		}
		if rating == domain.RatingUnsatisfied {
			a, err := s.Assignments.GetByProposal(txCtx, proposalID)
			if err != nil {
				return err
			}
			rework := domain.NewReworkRound(s.IDs.NewID(), proposalID, replyID, a.UnitID, comment, reply.Round+1, s.Clock.Now())
			if err := s.Reviews.SaveRework(txCtx, rework); err != nil {
				return err
			}
			caseValue := supervision.NewCase(s.IDs.NewID(), proposalID, "supervision-duty", s.Clock.Now())
			if err := s.Supervision.SaveCase(txCtx, caseValue); err != nil {
				return err
			}
			payload, _ := json.Marshal(map[string]string{"proposal_id": proposalID, "evaluation_id": evaluation.ID, "case_id": caseValue.ID})
			if err := s.Outbox.Enqueue(txCtx, s.IDs.NewID(), "evaluation.unsatisfied", payload); err != nil {
				return err
			}
		}
		if err := s.Proposals.Save(txCtx, p, version); err != nil {
			return err
		}
		return s.Audits.Append(txCtx, shared.NewAuditEvent(s.IDs.NewID(), "evaluation", evaluation.ID, representativeID, shared.AuditHTTP, nil, map[string]any{"rating": rating, "reply_id": replyID}, "representative evaluated reply", s.Clock.Now()))
	})
	if err != nil {
		return nil, fmt.Errorf("evaluate reply: %w", err)
	}
	return evaluation, nil
}

func (s Service) prepareEvaluation(ctx context.Context, proposalID, replyID, representativeID string, rating domain.Rating, comment string) (*domain.Evaluation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p, err := s.Proposals.Get(ctx, proposalID)
	if err != nil {
		return nil, err
	}
	if p.RepresentativeID != representativeID {
		return nil, shared.ErrForbidden
	}
	reply, err := s.Replies.GetReply(ctx, replyID)
	if err != nil {
		return nil, err
	}
	if reply.ProposalID != proposalID {
		return nil, shared.ErrConflict
	}
	return domain.NewEvaluation(s.IDs.NewID(), proposalID, replyID, representativeID, reply.Round, rating, comment, s.Clock.Now())
}
