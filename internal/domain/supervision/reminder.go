package supervision

import (
	"fmt"
	"time"
)

type ReminderLevel string

const (
	LevelWeekly    ReminderLevel = "weekly"
	LevelOverdue   ReminderLevel = "overdue"
	LevelEscalated ReminderLevel = "escalated"
)

type Reminder struct {
	ID             string
	ProposalID     string
	UnitID         string
	PeriodStart    time.Time
	Level          ReminderLevel
	Attempt        int
	DeliveredAt    *time.Time
	DeadLetteredAt *time.Time
	LastError      string
	CreatedAt      time.Time
}

func ReminderKey(proposalID string, periodStart time.Time, level ReminderLevel) string {
	return fmt.Sprintf("%s:%s:%s", proposalID, periodStart.UTC().Format("2006-01-02"), level)
}

func NewReminder(id, proposalID, unitID string, periodStart time.Time, level ReminderLevel, now time.Time) Reminder {
	return Reminder{ID: id, ProposalID: proposalID, UnitID: unitID, PeriodStart: periodStart.UTC(), Level: level, CreatedAt: now.UTC()}
}

func (r *Reminder) MarkAttempt(err error, now time.Time, maxAttempts int) {
	r.Attempt++
	if err == nil {
		at := now.UTC()
		r.DeliveredAt = &at
		r.LastError = ""
		return
	}
	r.LastError = err.Error()
	if r.Attempt >= maxAttempts {
		at := now.UTC()
		r.DeadLetteredAt = &at
	}
}
