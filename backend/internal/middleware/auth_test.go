package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/auth/authtest"
	"tabletalk/backend/internal/middleware"
)

type fakeUsers struct{ synced []auth.Identity }

func (f *fakeUsers) SyncUser(_ context.Context, id auth.Identity) error {
	f.synced = append(f.synced, id)
	return nil
}

// echoIdentity responds with the user ID from the context, or "anonymous".
var echoIdentity = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if id, ok := auth.FromContext(r.Context()); ok {
		_, _ = w.Write([]byte(id.UserID.String()))
		return
	}
	_, _ = w.Write([]byte("anonymous"))
})

func do(h http.Handler, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRequiredAuth(t *testing.T) {
	signer := authtest.NewSigner(t)
	otherSigner := authtest.NewSigner(t)
	userID := uuid.New()

	withClaim := func(key string, value any) string {
		c := authtest.Claims(userID, "alice")
		c[key] = value
		return signer.Sign(t, c)
	}

	tests := []struct {
		name          string
		authorization string
		wantStatus    int
	}{
		{"valid token", "Bearer " + signer.Token(t, userID, "alice"), http.StatusOK},
		{"lowercase scheme", "bearer " + signer.Token(t, userID, "alice"), http.StatusOK},
		{"audience array containing api", "Bearer " + withClaim("aud", []string{"account", authtest.Audience}), http.StatusOK},
		{"missing header", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic dXNlcjpwYXNz", http.StatusUnauthorized},
		{"garbage token", "Bearer not-a-jwt", http.StatusUnauthorized},
		{"signed by unknown key", "Bearer " + otherSigner.Token(t, userID, "alice"), http.StatusUnauthorized},
		{"wrong issuer", "Bearer " + withClaim("iss", "http://evil.test/realms/tabletalk"), http.StatusUnauthorized},
		{"wrong audience", "Bearer " + withClaim("aud", "account"), http.StatusUnauthorized},
		{"expired", "Bearer " + withClaim("exp", time.Now().Add(-time.Minute).Unix()), http.StatusUnauthorized},
		{"subject not a uuid", "Bearer " + withClaim("sub", "alice"), http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			users := &fakeUsers{}
			h := middleware.NewAuthenticator(signer.Verifier(), users).Required(echoIdentity)

			rec := do(h, tc.authorization)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.wantStatus, rec.Body)
			}
			if tc.wantStatus == http.StatusOK {
				if rec.Body.String() != userID.String() {
					t.Fatalf("identity = %q, want %q", rec.Body, userID)
				}
				if len(users.synced) != 1 || users.synced[0].Username != "alice" {
					t.Fatalf("user not synced: %+v", users.synced)
				}
				return
			}
			if len(users.synced) != 0 {
				t.Fatalf("rejected request must not sync a user")
			}
			var body struct {
				Error struct{ Code string } `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error.Code != "UNAUTHORIZED" {
				t.Fatalf("unexpected error body %s", rec.Body)
			}
			if rec.Header().Get("WWW-Authenticate") == "" {
				t.Fatalf("missing WWW-Authenticate header")
			}
		})
	}
}

func TestOptionalAuth(t *testing.T) {
	signer := authtest.NewSigner(t)
	userID := uuid.New()
	h := middleware.NewAuthenticator(signer.Verifier(), &fakeUsers{}).Optional(echoIdentity)

	if rec := do(h, ""); rec.Code != http.StatusOK || rec.Body.String() != "anonymous" {
		t.Fatalf("anonymous: got %d %q", rec.Code, rec.Body)
	}
	if rec := do(h, "Bearer "+signer.Token(t, userID, "bob")); rec.Code != http.StatusOK || rec.Body.String() != userID.String() {
		t.Fatalf("authenticated: got %d %q", rec.Code, rec.Body)
	}
	if rec := do(h, "Bearer not-a-jwt"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token must be rejected, got %d", rec.Code)
	}
}
