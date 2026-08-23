package postgres

import (
	"context"
	"time"

	"github.com/wyw14/cry-092/internal/domain/assignment"
	"github.com/wyw14/cry-092/internal/domain/handling"
	"github.com/wyw14/cry-092/internal/domain/response"
	"github.com/wyw14/cry-092/internal/domain/review"
	"github.com/wyw14/cry-092/internal/domain/supervision"
)

func (s *Store) SaveAssignment(ctx context.Context, a *assignment.Assignment) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO assignments(id,proposal_id,unit_id,rule_id,status,assigned_by,accepted_by,assigned_at,accepted_at,reason,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) ON CONFLICT(id) DO UPDATE SET unit_id=EXCLUDED.unit_id,status=EXCLUDED.status,accepted_by=EXCLUDED.accepted_by,accepted_at=EXCLUDED.accepted_at,reason=EXCLUDED.reason,version=EXCLUDED.version WHERE assignments.version<EXCLUDED.version`, a.ID, a.ProposalID, a.UnitID, a.RuleID, a.Status, a.AssignedBy, a.AcceptedBy, a.AssignedAt, a.AcceptedAt, a.Reason, a.Version)
	return err
}

func (s *Store) GetByProposal(ctx context.Context, proposalID string) (*assignment.Assignment, error) {
	var a assignment.Assignment
	err := s.executor(ctx).QueryRow(ctx, `SELECT id,proposal_id,unit_id,rule_id,status,assigned_by,accepted_by,assigned_at,accepted_at,reason,version FROM assignments WHERE proposal_id=$1 ORDER BY assigned_at DESC LIMIT 1`, proposalID).Scan(&a.ID, &a.ProposalID, &a.UnitID, &a.RuleID, &a.Status, &a.AssignedBy, &a.AcceptedBy, &a.AssignedAt, &a.AcceptedAt, &a.Reason, &a.Version)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &a, nil
}

func (s *Store) Rules(ctx context.Context) ([]assignment.Rule, error) {
	rows, err := s.executor(ctx).Query(ctx, `SELECT id,category,unit_id,priority,active FROM assignment_rules WHERE active=true ORDER BY priority DESC,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []assignment.Rule
	for rows.Next() {
		var r assignment.Rule
		if err := rows.Scan(&r.ID, &r.Category, &r.UnitID, &r.Priority, &r.Active); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) SavePlan(ctx context.Context, p *handling.Plan) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO handling_plans(id,proposal_id,assignment_id,unit_id,steps,materials,due_at,status,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(id) DO UPDATE SET steps=EXCLUDED.steps,materials=EXCLUDED.materials,due_at=EXCLUDED.due_at,status=EXCLUDED.status,version=EXCLUDED.version WHERE handling_plans.version<EXCLUDED.version`, p.ID, p.ProposalID, p.AssignmentID, p.UnitID, p.Steps, p.Materials, p.DueAt, p.Status, p.Version)
	return err
}

func (s *Store) GetPlan(ctx context.Context, idOrProposal string) (*handling.Plan, error) {
	var p handling.Plan
	err := s.executor(ctx).QueryRow(ctx, `SELECT id,proposal_id,assignment_id,unit_id,steps,materials,due_at,status,version FROM handling_plans WHERE id=$1 OR proposal_id=$1 ORDER BY id LIMIT 1`, idOrProposal).Scan(&p.ID, &p.ProposalID, &p.AssignmentID, &p.UnitID, &p.Steps, &p.Materials, &p.DueAt, &p.Status, &p.Version)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &p, nil
}

func (s *Store) SaveExtension(ctx context.Context, e *handling.Extension) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO extension_requests(id,plan_id,requested_days,reason,status,requested_by,reviewed_by,requested_at,reviewed_at,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, e.ID, e.PlanID, e.RequestedDays, e.Reason, e.Status, e.RequestedBy, e.ReviewedBy, e.RequestedAt, e.ReviewedAt, e.Version)
	return err
}

func (s *Store) SaveReply(ctx context.Context, r *response.Reply) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO replies(id,proposal_id,round,kind,summary,file_ids,submitted_by,submitted_at,viewed_at,supplements,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, r.ID, r.ProposalID, r.Round, r.Kind, r.Summary, r.FileIDs, r.SubmittedBy, r.SubmittedAt, r.ViewedAt, r.Supplements, r.Version)
	return err
}

func (s *Store) GetReply(ctx context.Context, id string) (*response.Reply, error) {
	var r response.Reply
	err := s.executor(ctx).QueryRow(ctx, `SELECT id,proposal_id,round,kind,summary,file_ids,submitted_by,submitted_at,viewed_at,supplements,version FROM replies WHERE id=$1`, id).Scan(&r.ID, &r.ProposalID, &r.Round, &r.Kind, &r.Summary, &r.FileIDs, &r.SubmittedBy, &r.SubmittedAt, &r.ViewedAt, &r.Supplements, &r.Version)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &r, nil
}

func (s *Store) ListByProposal(ctx context.Context, proposalID string) ([]response.Reply, error) {
	rows, err := s.executor(ctx).Query(ctx, `SELECT id,proposal_id,round,kind,summary,file_ids,submitted_by,submitted_at,viewed_at,supplements,version FROM replies WHERE proposal_id=$1 ORDER BY round`, proposalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []response.Reply
	for rows.Next() {
		var r response.Reply
		if err := rows.Scan(&r.ID, &r.ProposalID, &r.Round, &r.Kind, &r.Summary, &r.FileIDs, &r.SubmittedBy, &r.SubmittedAt, &r.ViewedAt, &r.Supplements, &r.Version); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) SaveEvaluation(ctx context.Context, e *review.Evaluation) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO evaluations(id,proposal_id,reply_id,round,representative_id,rating,comment,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, e.ID, e.ProposalID, e.ReplyID, e.Round, e.RepresentativeID, e.Rating, e.Comment, e.CreatedAt)
	return err
}
func (s *Store) SaveRework(ctx context.Context, r review.ReworkRound) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO rework_rounds(id,proposal_id,previous_reply_id,round,unit_id,reason,created_at,completed_at,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, r.ID, r.ProposalID, r.PreviousReplyID, r.Round, r.UnitID, r.Reason, r.CreatedAt, r.CompletedAt, r.Version)
	return err
}
func (s *Store) SaveReminder(ctx context.Context, r supervision.Reminder) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO reminders(id,proposal_id,unit_id,period_start,level,attempt,delivered_at,dead_lettered_at,last_error,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT(proposal_id,period_start,level) DO NOTHING`, r.ID, r.ProposalID, r.UnitID, r.PeriodStart, r.Level, r.Attempt, r.DeliveredAt, r.DeadLetteredAt, r.LastError, r.CreatedAt)
	return err
}
func (s *Store) SaveCase(ctx context.Context, c supervision.Case) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO supervision_cases(id,proposal_id,supervisor_id,level,opinions,opened_at,closed_at,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, c.ID, c.ProposalID, c.SupervisorID, c.Level, c.Opinions, c.OpenedAt, c.ClosedAt, c.Version)
	return err
}
func (s *Store) GetCaseByProposal(ctx context.Context, proposalID string) (*supervision.Case, error) {
	var c supervision.Case
	err := s.executor(ctx).QueryRow(ctx, `SELECT id,proposal_id,supervisor_id,level,opinions,opened_at,closed_at,version FROM supervision_cases WHERE proposal_id=$1 ORDER BY opened_at DESC LIMIT 1`, proposalID).Scan(&c.ID, &c.ProposalID, &c.SupervisorID, &c.Level, &c.Opinions, &c.OpenedAt, &c.ClosedAt, &c.Version)
	if err != nil {
		return nil, mapNotFound(err)
	}
	return &c, nil
}

var _ = time.Time{}
