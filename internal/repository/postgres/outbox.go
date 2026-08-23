package postgres

import (
	"context"
	"time"

	"github.com/wyw14/cry-092/internal/platform/outbox"
)

func (s *Store) Claim(ctx context.Context, limit int) ([]outbox.Message, error) {
	var messages []outbox.Message
	err := s.WithinTransaction(ctx, func(txCtx context.Context) error {
		rows, err := s.executor(txCtx).Query(txCtx, `
			SELECT id, topic, payload, attempts
			FROM outbox_messages
			WHERE delivered_at IS NULL
			  AND dead_lettered_at IS NULL
			  AND available_at <= now()
			ORDER BY created_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT $1`, limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var message outbox.Message
			if err := rows.Scan(&message.ID, &message.Topic, &message.Payload, &message.Attempts); err != nil {
				return err
			}
			messages = append(messages, message)
		}
		return rows.Err()
	})
	return messages, err
}

func (s *Store) Ack(ctx context.Context, id string) error {
	_, err := s.executor(ctx).Exec(ctx, `
		UPDATE outbox_messages
		SET delivered_at = now(), last_error = ''
		WHERE id = $1 AND delivered_at IS NULL`, id)
	return err
}

func (s *Store) Retry(ctx context.Context, id string, attempts int, availableAt time.Time, reason string) error {
	_, err := s.executor(ctx).Exec(ctx, `
		UPDATE outbox_messages
		SET attempts = $2, available_at = $3, last_error = $4
		WHERE id = $1 AND delivered_at IS NULL AND dead_lettered_at IS NULL`,
		id, attempts, availableAt.UTC(), reason)
	return err
}

func (s *Store) DeadLetter(ctx context.Context, id, reason string) error {
	_, err := s.executor(ctx).Exec(ctx, `
		UPDATE outbox_messages
		SET dead_lettered_at = now(), last_error = $2
		WHERE id = $1 AND delivered_at IS NULL`, id, reason)
	return err
}
