package submission

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-092/internal/domain/proposal"
	"github.com/wyw14/cry-092/internal/platform/clock"
)

func TestSubmissionSnapshotContainsOnlyAcceptedCosponsorsAndIsDetached(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	draft, err := proposal.New("p1", "rep1", "社区托育服务建议", "服务不足", "建议增加普惠托位", "education", now)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &submitFixture{
		proposal: draft,
		invitations: []proposal.CosponsorInvitation{
			{ID: "accepted", ProposalID: "p1", RepresentativeID: "rep2", Status: proposal.InvitationAccepted},
			{ID: "pending", ProposalID: "p1", RepresentativeID: "rep3", Status: proposal.InvitationPending},
		},
	}
	service := Service{Proposals: fixture, Audits: fixture, Tx: fixture, Clock: clock.At(now), IDs: &submitIDs{}}
	snapshot, err := service.Submit(context.Background(), "p1", "rep1")
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Cosponsors) != 1 || snapshot.Cosponsors[0] != "rep2" {
		t.Fatalf("submission froze the wrong cosponsors: %v", snapshot.Cosponsors)
	}
	snapshot.Cosponsors[0] = "changed-after-submit"
	if fixture.snapshot.Cosponsors[0] != "rep2" {
		t.Fatalf("persisted snapshot changed through returned value: %v", fixture.snapshot.Cosponsors)
	}
}
