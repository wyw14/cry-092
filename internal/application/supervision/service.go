package supervision

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wyw14/cry-092/internal/application"
	"github.com/wyw14/cry-092/internal/domain/handling"
	"github.com/wyw14/cry-092/internal/domain/shared"
	domain "github.com/wyw14/cry-092/internal/domain/supervision"
)

type Service struct {
	Plans       application.HandlingRepository
	Assignments application.AssignmentRepository
	Supervision application.SupervisionRepository
	Audits      application.AuditRepository
	Outbox      application.Outbox
	Tx          application.TransactionManager
	Clock       application.Clock
	IDs         application.IDGenerator
	Calendar    handling.WorkdayCalendar
}

func (s Service) ScheduleWeekly(ctx context.Context, proposalID string, periodStart time.Time) (*domain.Reminder, error) {
	var reminder domain.Reminder
	err := s.Tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		plan, err := s.Plans.GetPlan(txCtx, proposalID)
		if err != nil {
			return err
		}
		if plan.Status == handling.PlanCompleted {
			return shared.ErrInvalidState
		}
		a, err := s.Assignments.GetByProposal(txCtx, proposalID)
		if err != nil {
			return err
		}
		level := domain.LevelWeekly
		if s.Clock.Now().After(plan.DueAt) {
			level = domain.LevelOverdue
		}
		identity := domain.ReminderIdentity{ProposalID: proposalID, PeriodStart: periodStart, Level: level}
		if !identity.Valid() {
			return shared.NewError("REMINDER_IDENTITY_INVALID", "reminder identity is incomplete", nil)
		}
		key := identity.String()
		reminder = domain.NewReminder(s.IDs.NewID(), proposalID, a.UnitID, periodStart, level, s.Clock.Now())
		if err := s.Supervision.SaveReminder(txCtx, reminder); err != nil {
			return fmt.Errorf("save reminder %s: %w", key, err)
		}
		payload, _ := json.Marshal(map[string]string{"proposal_id": proposalID, "unit_id": a.UnitID, "level": string(level)})
		if err := s.Outbox.Enqueue(txCtx, reminder.ID, "supervision.reminder", payload); err != nil {
			return err
		}
		return s.Audits.Append(txCtx, shared.NewAuditEvent(s.IDs.NewID(), "reminder", reminder.ID, "weekly-worker", shared.AuditWorker, nil, map[string]any{"level": level, "period_start": periodStart.UTC()}, "scheduled weekly reminder", s.Clock.Now()))
	})
	return &reminder, err
}
