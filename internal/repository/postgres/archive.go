package postgres

import (
	"context"

	"github.com/wyw14/cry-092/internal/domain/archive"
)

func (s *Store) Aggregates(ctx context.Context, period archive.Period) ([]archive.UnitAggregate, error) {
	rows, err := s.executor(ctx).Query(ctx, `SELECT a.unit_id,count(*) FILTER(WHERE p.status IN ('answered','evaluated','archived')),count(*) FILTER(WHERE r.submitted_at<=hp.due_at),count(*) FILTER(WHERE e.rating='satisfied'),count(*) FILTER(WHERE e.rating='mostly_satisfied') FROM assignments a JOIN proposals p ON p.id=a.proposal_id LEFT JOIN handling_plans hp ON hp.proposal_id=p.id LEFT JOIN replies r ON r.proposal_id=p.id LEFT JOIN evaluations e ON e.proposal_id=p.id WHERE p.submitted_at >= $1 AND p.submitted_at < $2 GROUP BY a.unit_id`, period.Start.UTC(), period.End.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []archive.UnitAggregate
	for rows.Next() {
		var a archive.UnitAggregate
		if err := rows.Scan(&a.UnitID, &a.Eligible, &a.AnsweredOnTime, &a.Satisfied, &a.MostlySatisfied); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) SaveRanking(ctx context.Context, entries []archive.RankingEntry) error {
	for _, e := range entries {
		_, err := s.executor(ctx).Exec(ctx, `INSERT INTO unit_rankings(unit_id,period_start,period_end,period_label,numerator,denominator,score_basis_points,position) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(unit_id,period_start,period_end) DO UPDATE SET period_label=EXCLUDED.period_label,numerator=EXCLUDED.numerator,denominator=EXCLUDED.denominator,score_basis_points=EXCLUDED.score_basis_points,position=EXCLUDED.position`, e.UnitID, e.Period.Start.UTC(), e.Period.End.UTC(), e.Period.Label, e.Numerator, e.Denominator, e.ScoreBasisPoints, e.Position)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SaveRecord(ctx context.Context, r archive.Record) error {
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO archive_records(id,proposal_id,year,snapshot_id,reply_ids,evaluation_ids,audit_event_ids,archived_at,retention_until) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, r.ID, r.ProposalID, r.Year, r.SnapshotID, r.ReplyIDs, r.EvaluationIDs, r.AuditEventIDs, r.ArchivedAt, r.RetentionUntil)
	return err
}
