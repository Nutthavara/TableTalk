package handler

import (
	"net/http"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/middleware"
)

type Deps struct {
	DB   Pinger
	Auth *middleware.Authenticator
}

func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()
	required := func(h http.HandlerFunc) http.Handler { return d.Auth.Required(h) }

	mux.HandleFunc("GET /healthz", Health(d.DB))
	mux.Handle("GET /api/me", required(Me))

	// Unknown API paths get the shared error format instead of the default text 404.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		apperr.Write(w, r, apperr.ErrNotFound)
	})

	return middleware.Recover(middleware.Logger(mux))
}
