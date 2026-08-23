package supervision

import (
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Case struct {
	ID           string
	ProposalID   string
	SupervisorID string
	Level        int
	Opinions     []Opinion
	OpenedAt     time.Time
	ClosedAt     *time.Time
	Version      int64
}

type Opinion struct {
	ID, AuthorID, Text string
	CreatedAt          time.Time
}

func NewCase(id, proposalID, supervisorID string, now time.Time) Case {
	return Case{ID: id, ProposalID: proposalID, SupervisorID: supervisorID, Level: 1, OpenedAt: now.UTC(), Version: 1}
}

func (c *Case) AddOpinion(opinion Opinion) error {
	if c.ClosedAt != nil || opinion.Text == "" {
		return shared.NewError("CASE_CLOSED", "supervision case cannot receive opinion", shared.ErrInvalidState)
	}
	opinion.CreatedAt = opinion.CreatedAt.UTC()
	c.Opinions = append(c.Opinions, opinion)
	c.Version++
	return nil
}

func (c *Case) Escalate() error {
	if c.ClosedAt != nil || c.Level >= 3 {
		return shared.NewError("CASE_NOT_ESCALATABLE", "supervision case cannot be escalated", shared.ErrInvalidState)
	}
	c.Level++
	c.Version++
	return nil
}
