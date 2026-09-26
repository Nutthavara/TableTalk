// Package apperr defines the application errors that the HTTP layer maps to status codes and the
// shared error response body.
package apperr

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Error struct {
	Status  int
	Code    string
	Message string
	Details map[string]string
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// Validation returns a 400 with per-field messages.
func Validation(details map[string]string) *Error {
	return &Error{
		Status:  http.StatusBadRequest,
		Code:    "VALIDATION_ERROR",
		Message: "Some fields are invalid.",
		Details: details,
	}
}

var (
	ErrUnauthorized = New(http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required.")
	ErrForbidden    = New(http.StatusForbidden, "FORBIDDEN", "You do not have permission to perform this action.")
	ErrNotFound     = New(http.StatusNotFound, "NOT_FOUND", "The requested resource was not found.")
	ErrInternal     = New(http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong. Please try again.")
)

// As returns the *Error in err's chain, or ErrInternal when there is none.
func As(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return ErrInternal
}

type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Write sends err in the shared error format. Unexpected errors are logged and reported as a generic 500
// so internal details never reach the client.
func Write(w http.ResponseWriter, r *http.Request, err error) {
	e := As(err)
	if e == ErrInternal {
		slog.ErrorContext(r.Context(), "unhandled error", "method", r.Method, "path", r.URL.Path, "error", err)
	}
	if e.Status == http.StatusUnauthorized {
		w.Header().Set("WWW-Authenticate", `Bearer realm="tabletalk"`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	_ = json.NewEncoder(w).Encode(errorBody{Error: errorPayload{Code: e.Code, Message: e.Message, Details: e.Details}})
}
