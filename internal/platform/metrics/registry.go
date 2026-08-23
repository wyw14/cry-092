package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Registry struct {
	requests sync.Map
	errors   sync.Map
	inflight atomic.Int64
	started  time.Time
}

func NewRegistry(now time.Time) *Registry {
	return &Registry{started: now.UTC()}
}

func (r *Registry) Observe(method, route string, status int) {
	key := sanitize(method) + "_" + sanitize(route) + "_" + fmt.Sprintf("%d", status)
	counter, _ := r.requests.LoadOrStore(key, &atomic.Int64{})
	counter.(*atomic.Int64).Add(1)
	if status >= 500 {
		errorCounter, _ := r.errors.LoadOrStore(key, &atomic.Int64{})
		errorCounter.(*atomic.Int64).Add(1)
	}
}

func (r *Registry) Track(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		r.inflight.Add(1)
		defer r.inflight.Add(-1)
		next.ServeHTTP(writer, request)
	})
}

func (r *Registry) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4")
	lines := []string{fmt.Sprintf("cry092_inflight_requests %d", r.inflight.Load()), fmt.Sprintf("cry092_uptime_seconds %.0f", time.Since(r.started).Seconds())}
	r.requests.Range(func(key, value any) bool {
		lines = append(lines, fmt.Sprintf("cry092_http_requests_total{key=%q} %d", key.(string), value.(*atomic.Int64).Load()))
		return true
	})
	r.errors.Range(func(key, value any) bool {
		lines = append(lines, fmt.Sprintf("cry092_http_errors_total{key=%q} %d", key.(string), value.(*atomic.Int64).Load()))
		return true
	})
	sort.Strings(lines)
	_, _ = writer.Write([]byte(strings.Join(lines, "\n") + "\n"))
}

func sanitize(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer("/", "_", "-", "_", " ", "_")
	return strings.Trim(replacer.Replace(value), "_")
}
