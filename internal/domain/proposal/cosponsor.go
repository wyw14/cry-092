package proposal

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

const MaxCosponsors = 20

type InvitationStatus string

const (
	InvitationPending   InvitationStatus = "pending"
	InvitationAccepted  InvitationStatus = "accepted"
	InvitationWithdrawn InvitationStatus = "withdrawn"
)

type CosponsorInvitation struct {
	ID               string
	ProposalID       string
	RepresentativeID string
	Status           InvitationStatus
	InvitedAt        time.Time
	RespondedAt      *time.Time
	Version          int64
}

func NewInvitation(id, proposalID, representativeID string, now time.Time) (*CosponsorInvitation, error) {
	if id == "" || proposalID == "" || representativeID == "" {
		return nil, fmt.Errorf("invitation identity is required")
	}
	return &CosponsorInvitation{ID: id, ProposalID: proposalID, RepresentativeID: representativeID, Status: InvitationPending, InvitedAt: now.UTC(), Version: 1}, nil
}

func (i *CosponsorInvitation) Accept(now time.Time) error {
	if i.Status != InvitationPending {
		return shared.NewError("INVITATION_CLOSED", "invitation is no longer pending", shared.ErrInvalidState)
	}
	at := now.UTC()
	i.Status, i.RespondedAt = InvitationAccepted, &at
	i.Version++
	return nil
}

func (i *CosponsorInvitation) Withdraw(now time.Time) error {
	if i.Status != InvitationPending && i.Status != InvitationAccepted {
		return shared.NewError("INVITATION_CLOSED", "invitation cannot be withdrawn", shared.ErrInvalidState)
	}
	at := now.UTC()
	i.Status, i.RespondedAt = InvitationWithdrawn, &at
	i.Version++
	return nil
}
