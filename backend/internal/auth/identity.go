// Package auth verifies Keycloak access tokens and carries the authenticated identity in the request context.
package auth

import (
	"context"

	"github.com/google/uuid"
)

// Identity is derived only from a verified access token, never from request input.
type Identity struct {
	UserID      uuid.UUID
	Username    string
	DisplayName string
	Email       string
}

type ctxKey struct{}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext returns the identity set by the auth middleware; ok is false for anonymous requests.
func FromContext(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(Identity)
	return id, ok
}
