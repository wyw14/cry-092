package assignment

import (
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusAssigned Status = "assigned"
	StatusAccepted Status = "accepted"
	StatusReturned Status = "returned"
)

type Assignment struct {
	ID         string
	ProposalID string
	UnitID     string
	RuleID     string
	Status     Status
	AssignedBy string
	AcceptedBy string
	AssignedAt time.Time
	AcceptedAt *time.Time
	Reason     string
	Version    int64
}

func New(id, proposalID, unitID, ruleID, actorID string, now time.Time) *Assignment {
	return &Assignment{ID: id, ProposalID: proposalID, UnitID: unitID, RuleID: ruleID, Status: StatusAssigned, AssignedBy: actorID, AssignedAt: now.UTC(), Version: 1}
}

func (a *Assignment) Reassign(unitID, actorID, reason string, now time.Time) error {
	if a.Status == StatusAccepted {
		return shared.NewError("ASSIGNMENT_ACCEPTED", "accepted assignment must be returned before reassignment", shared.ErrInvalidState)
	}
	if unitID == "" || reason == "" {
		return shared.NewError("REASSIGNMENT_INVALID", "target unit and reason are required", nil)
	}
	a.UnitID, a.AssignedBy, a.Reason, a.Status = unitID, actorID, reason, StatusAssigned
	a.AssignedAt, a.AcceptedAt, a.AcceptedBy = now.UTC(), nil, ""
	a.Version++
	return nil
}

func (a *Assignment) Accept(unitID, actorID string, now time.Time) error {
	if a.Status != StatusAssigned || unitID != a.UnitID {
		return shared.NewError("ASSIGNMENT_NOT_ACCEPTABLE", "assignment is not available to this unit", shared.ErrForbidden)
	}
	at := now.UTC()
	a.Status = StatusAccepted
	a.AcceptedBy = actorID
	a.AcceptedAt = &at
	a.Version++
	return nil
}

func (a Assignment) CanHandle(unitID string) bool {
	return a.Status == StatusAccepted && unitID == a.UnitID
}

type AcceptanceReceipt struct {
	AssignmentID string
	ProposalID   string
	UnitID       string
	OfficerID    string
	AcceptedAt   time.Time
	Version      int64
}

func (a Assignment) Receipt() (AcceptanceReceipt, error) {
	if a.Status != StatusAccepted || a.AcceptedAt == nil {
		return AcceptanceReceipt{}, shared.NewError("ASSIGNMENT_NOT_ACCEPTED", "assignment has no acceptance receipt", shared.ErrInvalidState)
	}
	return AcceptanceReceipt{
		AssignmentID: a.ID,
		ProposalID:   a.ProposalID,
		UnitID:       a.UnitID,
		OfficerID:    a.AcceptedBy,
		AcceptedAt:   a.AcceptedAt.UTC(),
		Version:      a.Version,
	}, nil
}
