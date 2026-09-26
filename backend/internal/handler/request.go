package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
)

const maxJSONBody = 1 << 20

// decodeJSON reads a single JSON object. Unknown fields are rejected, so fields such as owner_id or
// user_id cannot be smuggled in; identity always comes from the token.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return apperr.Validation(map[string]string{"body": jsonErrorMessage(err)})
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return apperr.Validation(map[string]string{"body": "Body must contain a single JSON object."})
	}
	return nil
}

func jsonErrorMessage(err error) string {
	var typeErr *json.UnmarshalTypeError
	var maxErr *http.MaxBytesError
	switch {
	case errors.As(err, &typeErr):
		return fmt.Sprintf("Field %q has the wrong type.", typeErr.Field)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return "Unknown field " + strings.TrimPrefix(err.Error(), "json: unknown field ") + "."
	case errors.As(err, &maxErr):
		return "Request body is too large."
	case errors.Is(err, io.EOF):
		return "Request body is required."
	default:
		return "Request body must be valid JSON."
	}
}

// pathID parses a UUID path parameter; malformed IDs are reported as not found.
func pathID(r *http.Request, name string, notFound error) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		return uuid.Nil, notFound
	}
	return id, nil
}

// identity returns the authenticated user. Routes using it are wrapped in Required auth.
func identity(r *http.Request) (auth.Identity, error) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		return auth.Identity{}, apperr.ErrUnauthorized
	}
	return id, nil
}
