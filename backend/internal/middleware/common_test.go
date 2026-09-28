package middleware_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tabletalk/backend/internal/middleware"
)

func TestRecover_DoesNotLeakInternals(t *testing.T) {
	slog.SetDefault(slog.New(slog.DiscardHandler))
	h := middleware.Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("db password is hunter2")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/x", nil))

	body, _ := io.ReadAll(rec.Body)
	if rec.Code != http.StatusInternalServerError || strings.Contains(string(body), "hunter2") {
		t.Fatalf("got %d %s", rec.Code, body)
	}
	var parsed struct {
		Error struct{ Code string } `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected shared error format, got %s", body)
	}
}

func TestSecurityHeaders(t *testing.T) {
	h := middleware.SecurityHeaders(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil))
	for header, want := range map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Cache-Control":          "no-store",
	} {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}
