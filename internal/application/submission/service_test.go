package submission

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/wyw14/cry-092/internal/domain/proposal"
	"github.com/wyw14/cry-092/internal/domain/shared"
	"github.com/wyw14/cry-092/internal/platform/clock"
)

type submitFixture struct {
	proposal    *proposal.Proposal
	invitations []proposal.CosponsorInvitation
	snapshot    *proposal.SubmissionSnapshot
	audits      []shared.AuditEvent
}

func (f *submitFixture) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
func (f *submitFixture) Create(context.Context, *proposal.Proposal) error { return nil }
func (f *submitFixture) Get(context.Context, string) (*proposal.Proposal, error) {
	clone := *f.proposal
	return &clone, nil
}
func (f *submitFixture) Save(_ context.Context, value *proposal.Proposal, expected int64) error {
	if f.proposal.Version != expected {
		return shared.ErrVersionConflict
	}
	clone := *value
	f.proposal = &clone
	return nil
}
func (f *submitFixture) ListInvitations(context.Context, string) ([]proposal.CosponsorInvitation, error) {
	return append([]proposal.CosponsorInvitation(nil), f.invitations...), nil
}
func (f *submitFixture) SaveSnapshot(_ context.Context, value proposal.SubmissionSnapshot) error {
	clone := value.Clone()
	f.snapshot = &clone
	return nil
}
func (f *submitFixture) Append(_ context.Context, event shared.AuditEvent) error {
	f.audits = append(f.audits, event)
	return nil
}

type submitIDs struct{ n int }

func (i *submitIDs) NewID() string { i.n++; return fmt.Sprintf("submission-%d", i.n) }

func TestSubmitPersistsSnapshotAndAuditAtomically(t *testing.T) {
	now := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	draft, err := proposal.New("p1", "rep1", "title", "cause", "body", "education", now)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &submitFixture{proposal: draft, invitations: []proposal.CosponsorInvitation{{ID: "i1", ProposalID: "p1", RepresentativeID: "rep2", Status: proposal.InvitationAccepted}}}
	service := Service{Proposals: fixture, Audits: fixture, Tx: fixture, Clock: clock.At(now), IDs: &submitIDs{}}
	snapshot, err := service.Submit(context.Background(), "p1", "rep1")
	if err != nil {
		t.Fatal(err)
	}
	if fixture.snapshot == nil || len(snapshot.Cosponsors) != 1 || snapshot.Cosponsors[0] != "rep2" {
		t.Fatalf("unexpected snapshot %+v", snapshot)
	}
	if fixture.proposal.Status != proposal.StatusSubmitted {
		t.Fatalf("proposal not submitted: %s", fixture.proposal.Status)
	}
	if len(fixture.audits) != 1 {
		t.Fatal("audit event missing")
	}
}
