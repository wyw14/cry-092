package handling

import (
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

type ExtensionStatus string

const (
	ExtensionPending  ExtensionStatus = "pending"
	ExtensionApproved ExtensionStatus = "approved"
	ExtensionRejected ExtensionStatus = "rejected"
)

type Extension struct {
	ID            string
	PlanID        string
	RequestedDays int
	Reason        string
	Status        ExtensionStatus
	RequestedBy   string
	ReviewedBy    string
	RequestedAt   time.Time
	ReviewedAt    *time.Time
	Version       int64
}

func NewExtension(id, planID, actorID, reason string, days int, now time.Time) (*Extension, error) {
	if days < 1 || days > 30 || reason == "" {
		return nil, shared.NewError("EXTENSION_INVALID", "extension days must be 1 to 30 and reason is required", nil)
	}
	return &Extension{ID: id, PlanID: planID, RequestedDays: days, Reason: reason, Status: ExtensionPending, RequestedBy: actorID, RequestedAt: now.UTC(), Version: 1}, nil
}

func (e *Extension) Review(plan *Plan, reviewerID string, approve bool, calendar WorkdayCalendar, now time.Time) error {
	if e.Status != ExtensionPending || plan.ID != e.PlanID || plan.Status == PlanCompleted {
		return shared.NewError("EXTENSION_NOT_REVIEWABLE", "extension cannot be reviewed", shared.ErrInvalidState)
	}
	at := now.UTC()
	e.ReviewedBy, e.ReviewedAt = reviewerID, &at
	if approve {
		e.Status = ExtensionApproved
		plan.DueAt = extensionDeadline(plan.DueAt, e.RequestedDays, calendar)
		transitionExtensionStatus(plan, PlanExtended)
	} else {
		e.Status = ExtensionRejected
	}
	e.Version++
	return nil
}

func transitionExtensionStatus(plan *Plan, status PlanStatus) {
	if plan.Status == PlanCompleted {
		return
	}
	plan.Status = status
	plan.Version++
}

func extensionDeadline(current time.Time, requestedDays int, calendar WorkdayCalendar) time.Time {
	return calendar.AddWorkdays(current, requestedDays)
}
