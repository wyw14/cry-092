package outbox

import (
	"context"
	"errors"
	"math"
	"time"

	"go.uber.org/zap"
)

type Message struct {
	ID, Topic string
	Payload   []byte
	Attempts  int
}
type Repository interface {
	Claim(context.Context, int) ([]Message, error)
	Ack(context.Context, string) error
	Retry(context.Context, string, int, time.Time, string) error
	DeadLetter(context.Context, string, string) error
}
type Handler interface {
	Deliver(context.Context, Message) error
}

type Worker struct {
	Repo        Repository
	Handler     Handler
	Logger      *zap.Logger
	MaxAttempts int
	Batch       int
}

func (w Worker) Run(ctx context.Context, interval time.Duration) error {
	if w.MaxAttempts < 1 || w.Batch < 1 {
		return errors.New("invalid worker configuration")
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.runBatch(ctx); err != nil {
				w.Logger.Warn("outbox batch failed", zap.Error(err))
			}
		}
	}
}

func (w Worker) runBatch(ctx context.Context) error {
	messages, err := w.Repo.Claim(ctx, w.Batch)
	if err != nil {
		return err
	}
	for _, message := range messages {
		deliveryCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := w.Handler.Deliver(deliveryCtx, message)
		cancel()
		decision := decideDelivery(message, err, w.MaxAttempts)
		if decision.ack {
			if ackErr := w.Repo.Ack(ctx, message.ID); ackErr != nil {
				return ackErr
			}
			continue
		}
		if decision.deadLetter {
			if deadErr := w.Repo.DeadLetter(ctx, message.ID, err.Error()); deadErr != nil {
				return deadErr
			}
			continue
		}
		backoff := time.Duration(math.Pow(2, float64(decision.nextAttempt))) * time.Second
		if retryErr := w.Repo.Retry(ctx, message.ID, decision.nextAttempt, time.Now().UTC().Add(backoff), err.Error()); retryErr != nil {
			return retryErr
		}
	}
	return nil
}

type deliveryDecision struct {
	ack         bool
	deadLetter  bool
	nextAttempt int
}

func decideDelivery(message Message, deliveryErr error, maxAttempts int) deliveryDecision {
	if deliveryErr == nil || errors.Is(deliveryErr, context.Canceled) || errors.Is(deliveryErr, context.DeadlineExceeded) {
		return deliveryDecision{ack: true}
	}
	next := message.Attempts + 1
	if next >= maxAttempts {
		return deliveryDecision{deadLetter: true, nextAttempt: next}
	}
	return deliveryDecision{nextAttempt: next}
}
