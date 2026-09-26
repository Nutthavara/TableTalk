// Package authtest issues RS256 tokens signed by an in-memory key for tests.
package authtest

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/google/uuid"

	"tabletalk/backend/internal/auth"
)

const (
	Issuer   = "http://keycloak.test/realms/tabletalk"
	Audience = "tabletalk-api"
)

// Signer mints tokens; Verifier returns an auth.Verifier that trusts only this signer.
type Signer struct {
	key *rsa.PrivateKey
}

func NewSigner(t testing.TB) *Signer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return &Signer{key: key}
}

func (s *Signer) Verifier() *auth.Verifier {
	return auth.NewVerifier(Issuer, Audience, &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&s.key.PublicKey}})
}

// Claims builds valid claims for a user; tests mutate the map to produce invalid tokens.
func Claims(userID uuid.UUID, username string) map[string]any {
	now := time.Now()
	return map[string]any{
		"iss":                Issuer,
		"aud":                Audience,
		"sub":                userID.String(),
		"iat":                now.Unix(),
		"exp":                now.Add(5 * time.Minute).Unix(),
		"preferred_username": username,
		"name":               username,
		"email":              username + "@example.com",
	}
}

func (s *Signer) Sign(t testing.TB, claims map[string]any) string {
	t.Helper()
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: s.key}, nil)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	obj, err := signer.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}
	token, err := obj.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}
	return token
}

// Token is a shortcut for a valid token for the given user.
func (s *Signer) Token(t testing.TB, userID uuid.UUID, username string) string {
	return s.Sign(t, Claims(userID, username))
}
