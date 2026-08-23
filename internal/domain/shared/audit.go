package shared

import "time"

type AuditSource string

const (
	AuditHTTP   AuditSource = "http"
	AuditWorker AuditSource = "worker"
	AuditSeed   AuditSource = "seed"
)

type AuditEvent struct {
	ID          string
	Aggregate   string
	AggregateID string
	ActorID     string
	Source      AuditSource
	Before      map[string]any
	After       map[string]any
	Reason      string
	OccurredAt  time.Time
}

func NewAuditEvent(id, aggregate, aggregateID, actorID string, source AuditSource, before, after map[string]any, reason string, at time.Time) AuditEvent {
	return AuditEvent{ID: id, Aggregate: aggregate, AggregateID: aggregateID, ActorID: actorID, Source: source, Before: before, After: after, Reason: reason, OccurredAt: at.UTC()}
}
