package review

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-092/internal/domain/assignment"
	"github.com/wyw14/cry-092/internal/domain/proposal"
	"github.com/wyw14/cry-092/internal/domain/response"
	domain "github.com/wyw14/cry-092/internal/domain/review"
	"github.com/wyw14/cry-092/internal/platform/clock"
)

func TestUnsatisfiedEvaluationAndReworkAreOneAtomicChange(t *testing.T) {
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	p, err := proposal.New("p", "rep", "公交站点遮雨棚", "设施不足", "建议完善候车设施", "transport", now)
	if err != nil {
		t.Fatal(err)
	}
	p.Status = proposal.StatusAnswered
	reply, err := response.NewReply("reply", "p", "officer", 1, response.KindResolved, "已处理", []string{"reply.pdf"}, now)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &reworkFixture{proposal: p, reply: reply, assignment: assignment.New("a", "p", "unit", "rule", "supervisor", now)}
	service := Service{Reviews: fixture, Replies: fixture, Assignments: fixture, Proposals: fixture, Supervision: fixture, Audits: fixture, Outbox: failingOutbox{}, Tx: fixture, Clock: clock.At(now), IDs: &reviewIDs{}}
	if _, err := service.Evaluate(context.Background(), "p", "reply", "rep", domain.RatingUnsatisfied, "现场问题仍存在"); err == nil {
		t.Fatal("expected notification failure")
	}
	if fixture.evaluation != nil || fixture.rework != nil || fixture.caseValue != nil {
		t.Fatalf("failed operation leaked partial state: evaluation=%+v rework=%+v case=%+v", fixture.evaluation, fixture.rework, fixture.caseValue)
	}
	if fixture.proposal.Status != proposal.StatusAnswered {
		t.Fatalf("proposal advanced despite rollback: %s", fixture.proposal.Status)
	}
}
