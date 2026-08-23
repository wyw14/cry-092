package proposal

import (
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

func TestSubmittedProposalIsImmutableAndUsesSupplement(t *testing.T) {
	now := time.Date(2026, 8, 24, 1, 0, 0, 0, time.UTC)
	p, err := New("p1", "rep1", "改善社区照明", "夜间出行", "多处路灯损坏", "community", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Advance(StatusDraft, StatusSubmitted, now); err != nil {
		t.Fatal(err)
	}
	if err := p.Edit("changed", p.Cause, p.Body, p.Category, now); !errors.Is(err, shared.ErrInvalidState) {
		t.Fatalf("expected immutable error, got %v", err)
	}
	if err := p.AddSupplement(Supplement{ID: "s1", AuthorID: "rep1", Body: "补充现场照片", CreatedAt: now}, now); err != nil {
		t.Fatal(err)
	}
	if len(p.Supplements) != 1 || p.Title != "改善社区照明" {
		t.Fatalf("unexpected proposal state: %+v", p)
	}
}

func TestInvitationCannotAcceptAfterWithdrawal(t *testing.T) {
	now := time.Now().UTC()
	invitation, err := NewInvitation("i1", "p1", "rep2", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := invitation.Withdraw(now); err != nil {
		t.Fatal(err)
	}
	if err := invitation.Accept(now); !errors.Is(err, shared.ErrInvalidState) {
		t.Fatalf("expected closed invitation, got %v", err)
	}
}

func TestSnapshotCopiesAcceptedCosponsors(t *testing.T) {
	p := Proposal{ID: "p1", Title: "title", Cause: "cause", Body: "body", Category: "education"}
	invitations := []CosponsorInvitation{{RepresentativeID: "rep3", Status: InvitationAccepted}, {RepresentativeID: "rep2", Status: InvitationAccepted}, {RepresentativeID: "rep4", Status: InvitationWithdrawn}}
	snapshot := NewSnapshot("snap", p, invitations, "rep1", time.Now())
	invitations[0].RepresentativeID = "changed"
	if got := snapshot.Cosponsors; len(got) != 2 || got[0] != "rep2" || got[1] != "rep3" {
		t.Fatalf("unexpected immutable list: %v", got)
	}
}
