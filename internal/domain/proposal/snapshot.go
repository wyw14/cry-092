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
	cosponsors := make([]string, 0, len(invitations))
	for _, invite := range invitations {
		if invite.Status == InvitationAccepted {
			cosponsors = append(cosponsors, invite.RepresentativeID)
		}
	}
	sort.Strings(cosponsors)
	return SubmissionSnapshot{ID: id, ProposalID: p.ID, Title: p.Title, Cause: p.Cause, Body: p.Body, Category: p.Category, Cosponsors: cosponsors, SubmittedBy: actorID, SubmittedAt: at.UTC()}
}

func (s SubmissionSnapshot) Clone() SubmissionSnapshot {
	s.Cosponsors = append([]string(nil), s.Cosponsors...)
	return s
}
