package outbox

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

type canceledBatchRepository struct {
	acked bool
}

func (r *canceledBatchRepository) Claim(context.Context, int) ([]Message, error) {
	return []Message{{ID: "message", Topic: "alerts", Payload: []byte("payload")}}, nil
}
func (r *canceledBatchRepository) Ack(context.Context, string) error {
	r.acked = true
	return nil
}
func (r *canceledBatchRepository) Retry(context.Context, string, int, time.Time, string) error {
	return nil
}
func (r *canceledBatchRepository) DeadLetter(context.Context, string, string) error { return nil }

type canceledBatchHandler struct{}

func (canceledBatchHandler) Deliver(context.Context, Message) error { return context.Canceled }

func TestCanceledBatchCannotAcknowledgeUndeliveredMessage(t *testing.T) {
	repository := &canceledBatchRepository{}
	worker := Worker{Repo: repository, Handler: canceledBatchHandler{}, Logger: zap.NewNop(), MaxAttempts: 3, Batch: 1}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := worker.runBatch(ctx); err != nil && err != context.Canceled {
		t.Fatal(err)
	}
	if repository.acked {
		t.Fatal("canceled delivery was acknowledged")
	}
}
