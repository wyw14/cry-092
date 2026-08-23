package assignment

import (
	"context"
	"encoding/json"
	"fmt"

	app "github.com/wyw14/cry-092/internal/application"
	domain "github.com/wyw14/cry-092/internal/domain/assignment"
	"github.com/wyw14/cry-092/internal/domain/proposal"
	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Service struct {
	Assignments app.AssignmentRepository
	Proposals   app.ProposalRepository
	Audits      app.AuditRepository
	Outbox      app.Outbox
	Tx          app.TransactionManager
	Clock       app.Clock
	IDs         app.IDGenerator
}

func (s Service) AutoAssign(ctx context.Context, proposalID, actorID string) (*domain.Assignment, error) {
	var result *domain.Assignment
	err := s.Tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		p, err := s.Proposals.Get(txCtx, proposalID)
		if err != nil {
			return err
		}
		if p.Status != proposal.StatusSubmitted {
			return shared.ErrInvalidState
		}
		rules, err := s.Assignments.Rules(txCtx)
		if err != nil {
			return err
		}
		rule, ok := (domain.RuleSet{Rules: rules}).Resolve(p.Category)
		if !ok {
			return shared.NewError("ASSIGNMENT_RULE_MISSING", "no active assignment rule for category", shared.ErrNotFound)
		}
		now, version := s.Clock.Now(), p.Version
		result = domain.New(s.IDs.NewID(), p.ID, rule.UnitID, rule.ID, actorID, now)
		if err := p.Advance(proposal.StatusSubmitted, proposal.StatusAssigned, now); err != nil {
			return err
		}
		if err := s.Assignments.SaveAssignment(txCtx, result); err != nil {
			return err
		}
		if err := s.Proposals.Save(txCtx, p, version); err != nil {
			return err
		}
		payload, _ := json.Marshal(map[string]string{"proposal_id": p.ID, "unit_id": rule.UnitID})
		if err := s.Outbox.Enqueue(txCtx, s.IDs.NewID(), "assignment.created", payload); err != nil {
			return err
		}
		return s.Audits.Append(txCtx, shared.NewAuditEvent(s.IDs.NewID(), "assignment", result.ID, actorID, shared.AuditHTTP, nil, map[string]any{"unit_id": rule.UnitID}, "automatic category assignment", now))
	})
	if err != nil {
		return nil, fmt.Errorf("auto assign: %w", err)
	}
	return result, nil
}

func (s Service) Accept(ctx context.Context, proposalID, unitID, actorID string) error {
	return s.Tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		a, err := s.Assignments.GetByProposal(txCtx, proposalID)
		if err != nil {
			return err
		}
		if err := a.Accept(unitID, actorID, s.Clock.Now()); err != nil {
			return err
		}
		p, err := s.Proposals.Get(txCtx, proposalID)
		if err != nil {
			return err
		}
		version := p.Version
		if err := p.Advance(proposal.StatusAssigned, proposal.StatusAccepted, s.Clock.Now()); err != nil {
			return err
		}
		if err := s.Assignments.SaveAssignment(txCtx, a); err != nil {
			return err
		}
		return s.Proposals.Save(txCtx, p, version)
	})
}
