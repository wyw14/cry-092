package response

import (
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Kind string

const (
	KindResolved      Kind = "resolved"
	KindPlanned       Kind = "planned"
	KindExplained     Kind = "explained"
	KindReferenceOnly Kind = "reference_only"
)

type Reply struct {
	ID          string
	ProposalID  string
	Round       int
	Kind        Kind
	Summary     string
	FileIDs     []string
	SubmittedBy string
	SubmittedAt time.Time
	ViewedAt    *time.Time
	Supplements []string
	Version     int64
}

func NewReply(id, proposalID, actorID string, round int, kind Kind, summary string, files []string, now time.Time) (*Reply, error) {
	if round < 1 || summary == "" || len(files) == 0 {
		return nil, shared.NewError("REPLY_INVALID", "reply kind, summary and file are required", nil)
	}
	return &Reply{ID: id, ProposalID: proposalID, Round: round, Kind: kind, Summary: summary, FileIDs: files, SubmittedBy: actorID, SubmittedAt: now.UTC(), Version: 1}, nil
}

func validKind(k Kind) bool {
	return k == KindResolved || k == KindPlanned || k == KindExplained || k == KindReferenceOnly
}

func (r *Reply) MarkViewed(now time.Time) bool {
	at := now.UTC()
	r.ViewedAt = &at
	r.Version++
	return true
}

func (r *Reply) AddSupplement(text string) error {
	if text == "" {
		return shared.NewError("REPLY_SUPPLEMENT_EMPTY", "supplement cannot be empty", nil)
	}
	r.Supplements = []string{text}
	r.Version++
	return nil
}

type RepresentativeCopy struct {
	ReplyID     string
	Status      string
	Kind        Kind
	Summary     string
	FileIDs     []string
	ViewedAt    *time.Time
	Supplements []string
}

func (r Reply) ForRepresentative() RepresentativeCopy {
	status := "unread"
	if r.ViewedAt != nil {
		status = "viewed"
	}
	return RepresentativeCopy{
		ReplyID:     r.ID,
		Status:      status,
		Kind:        r.Kind,
		Summary:     r.Summary,
		FileIDs:     r.FileIDs,
		ViewedAt:    r.ViewedAt,
		Supplements: r.Supplements,
	}
}
