package httptransport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func TestRouterRequiresAccessToken(t *testing.T) {
	router := testRouter(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/proposals", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"code", "message", "field_errors", "request_id"} {
		if _, ok := body[field]; !ok {
			t.Fatalf("missing error field %s", field)
		}
	}
}

func TestRouterRejectsNonWhitelistedSort(t *testing.T) {
	router := testRouter(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/proposals?sort=secret_column", nil)
	request.Header.Set("Authorization", "Bearer "+signToken(t))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func testRouter(t *testing.T) http.Handler {
	t.Helper()
	return NewRouter(Dependencies{SigningKey: []byte("01234567890123456789012345678901"), Issuer: "cry-092", NewID: func() string { return "request-test" }, Ready: func(context.Context) error { return nil }, Logger: zap.NewNop()})
}

func signToken(t *testing.T) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "rep1", "iss": "cry-092", "exp": time.Now().Add(time.Hour).Unix(), "roles": []string{"representative"}})
	value, err := token.SignedString([]byte("01234567890123456789012345678901"))
	if err != nil {
		t.Fatal(err)
	}
	return value
}
