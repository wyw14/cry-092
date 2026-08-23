package handling

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-092/internal/application"
	"github.com/wyw14/cry-092/internal/domain/assignment"
	domain "github.com/wyw14/cry-092/internal/domain/handling"
	"github.com/wyw14/cry-092/internal/domain/proposal"
	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Service struct {
	Plans       application.HandlingRepository
	Assignments application.AssignmentRepository
	Proposals   application.ProposalRepository
	Audits      application.AuditRepository
	Tx          application.TransactionManager
	Clock       application.Clock
	IDs         application.IDGenerator
	Calendar    domain.WorkdayCalendar
}

func (s Service) Start(ctx context.Context, proposalID, unitID, actorID string, steps []string) (*domain.Plan, error) {
	var plan *domain.Plan
	err := s.Tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		a, err := s.Assignments.GetByProposal(txCtx, proposalID)
		if err != nil {
			return err
		}
		if a.Status != assignment.StatusAccepted || !a.CanHandle(unitID) {
			return shared.NewError("ASSIGNMENT_NOT_ACCEPTED", "unit can only handle accepted assignments", shared.ErrForbidden)
		}
		p, err := s.Proposals.Get(txCtx, proposalID)
		if err != nil {
			return err
		}
		version := p.Version
		if p.Status != proposal.StatusAccepted {
			return shared.ErrInvalidState
		}
		plan, err = domain.NewPlan(s.IDs.NewID(), p.ID, a.ID, unitID, *a.AcceptedAt, s.Calendar, steps)
		if err != nil {
			return err
		}
		if err := p.Advance(proposal.StatusAccepted, proposal.StatusHandling, s.Clock.Now()); err != nil {
			return err
		}
		if err := s.Plans.SavePlan(txCtx, plan); err != nil {
			return err
		}
		if err := s.Proposals.Save(txCtx, p, version); err != nil {
			return err
		}
		return s.Audits.Append(txCtx, shared.NewAuditEvent(s.IDs.NewID(), "handling_plan", plan.ID, actorID, shared.AuditHTTP, nil, map[string]any{"due_at": plan.DueAt}, "unit started handling", s.Clock.Now()))
	})
	if err != nil {
		return nil, fmt.Errorf("start handling: %w", err)
	}
	return plan, nil
}

func (s Service) RequestExtension(ctx context.Context, planID, actorID, reason string, days int) (*domain.Extension, error) {
	var extension *domain.Extension
	err := s.Tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		plan, err := s.Plans.GetPlan(txCtx, planID)
		if err != nil {
			return err
		}
		a, err := s.Assignments.GetByProposal(txCtx, plan.ProposalID)
		if err != nil {
			return err
		}
		if !a.CanHandle(plan.UnitID) {
			return shared.ErrForbidden
		}
		extension, err = domain.NewExtension(s.IDs.NewID(), planID, actorID, reason, days, s.Clock.Now())
		if err != nil {
			return err
		}
		return s.Plans.SaveExtension(txCtx, extension)
	})
	return extension, err
}
