package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wyw14/cry-092/internal/domain/proposal"
	"github.com/wyw14/cry-092/internal/domain/shared"
)

func (s *Store) Create(ctx context.Context, p *proposal.Proposal) error {
	attachments, _ := json.Marshal(p.Attachments)
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO proposals(id,representative_id,title,cause,body,category,attachments,status,version,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, p.ID, p.RepresentativeID, p.Title, p.Cause, p.Body, p.Category, attachments, p.Status, p.Version, p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *Store) Get(ctx context.Context, id string) (*proposal.Proposal, error) {
	var p proposal.Proposal
	var attachments []byte
	var submittedAt *time.Time
	err := s.executor(ctx).QueryRow(ctx, `SELECT id,representative_id,title,cause,body,category,attachments,status,submitted_at,version,created_at,updated_at FROM proposals WHERE id=$1`, id).Scan(&p.ID, &p.RepresentativeID, &p.Title, &p.Cause, &p.Body, &p.Category, &attachments, &p.Status, &submittedAt, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if err := json.Unmarshal(attachments, &p.Attachments); err != nil {
		return nil, err
	}
	p.SubmittedAt = submittedAt
	return &p, nil
}

func (s *Store) Save(ctx context.Context, p *proposal.Proposal, expectedVersion int64) error {
	attachments, _ := json.Marshal(p.Attachments)
	tag, err := s.executor(ctx).Exec(ctx, `UPDATE proposals SET title=$2,cause=$3,body=$4,category=$5,attachments=$6,status=$7,submitted_at=$8,version=$9,updated_at=$10 WHERE id=$1 AND version=$11`, p.ID, p.Title, p.Cause, p.Body, p.Category, attachments, p.Status, p.SubmittedAt, p.Version, p.UpdatedAt, expectedVersion)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return shared.ErrVersionConflict
	}
	return nil
}

func (s *Store) ListInvitations(ctx context.Context, proposalID string) ([]proposal.CosponsorInvitation, error) {
	rows, err := s.executor(ctx).Query(ctx, `SELECT id,proposal_id,representative_id,status,invited_at,responded_at,version FROM cosponsor_invitations WHERE proposal_id=$1 ORDER BY invited_at,id`, proposalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []proposal.CosponsorInvitation
	for rows.Next() {
		var i proposal.CosponsorInvitation
		if err := rows.Scan(&i.ID, &i.ProposalID, &i.RepresentativeID, &i.Status, &i.InvitedAt, &i.RespondedAt, &i.Version); err != nil {
			return nil, err
		}
		result = append(result, i)
	}
	return result, rows.Err()
}

func (s *Store) SaveSnapshot(ctx context.Context, snap proposal.SubmissionSnapshot) error {
	cosponsors, _ := json.Marshal(snap.Cosponsors)
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO proposal_snapshots(id,proposal_id,title,cause,body,category,cosponsors,submitted_by,submitted_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, snap.ID, snap.ProposalID, snap.Title, snap.Cause, snap.Body, snap.Category, cosponsors, snap.SubmittedBy, snap.SubmittedAt)
	return err
}

func (s *Store) Append(ctx context.Context, event shared.AuditEvent) error {
	before, _ := json.Marshal(event.Before)
	after, _ := json.Marshal(event.After)
	_, err := s.executor(ctx).Exec(ctx, `INSERT INTO audit_events(id,aggregate,aggregate_id,actor_id,source,before_data,after_data,reason,occurred_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, event.ID, event.Aggregate, event.AggregateID, event.ActorID, event.Source, before, after, event.Reason, event.OccurredAt)
	return err
}
