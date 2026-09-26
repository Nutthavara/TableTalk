// Package middleware contains HTTP middleware shared by all routes.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
)

type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (auth.Identity, error)
}

type UserSyncer interface {
	SyncUser(ctx context.Context, id auth.Identity) error
}

type Authenticator struct {
	verifier TokenVerifier
	users    UserSyncer
}

func NewAuthenticator(verifier TokenVerifier, users UserSyncer) *Authenticator {
	return &Authenticator{verifier: verifier, users: users}
}

// Required rejects requests without a valid bearer token with 401.
func (a *Authenticator) Required(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, present := bearerToken(r)
		if !present {
			apperr.Write(w, r, apperr.ErrUnauthorized)
			return
		}
		a.serveAuthenticated(w, r, raw, next)
	})
}

// Optional lets anonymous requests through but still rejects a token that is present and invalid,
// so a client never silently gets anonymous results because of a bad token.
func (a *Authenticator) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, present := bearerToken(r)
		if !present {
			next.ServeHTTP(w, r)
			return
		}
		a.serveAuthenticated(w, r, raw, next)
	})
}

func (a *Authenticator) serveAuthenticated(w http.ResponseWriter, r *http.Request, raw string, next http.Handler) {
	id, err := a.verifier.Verify(r.Context(), raw)
	if err != nil {
		apperr.Write(w, r, apperr.ErrUnauthorized)
		return
	}
	if err := a.users.SyncUser(r.Context(), id); err != nil {
		apperr.Write(w, r, fmt.Errorf("sync user: %w", err))
		return
	}
	next.ServeHTTP(w, r.WithContext(auth.WithIdentity(r.Context(), id)))
}

// bearerToken reports present=true whenever an Authorization header was sent, even if malformed,
// so a malformed header is treated as an invalid token rather than as anonymous.
func bearerToken(r *http.Request) (token string, present bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	scheme, token, ok := strings.Cut(h, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", true
	}
	return strings.TrimSpace(token), true
}
