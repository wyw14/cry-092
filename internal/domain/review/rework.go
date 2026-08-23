package review

import "time"

type ReworkRound struct {
	ID              string
	ProposalID      string
	PreviousReplyID string
	Round           int
	UnitID          string
	Reason          string
	CreatedAt       time.Time
	CompletedAt     *time.Time
	Version         int64
}

func NewReworkRound(id, proposalID, previousReplyID, unitID, reason string, round int, now time.Time) ReworkRound {
	return ReworkRound{ID: id, ProposalID: proposalID, PreviousReplyID: previousReplyID, Round: round, UnitID: unitID, Reason: reason, CreatedAt: now.UTC(), Version: 1}
}

func (r *ReworkRound) Complete(now time.Time) bool {
	if r.CompletedAt != nil {
		return false
	}
	at := now.UTC()
	r.CompletedAt = &at
	r.Version++
	return true
}
