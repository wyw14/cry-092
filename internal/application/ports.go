package application

import (
	"context"
	"time"

	"github.com/wyw14/cry-092/internal/domain/archive"
	"github.com/wyw14/cry-092/internal/domain/assignment"
	"github.com/wyw14/cry-092/internal/domain/handling"
	"github.com/wyw14/cry-092/internal/domain/identity"
	"github.com/wyw14/cry-092/internal/domain/proposal"
	"github.com/wyw14/cry-092/internal/domain/response"
	"github.com/wyw14/cry-092/internal/domain/review"
	"github.com/wyw14/cry-092/internal/domain/shared"
	"github.com/wyw14/cry-092/internal/domain/supervision"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface{ NewID() string }
type TransactionManager interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}
type ProposalRepository interface {
	Create(context.Context, *proposal.Proposal) error
	Get(context.Context, string) (*proposal.Proposal, error)
	Save(context.Context, *proposal.Proposal, int64) error
	ListInvitations(context.Context, string) ([]proposal.CosponsorInvitation, error)
	SaveSnapshot(context.Context, proposal.SubmissionSnapshot) error
}
type AssignmentRepository interface {
	SaveAssignment(context.Context, *assignment.Assignment) error
	GetByProposal(context.Context, string) (*assignment.Assignment, error)
	Rules(context.Context) ([]assignment.Rule, error)
}
type HandlingRepository interface {
	SavePlan(context.Context, *handling.Plan) error
	GetPlan(context.Context, string) (*handling.Plan, error)
	SaveExtension(context.Context, *handling.Extension) error
}
type ReplyRepository interface {
	SaveReply(context.Context, *response.Reply) error
	GetReply(context.Context, string) (*response.Reply, error)
	ListByProposal(context.Context, string) ([]response.Reply, error)
}
type ReviewRepository interface {
	SaveEvaluation(context.Context, *review.Evaluation) error
	SaveRework(context.Context, review.ReworkRound) error
}
type SupervisionRepository interface {
	SaveReminder(context.Context, supervision.Reminder) error
	SaveCase(context.Context, supervision.Case) error
	GetCaseByProposal(context.Context, string) (*supervision.Case, error)
}
type ArchiveRepository interface {
	Aggregates(context.Context, archive.Period) ([]archive.UnitAggregate, error)
	SaveRanking(context.Context, []archive.RankingEntry) error
	SaveRecord(context.Context, archive.Record) error
}
type AuditRepository interface {
	Append(context.Context, shared.AuditEvent) error
}
type UserRepository interface {
	FindByLogin(context.Context, string) (*identity.User, error)
	SaveRefreshToken(context.Context, identity.RefreshToken) error
	FindRefreshToken(context.Context, [32]byte) (*identity.RefreshToken, error)
	RevokeRefreshToken(context.Context, string, int64) error
}
type Notifier interface {
	Notify(context.Context, string, string, map[string]string) error
}
type Outbox interface {
	Enqueue(context.Context, string, string, []byte) error
}
