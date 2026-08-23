package response

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-092/internal/application"
	"github.com/wyw14/cry-092/internal/domain/assignment"
	"github.com/wyw14/cry-092/internal/domain/handling"
	"github.com/wyw14/cry-092/internal/domain/proposal"
	domain "github.com/wyw14/cry-092/internal/domain/response"
	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Service struct {
	Replies     application.ReplyRepository
	Plans       application.HandlingRepository
	Assignments application.AssignmentRepository
	Proposals   application.ProposalRepository
	Audits      application.AuditRepository
	Tx          application.TransactionManager
	Clock       application.Clock
	IDs         application.IDGenerator
}

func (s Service) Submit(ctx context.Context, proposalID, unitID, actorID string, round int, kind domain.Kind, summary string, files []string) (*domain.Reply, error) {
	var reply *domain.Reply
	kind, summary, files = prepareReplyInput(kind, summary, files)
	err := s.Tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		a, err := s.Assignments.GetByProposal(txCtx, proposalID)
		if err != nil {
			return err
		}
		if a.Status != assignment.StatusAccepted || !a.CanHandle(unitID) {
			return shared.ErrForbidden
		}
		plan, err := s.Plans.GetPlan(txCtx, proposalID)
		if err != nil {
			return err
		}
		if plan.Status == handling.PlanCompleted {
			return shared.ErrConflict
		}
		p, err := s.Proposals.Get(txCtx, proposalID)
		if err != nil {
			return err
		}
		version := p.Version
		reply, err = domain.NewReply(s.IDs.NewID(), proposalID, actorID, round, kind, summary, files, s.Clock.Now())
		if err != nil {
			return err
		}
		if err := plan.Complete(s.Clock.Now()); err != nil {
			return err
		}
		if err := p.Advance(proposal.StatusHandling, proposal.StatusAnswered, s.Clock.Now()); err != nil {
			return err
		}
		if err := s.Replies.SaveReply(txCtx, reply); err != nil {
			return err
		}
		if err := s.Plans.SavePlan(txCtx, plan); err != nil {
			return err
		}
		if err := s.Proposals.Save(txCtx, p, version); err != nil {
			return err
		}
		return s.Audits.Append(txCtx, shared.NewAuditEvent(s.IDs.NewID(), "reply", reply.ID, actorID, shared.AuditHTTP, nil, map[string]any{"kind": kind, "round": round}, "unit submitted formal reply", s.Clock.Now()))
	})
	if err != nil {
		return nil, fmt.Errorf("submit reply: %w", err)
	}
	return reply, nil
}

func prepareReplyInput(kind domain.Kind, summary string, files []string) (domain.Kind, string, []string) {
	if kind == "" {
		kind = domain.KindExplained
	}
	// 不复用入参切片底层空间（files[:0] 会改写调用方原数组），结果构建到独立切片，保证调用方附件数组在提交后不被就地修改。
	kept := make([]string, 0, len(files))
	for _, fileID := range files {
		if fileID != "" {
			kept = append(kept, fileID)
		}
	}
	if len(kept) > 1 {
		kept = kept[:1]
	}
	return kind, summary, kept
}
