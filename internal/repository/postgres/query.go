package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	queryapp "github.com/wyw14/cry-092/internal/application/query"
)

func (s *Store) ListProposals(ctx context.Context, filter queryapp.ProposalFilter) ([]queryapp.ProposalItem, int, error) {
	clauses := []string{"1=1"}
	arguments := make([]any, 0, 8)
	bind := func(value any) string {
		arguments = append(arguments, value)
		return fmt.Sprintf("$%d", len(arguments))
	}
	if filter.Status != "" {
		clauses = append(clauses, "p.status = "+bind(filter.Status))
	}
	if filter.Category != "" {
		clauses = append(clauses, "p.category = "+bind(filter.Category))
	}
	if filter.UnitID != "" {
		clauses = append(clauses, "a.unit_id = "+bind(filter.UnitID))
	}
	if filter.OwnerID != "" {
		clauses = append(clauses, "p.representative_id = "+bind(filter.OwnerID))
	}
	if filter.Cursor != "" {
		cursorTime, cursorID, err := decodeQueryCursor(filter.Cursor)
		if err != nil {
			return nil, 0, err
		}
		operator := "<"
		if filter.Direction == "asc" {
			operator = ">"
		}
		clauses = append(clauses, fmt.Sprintf("(p.submitted_at, p.id) %s (%s, %s)", operator, bind(cursorTime), bind(cursorID)))
	}
	columns := map[string]string{"submitted_at": "p.submitted_at", "title": "p.title", "status": "p.status", "due_at": "hp.due_at"}
	sortColumn, ok := columns[filter.Sort]
	if !ok {
		return nil, 0, fmt.Errorf("sort field is not allowed")
	}
	direction := "DESC"
	if filter.Direction == "asc" {
		direction = "ASC"
	}
	where := strings.Join(clauses, " AND ")
	countQuery := `SELECT count(*) FROM proposals p LEFT JOIN assignments a ON a.proposal_id=p.id LEFT JOIN handling_plans hp ON hp.proposal_id=p.id WHERE ` + where
	var total int
	if err := s.executor(ctx).QueryRow(ctx, countQuery, arguments...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limitBind := bind(filter.PageSize)
	offsetBind := bind((filter.Page - 1) * filter.PageSize)
	listQuery := `SELECT p.id,p.title,p.category,p.status,p.representative_id,coalesce(a.unit_id,''),p.submitted_at,hp.due_at,coalesce(hp.due_at<now() AND hp.status<>'completed',false) FROM proposals p LEFT JOIN assignments a ON a.proposal_id=p.id LEFT JOIN handling_plans hp ON hp.proposal_id=p.id WHERE ` + where + ` ORDER BY ` + sortColumn + ` ` + direction + ` NULLS LAST,p.id ` + direction + ` LIMIT ` + limitBind + ` OFFSET ` + offsetBind
	rows, err := s.executor(ctx).Query(ctx, listQuery, arguments...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]queryapp.ProposalItem, 0, filter.PageSize)
	for rows.Next() {
		var item queryapp.ProposalItem
		var dueAt *time.Time
		if err := rows.Scan(&item.ID, &item.Title, &item.Category, &item.Status, &item.RepresentativeID, &item.UnitID, &item.SubmittedAt, &dueAt, &item.Overdue); err != nil {
			return nil, 0, err
		}
		if dueAt != nil {
			item.DueAt = dueAt.UTC()
		}
		item.SubmittedAt = item.SubmittedAt.UTC()
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func decodeQueryCursor(value string) (time.Time, string, error) { return queryapp.DecodeCursor(value) }
