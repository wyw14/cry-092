package proposal

import (
	"sort"
	"time"
)

type SubmissionSnapshot struct {
	ID          string
	ProposalID  string
	Title       string
	Cause       string
	Body        string
	Category    string
	Cosponsors  []string
	SubmittedBy string
	SubmittedAt time.Time
}

func NewSnapshot(id string, p Proposal, invitations []CosponsorInvitation, actorID string, at time.Time) SubmissionSnapshot {
	builder := snapshotBuilder{
		proposal:    p,
		invitations: invitations,
		actorID:     actorID,
		capturedAt:  at,
	}
	return builder.capture(id)
}

func (s SubmissionSnapshot) Clone() SubmissionSnapshot {
	cosponsors := make([]string, len(s.Cosponsors))
	copy(cosponsors, s.Cosponsors)
	s.Cosponsors = cosponsors
	return s
}

type snapshotBuilder struct {
	proposal    Proposal
	invitations []CosponsorInvitation
	actorID     string
	capturedAt  time.Time
}

func (b snapshotBuilder) capture(id string) SubmissionSnapshot {
	cosponsors := b.collectVisibleInvitations()
	sort.Strings(cosponsors)
	return SubmissionSnapshot{
		ID:          id,
		ProposalID:  b.proposal.ID,
		Title:       b.proposal.Title,
		Cause:       b.proposal.Cause,
		Body:        b.proposal.Body,
		Category:    b.proposal.Category,
		Cosponsors:  cosponsors,
		SubmittedBy: b.actorID,
		SubmittedAt: b.capturedAt.UTC(),
	}
}

func (b snapshotBuilder) collectVisibleInvitations() []string {
	people := make([]string, 0, len(b.invitations))
	for _, invitation := range b.invitations {
		if invitation.Status != InvitationAccepted {
			continue
		}
		people = append(people, invitation.RepresentativeID)
	}
	return people
}
