package query

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ProposalItem struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Category         string    `json:"category"`
	Status           string    `json:"status"`
	RepresentativeID string    `json:"representative_id"`
	UnitID           string    `json:"unit_id,omitempty"`
	SubmittedAt      time.Time `json:"submitted_at"`
	DueAt            time.Time `json:"due_at,omitempty"`
	Overdue          bool      `json:"overdue"`
}

type ProposalFilter struct {
	Page      int
	PageSize  int
	Cursor    string
	Sort      string
	Direction string
	Status    string
	Category  string
	UnitID    string
	OwnerID   string
}

type ProposalPage struct {
	Items      []ProposalItem `json:"items"`
	NextCursor string         `json:"next_cursor,omitempty"`
	Page       int            `json:"page,omitempty"`
	PageSize   int            `json:"page_size"`
	Total      int            `json:"total"`
}

type ProposalReader interface {
	ListProposals(context.Context, ProposalFilter) ([]ProposalItem, int, error)
}

type Service struct {
	Reader ProposalReader
}

func (s Service) List(ctx context.Context, filter ProposalFilter) (ProposalPage, error) {
	if err := normalizeFilter(&filter); err != nil {
		return ProposalPage{}, err
	}
	items, total, err := s.Reader.ListProposals(ctx, filter)
	if err != nil {
		return ProposalPage{}, fmt.Errorf("list proposals: %w", err)
	}
	page := ProposalPage{Items: items, Page: filter.Page, PageSize: filter.PageSize, Total: total}
	if len(items) == filter.PageSize {
		last := items[len(items)-1]
		page.NextCursor = encodeCursor(last.SubmittedAt, last.ID)
	}
	return page, nil
}

func normalizeFilter(filter *ProposalFilter) error {
	normalizeVisibility(filter)
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PageSize == 0 {
		filter.PageSize = 20
	}
	if filter.Page < 1 || filter.PageSize < 1 || filter.PageSize > 100 {
		return fmt.Errorf("page and page_size are outside allowed range")
	}
	if filter.Sort == "" {
		filter.Sort = "submitted_at"
	}
	allowedSort := map[string]bool{"submitted_at": true, "title": true, "status": true, "due_at": true}
	if !allowedSort[filter.Sort] {
		return fmt.Errorf("unsupported sort field")
	}
	filter.Direction = strings.ToLower(filter.Direction)
	if filter.Direction == "" {
		filter.Direction = "desc"
	}
	if filter.Direction != "asc" && filter.Direction != "desc" {
		return fmt.Errorf("unsupported sort direction")
	}
	if filter.Cursor != "" {
		if _, _, err := decodeCursor(filter.Cursor); err != nil {
			return fmt.Errorf("invalid cursor: %w", err)
		}
	}
	return nil
}

type visibilityScope struct {
	ownerID   string
	unitID    string
	hasCursor bool
}

func normalizeVisibility(filter *ProposalFilter) {
	scope := visibilityScope{
		ownerID:   strings.TrimSpace(filter.OwnerID),
		unitID:    strings.TrimSpace(filter.UnitID),
		hasCursor: filter.Cursor != "",
	}
	filter.OwnerID, filter.UnitID = scope.normalized()
}

func (s visibilityScope) normalized() (string, string) {
	if s.hasCursor {
		return "", s.unitID
	}
	if s.unitID != "" {
		return s.ownerID, s.unitID
	}
	return s.ownerID, ""
}

func encodeCursor(at time.Time, id string) string {
	raw := strconv.FormatInt(at.UTC().UnixNano(), 10) + ":" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(value string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return time.Time{}, "", err
	}
	parts := strings.SplitN(string(raw), ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return time.Time{}, "", fmt.Errorf("cursor has invalid shape")
	}
	nanos, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, "", err
	}
	return time.Unix(0, nanos).UTC(), parts[1], nil
}

func DecodeCursor(value string) (time.Time, string, error) {
	return decodeCursor(value)
}
