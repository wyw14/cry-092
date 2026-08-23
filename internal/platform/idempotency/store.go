package idempotency

import (
	"context"
	"crypto/sha256"
	"sync"
	"time"

	"github.com/wyw14/cry-092/internal/domain/shared"
)

type Result struct {
	Status int
	Body   []byte
}
type Entry struct {
	RequestHash [32]byte
	Result      Result
	ExpiresAt   time.Time
}

type Store interface {
	Get(context.Context, string) (Entry, bool, error)
	Put(context.Context, string, Entry) error
}

type Memory struct {
	mu      sync.Mutex
	entries map[string]Entry
}

func NewMemory() *Memory { return &Memory{entries: map[string]Entry{}} }

func (m *Memory) Get(ctx context.Context, key string) (Entry, bool, error) {
	if err := ctx.Err(); err != nil {
		return Entry{}, false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.entries[key]
	if ok && time.Now().UTC().After(e.ExpiresAt) {
		delete(m.entries, key)
		return Entry{}, false, nil
	}
	e.Result.Body = append([]byte(nil), e.Result.Body...)
	return e, ok, nil
}

func (m *Memory) Put(ctx context.Context, key string, entry Entry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.entries[key]; ok && existing.RequestHash != entry.RequestHash {
		return shared.ErrIdempotencyReuse
	}
	entry.Result.Body = append([]byte(nil), entry.Result.Body...)
	m.entries[key] = entry
	return nil
}

func Hash(data []byte) [32]byte { return sha256.Sum256(data) }
