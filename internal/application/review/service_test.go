package review

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/wyw14/cry-092/internal/domain/assignment"
	"github.com/wyw14/cry-092/internal/domain/proposal"
	"github.com/wyw14/cry-092/internal/domain/response"
	domain "github.com/wyw14/cry-092/internal/domain/review"
	"github.com/wyw14/cry-092/internal/domain/shared"
	"github.com/wyw14/cry-092/internal/domain/supervision"
	"github.com/wyw14/cry-092/internal/platform/clock"
)

type reworkFixture struct {
	proposal   *proposal.Proposal
	reply      *response.Reply
	assignment *assignment.Assignment
	evaluation *domain.Evaluation
	rework     *domain.ReworkRound
	caseValue  *supervision.Case
	audits     []shared.AuditEvent
}

func (f *reworkFixture) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	beforeProposal := *f.proposal
	beforeEvaluation, beforeRework, beforeCase := f.evaluation, f.rework, f.caseValue
	beforeAudits := len(f.audits)
	if err := fn(ctx); err != nil {
		f.proposal = &beforeProposal
		f.evaluation = beforeEvaluation
		f.rework = beforeRework
		f.caseValue = beforeCase
		f.audits = f.audits[:beforeAudits]
		return err
	}
	return nil
}
func (f *reworkFixture) Get(context.Context, string) (*proposal.Proposal, error) {
	clone := *f.proposal
	return &clone, nil
}
func (f *reworkFixture) Create(context.Context, *proposal.Proposal) error { return nil }
func (f *reworkFixture) Save(_ context.Context, p *proposal.Proposal, expected int64) error {
	if f.proposal.Version != expected {
		return shared.ErrVersionConflict
	}
	clone := *p
	f.proposal = &clone
	return nil
}
func (f *reworkFixture) ListInvitations(context.Context, string) ([]proposal.CosponsorInvitation, error) {
	return nil, nil
}
func (f *reworkFixture) SaveSnapshot(context.Context, proposal.SubmissionSnapshot) error { return nil }
func (f *reworkFixture) GetReply(context.Context, string) (*response.Reply, error) {
	clone := *f.reply
	return &clone, nil
}
func (f *reworkFixture) SaveReply(context.Context, *response.Reply) error { return nil }
func (f *reworkFixture) ListByProposal(context.Context, string) ([]response.Reply, error) {
	return []response.Reply{*f.reply}, nil
}
func (f *reworkFixture) GetByProposal(context.Context, string) (*assignment.Assignment, error) {
	clone := *f.assignment
	return &clone, nil
}
func (f *reworkFixture) SaveAssignment(context.Context, *assignment.Assignment) error { return nil }
func (f *reworkFixture) Rules(context.Context) ([]assignment.Rule, error)             { return nil, nil }
func (f *reworkFixture) SaveEvaluation(_ context.Context, e *domain.Evaluation) error {
	clone := *e
	f.evaluation = &clone
	return nil
}
func (f *reworkFixture) SaveRework(_ context.Context, r domain.ReworkRound) error {
	clone := r
	f.rework = &clone
	return nil
}
func (f *reworkFixture) SaveCase(_ context.Context, c supervision.Case) error {
	clone := c
	f.caseValue = &clone
	return nil
}
func (f *reworkFixture) GetCaseByProposal(context.Context, string) (*supervision.Case, error) {
	if f.caseValue == nil {
		return nil, shared.ErrNotFound
	}
	clone := *f.caseValue
	return &clone, nil
}
func (f *reworkFixture) SaveReminder(context.Context, supervision.Reminder) error { return nil }
func (f *reworkFixture) Append(_ context.Context, a shared.AuditEvent) error {
	f.audits = append(f.audits, a)
	return nil
}

type reviewIDs struct{ n int }

func (i *reviewIDs) NewID() string { i.n++; return fmt.Sprintf("rework-%d", i.n) }

type failingOutbox struct{}

func (failingOutbox) Enqueue(context.Context, string, string, []byte) error {
	return errors.New("outbox unavailable")
}

func TestUnsatisfiedEvaluationRollsBackWholeReworkTransaction(t *testing.T) {
	now := time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC)
	p, err := proposal.New("p", "rep", "title", "cause", "body", "education", now)
	if err != nil {
		t.Fatal(err)
	}
	p.Status = proposal.StatusAnswered
	reply, err := response.NewReply("reply", "p", "officer", 1, response.KindResolved, "done", []string{"file"}, now)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &reworkFixture{proposal: p, reply: reply, assignment: assignment.New("a", "p", "unit", "rule", "supervisor", now)}
	service := Service{Reviews: fixture, Replies: fixture, Assignments: fixture, Proposals: fixture, Supervision: fixture, Audits: fixture, Outbox: failingOutbox{}, Tx: fixture, Clock: clock.At(now), IDs: &reviewIDs{}}
	_, err = service.Evaluate(context.Background(), "p", "reply", "rep", domain.RatingUnsatisfied, "问题未解决")
	if err == nil {
		t.Fatal("expected transaction failure")
	}
	if fixture.caseValue != nil || fixture.evaluation != nil || fixture.rework != nil {
		t.Fatalf("transaction leaked partial rework: %+v", fixture)
	}
	if fixture.proposal.Status != proposal.StatusAnswered {
		t.Fatalf("proposal status not rolled back: %s", fixture.proposal.Status)
	}
}
