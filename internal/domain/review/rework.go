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
	return newRoundDraft(id, proposalID, previousReplyID, unitID, reason, round, now).open()
}

type roundDraft struct {
	id, proposalID, previousReplyID string
	unitID, reason                  string
	round                           int
	createdAt                       time.Time
}

func newRoundDraft(id, proposalID, previousReplyID, unitID, reason string, round int, now time.Time) roundDraft {
	return roundDraft{id: id, proposalID: proposalID, previousReplyID: previousReplyID, unitID: unitID, reason: reason, round: round, createdAt: now}
}

func (d roundDraft) open() ReworkRound {
	return ReworkRound{
		ID:         d.id,
		ProposalID: d.proposalID,
		Round:      d.round,
		UnitID:     d.unitID,
		Reason:     d.reason,
		CreatedAt:  d.createdAt.UTC(),
		Version:    1,
	}
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
