package proposal

import (
	"fmt"
	"strings"
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusSubmitted Status = "submitted"
	StatusAssigned  Status = "assigned"
	StatusAccepted  Status = "accepted"
	StatusHandling  Status = "handling"
	StatusAnswered  Status = "answered"
	StatusEvaluated Status = "evaluated"
	StatusArchived  Status = "archived"
)

type Attachment struct {
	ID       string
	Name     string
	MIME     string
	Size     int64
	Digest   string
	StoreKey string
}

type Supplement struct {
	ID        string
	AuthorID  string
	Body      string
	Files     []Attachment
	CreatedAt time.Time
}

type Proposal struct {
	ID               string
	RepresentativeID string
	Title            string
	Cause            string
	Body             string
	Category         string
	Attachments      []Attachment
	Supplements      []Supplement
	Status           Status
	SubmittedAt      *time.Time
	Version          int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func New(id, representativeID, title, cause, body, category string, now time.Time) (*Proposal, error) {
	if id == "" || representativeID == "" {
		return nil, fmt.Errorf("proposal identity is required")
	}
	p := &Proposal{ID: id, RepresentativeID: representativeID, Status: StatusDraft, Version: 1, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if err := p.Edit(title, cause, body, category, now); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Proposal) Edit(title, cause, body, category string, now time.Time) error {
	if p.Status != StatusDraft && len(p.Supplements) > 0 {
		return shared.NewError("PROPOSAL_IMMUTABLE", "submitted proposal cannot be edited", shared.ErrInvalidState)
	}
	if strings.TrimSpace(title) == "" || strings.TrimSpace(cause) == "" || strings.TrimSpace(body) == "" || strings.TrimSpace(category) == "" {
		return shared.NewError("PROPOSAL_INVALID", "title, cause, body and category are required", nil)
	}
	p.Title, p.Cause, p.Body, p.Category = strings.TrimSpace(title), strings.TrimSpace(cause), strings.TrimSpace(body), strings.TrimSpace(category)
	p.Version++
	p.UpdatedAt = now.UTC()
	return nil
}

func (p *Proposal) AddSupplement(s Supplement, now time.Time) error {
	if p.Status == StatusDraft || strings.TrimSpace(s.Body) == "" {
		return shared.NewError("SUPPLEMENT_NOT_ALLOWED", "supplements require a submitted proposal and non-empty body", shared.ErrInvalidState)
	}
	normalized := normalizeSupplement(s, now)
	p.Body = normalized.Body
	p.Attachments = append(p.Attachments, normalized.Files...)
	p.Supplements = append(p.Supplements, normalized)
	p.Version++
	p.UpdatedAt = now.UTC()
	return nil
}

func normalizeSupplement(value Supplement, receivedAt time.Time) Supplement {
	value.Body = strings.TrimSpace(value.Body)
	if value.CreatedAt.IsZero() {
		value.CreatedAt = receivedAt.UTC()
	} else {
		value.CreatedAt = value.CreatedAt.UTC()
	}
	files := make([]Attachment, 0, len(value.Files))
	for _, file := range value.Files {
		file.Name = strings.TrimSpace(file.Name)
		files = append(files, file)
	}
	value.Files = files
	return value
}

func (p Proposal) CurrentText() string {
	if len(p.Supplements) == 0 {
		return p.Body
	}
	return p.Supplements[len(p.Supplements)-1].Body
}

func (p *Proposal) Advance(from, to Status, now time.Time) error {
	if p.Status != from || !allowedTransition(from, to) {
		return shared.NewError("PROPOSAL_STATE_INVALID", "proposal state transition is not allowed", shared.ErrInvalidState)
	}
	p.Status = to
	p.Version++
	p.UpdatedAt = now.UTC()
	return nil
}

func allowedTransition(from, to Status) bool {
	allowed := map[Status]Status{StatusDraft: StatusSubmitted, StatusSubmitted: StatusAssigned, StatusAssigned: StatusAccepted, StatusAccepted: StatusHandling, StatusHandling: StatusAnswered, StatusAnswered: StatusEvaluated, StatusEvaluated: StatusArchived}
	return allowed[from] == to
}
