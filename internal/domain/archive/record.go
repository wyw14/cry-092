package archive

import "time"

type Record struct {
	ID             string
	ProposalID     string
	Year           int
	SnapshotID     string
	ReplyIDs       []string
	EvaluationIDs  []string
	AuditEventIDs  []string
	ArchivedAt     time.Time
	RetentionUntil time.Time
}

func NewRecord(id, proposalID, snapshotID string, replies, evaluations, audits []string, at time.Time, retentionYears int) Record {
	utc := at.UTC()
	return Record{ID: id, ProposalID: proposalID, Year: utc.Year(), SnapshotID: snapshotID, ReplyIDs: append([]string(nil), replies...), EvaluationIDs: append([]string(nil), evaluations...), AuditEventIDs: append([]string(nil), audits...), ArchivedAt: utc, RetentionUntil: utc.AddDate(retentionYears, 0, 0)}
}
