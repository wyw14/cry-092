package idempotency

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestConcurrentIdempotentReplayKeepsOneResult(t *testing.T) {
	store := NewMemory()
	start := make(chan struct{})
	var wait sync.WaitGroup
	errors := make(chan error, 8)
	entry := Entry{RequestHash: Hash([]byte("same request")), Result: Result{Status: 201, Body: []byte(`{"id":"p1"}`)}, ExpiresAt: time.Now().Add(time.Hour)}
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() { defer wait.Done(); <-start; errors <- store.Put(context.Background(), "submit:p1:key", entry) }()
	}
	close(start)
	wait.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	got, ok, err := store.Get(context.Background(), "submit:p1:key")
	if err != nil || !ok || string(got.Result.Body) != `{"id":"p1"}` {
		t.Fatalf("unexpected replay result: %+v %v %v", got, ok, err)
	}
}
