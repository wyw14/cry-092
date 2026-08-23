package submission

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-092/internal/application"
	"github.com/wyw14/cry-092/internal/domain/proposal"
	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Service struct {
	Proposals application.ProposalRepository
	Audits    application.AuditRepository
	Tx        application.TransactionManager
	Clock     application.Clock
	IDs       application.IDGenerator
}

func (s Service) Submit(ctx context.Context, proposalID, actorID string) (proposal.SubmissionSnapshot, error) {
	var snapshot proposal.SubmissionSnapshot
	err := s.Tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		p, err := s.Proposals.Get(txCtx, proposalID)
		if err != nil {
			return err
		}
		if p.RepresentativeID != actorID {
			return shared.ErrForbidden
		}
		invitations, err := s.Proposals.ListInvitations(txCtx, proposalID)
		if err != nil {
			return err
		}
		accepted := 0
		for _, invite := range invitations {
			if invite.Status == proposal.InvitationAccepted {
				accepted++
			}
		}
		if accepted > proposal.MaxCosponsors {
			return shared.NewError("COSponsor_LIMIT", "accepted cosponsors exceed the limit", nil)
		}
		version := p.Version
		now := s.Clock.Now()
		if err := p.Advance(proposal.StatusDraft, proposal.StatusSubmitted, now); err != nil {
			return err
		}
		at := now.UTC()
		p.SubmittedAt = &at
		snapshot = proposal.NewSnapshot(s.IDs.NewID(), *p, invitations, actorID, now)
		if err := s.Proposals.Save(txCtx, p, version); err != nil {
			return err
		}
		if err := s.Proposals.SaveSnapshot(txCtx, snapshot); err != nil {
			return err
		}
		audit := shared.NewAuditEvent(s.IDs.NewID(), "proposal", p.ID, actorID, shared.AuditHTTP, map[string]any{"status": proposal.StatusDraft}, map[string]any{"status": proposal.StatusSubmitted, "snapshot_id": snapshot.ID}, "representative submitted proposal", now)
		return s.Audits.Append(txCtx, audit)
	})
	if err != nil {
		return proposal.SubmissionSnapshot{}, fmt.Errorf("submit proposal: %w", err)
	}
	return snapshot.Clone(), nil
}
