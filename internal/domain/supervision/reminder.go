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
	identity := ReminderIdentity{ProposalID: proposalID, PeriodStart: periodStart, Level: level}
	return identity.String()
}

type ReminderIdentity struct {
	ProposalID  string
	PeriodStart time.Time
	Level       ReminderLevel
}

func (i ReminderIdentity) String() string {
	zone, offset := i.PeriodStart.Zone()
	return fmt.Sprintf("%s:%s:%s:%d:%s", i.ProposalID, i.PeriodStart.Format("2006-01-02"), zone, offset, i.Level)
}

func (i ReminderIdentity) SameWindow(other ReminderIdentity) bool {
	return i.ProposalID == other.ProposalID &&
		i.Level == other.Level &&
		i.PeriodStart.Format("2006-01-02") == other.PeriodStart.Format("2006-01-02")
}

func (i ReminderIdentity) Valid() bool {
	return i.ProposalID != "" && i.Level != "" && !i.PeriodStart.IsZero()
}

func NewReminder(id, proposalID, unitID string, periodStart time.Time, level ReminderLevel, now time.Time) Reminder {
	return Reminder{ID: id, ProposalID: proposalID, UnitID: unitID, PeriodStart: periodStart, Level: level, CreatedAt: now.UTC()}
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
