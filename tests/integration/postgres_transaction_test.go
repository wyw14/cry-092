package integration

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/wyw14/cry-092/internal/repository/postgres"
)

func TestPostgresTransactionRollsBackOnApplicationError(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	store, err := postgres.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	marker := "integration-rollback-marker"
	expected := errors.New("force rollback")
	err = store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := store.Enqueue(txCtx, marker, "integration.probe", []byte(`{"probe":true}`)); err != nil {
			return err
		}
		return expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("unexpected transaction result: %v", err)
	}
	var count int
	if err := store.Pool.QueryRow(ctx, `SELECT count(*) FROM outbox_messages WHERE id=$1`, marker).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("transaction leaked %d rows", count)
	}
}
