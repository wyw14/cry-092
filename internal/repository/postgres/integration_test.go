package postgres

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestPostgresRepositoryPing(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	store, err := Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}
}
