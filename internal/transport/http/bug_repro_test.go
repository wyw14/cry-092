package httptransport

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	queryapp "github.com/wyw14/cry-092/internal/application/query"
)

type scopeRecordingReader struct {
	filter queryapp.ProposalFilter
	called bool
}

func (r *scopeRecordingReader) ListProposals(_ context.Context, filter queryapp.ProposalFilter) ([]queryapp.ProposalItem, int, error) {
	r.filter, r.called = filter, true
	return []queryapp.ProposalItem{}, 0, nil
}

func TestRepresentativeCursorListKeepsOwnerScope(t *testing.T) {
	reader := &scopeRecordingReader{}
	router := NewRouter(Dependencies{
		Query:      queryapp.Service{Reader: reader},
		SigningKey: []byte("01234567890123456789012345678901"), Issuer: "cry-092",
		NewID: func() string { return "request-owner" }, Ready: func(context.Context) error { return nil },
	})
	cursorRaw := strconv.FormatInt(time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC).UnixNano(), 10) + ":proposal-20"
	cursor := base64.RawURLEncoding.EncodeToString([]byte(cursorRaw))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/proposals?cursor="+cursor+"&page_size=20", nil)
	request.Header.Set("Authorization", "Bearer "+signToken(t))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("cursor list failed: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if !reader.called || reader.filter.OwnerID != "rep1" {
		t.Fatalf("representative ownership scope was lost: called=%v filter=%+v", reader.called, reader.filter)
	}
}
