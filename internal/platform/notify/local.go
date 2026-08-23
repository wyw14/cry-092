package notify

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Delivery struct {
	Recipient string
	Template  string
	Fields    map[string]string
	SentAt    time.Time
}

type LocalAdapter struct {
	mu         sync.RWMutex
	deliveries []Delivery
	clock      func() time.Time
}

func NewLocalAdapter(clock func() time.Time) *LocalAdapter {
	return &LocalAdapter{clock: clock}
}

func (a *LocalAdapter) Notify(ctx context.Context, recipient, template string, fields map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(recipient) == "" || strings.TrimSpace(template) == "" {
		return fmt.Errorf("notification recipient and template are required")
	}
	copyFields := make(map[string]string, len(fields))
	for key, value := range fields {
		if strings.Contains(strings.ToLower(key), "password") || strings.Contains(strings.ToLower(key), "token") {
			return fmt.Errorf("sensitive notification field rejected")
		}
		copyFields[key] = value
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.deliveries = append(a.deliveries, Delivery{Recipient: recipient, Template: template, Fields: copyFields, SentAt: a.clock().UTC()})
	return nil
}

func (a *LocalAdapter) Deliveries() []Delivery {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]Delivery, len(a.deliveries))
	for index, delivery := range a.deliveries {
		result[index] = delivery
		result[index].Fields = make(map[string]string, len(delivery.Fields))
		for key, value := range delivery.Fields {
			result[index].Fields[key] = value
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].SentAt.Equal(result[j].SentAt) {
			return result[i].Recipient < result[j].Recipient
		}
		return result[i].SentAt.Before(result[j].SentAt)
	})
	return result
}
