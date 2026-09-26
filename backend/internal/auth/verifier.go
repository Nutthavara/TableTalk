package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid access token")

type Verifier struct {
	verifier *oidc.IDTokenVerifier
}

// NewVerifier checks signature (via keySet), issuer, audience and expiry of RS256 JWT access tokens.
func NewVerifier(issuer, audience string, keySet oidc.KeySet) *Verifier {
	return &Verifier{verifier: oidc.NewVerifier(issuer, keySet, &oidc.Config{
		ClientID:             audience,
		SupportedSigningAlgs: []string{oidc.RS256},
	})}
}

// NewKeycloakVerifier fetches (and caches/rotates) signing keys from the realm's JWKS endpoint on demand,
// so the API can start before Keycloak is ready.
func NewKeycloakVerifier(ctx context.Context, issuer, audience string) *Verifier {
	jwksURL := strings.TrimRight(issuer, "/") + "/protocol/openid-connect/certs"
	return NewVerifier(issuer, audience, oidc.NewRemoteKeySet(ctx, jwksURL))
}

type claims struct {
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	Email             string `json:"email"`
}

func (v *Verifier) Verify(ctx context.Context, rawToken string) (Identity, error) {
	tok, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return Identity{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	userID, err := uuid.Parse(tok.Subject)
	if err != nil {
		return Identity{}, fmt.Errorf("%w: subject is not a UUID", ErrInvalidToken)
	}
	var c claims
	if err := tok.Claims(&c); err != nil {
		return Identity{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	id := Identity{UserID: userID, Username: c.PreferredUsername, DisplayName: c.Name, Email: c.Email}
	if id.Username == "" {
		id.Username = tok.Subject
	}
	if id.DisplayName == "" {
		id.DisplayName = id.Username
	}
	return id, nil
}
