package handler

import (
	"net/http"
	"strings"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/middleware"
)

type Deps struct {
	DB           Pinger
	Auth         *middleware.Authenticator
	Restaurants  *RestaurantHandler
	Reservations *ReservationHandler
	Reviews      *ReviewHandler
	// Uploads serves stored images in development (local storage).
	Uploads http.Handler
}

type AuthMode int

const (
	// Public routes accept a token when present (to personalise results) but never require one.
	Public AuthMode = iota
	Required
)

type Route struct {
	Method  string
	Path    string
	Auth    AuthMode
	handler http.HandlerFunc
}

// Routes is the complete API surface. Tests iterate over it to check that every route that is not
// a read requires authentication.
func Routes(d Deps) []Route {
	rh, resv, rv := d.Restaurants, d.Reservations, d.Reviews
	return []Route{
		{http.MethodGet, "/api/me", Required, Me},

		{http.MethodGet, "/api/restaurants", Public, rh.List},
		{http.MethodGet, "/api/restaurants/{id}", Public, rh.Get},
		{http.MethodPost, "/api/restaurants", Required, rh.Create},
		{http.MethodPut, "/api/restaurants/{id}", Required, rh.Update},
		{http.MethodDelete, "/api/restaurants/{id}", Required, rh.Delete},
		{http.MethodPost, "/api/restaurants/{id}/images", Required, rh.AddImages},
		{http.MethodDelete, "/api/restaurants/{id}/images/{imageId}", Required, rh.DeleteImage},
		{http.MethodGet, "/api/me/restaurants", Required, rh.ListMine},

		{http.MethodPost, "/api/restaurants/{id}/reservations", Required, resv.Create},
		{http.MethodGet, "/api/me/reservations", Required, resv.ListMine},
		{http.MethodGet, "/api/reservations/{id}", Required, resv.Get},
		{http.MethodPut, "/api/reservations/{id}", Required, resv.Update},
		{http.MethodDelete, "/api/reservations/{id}", Required, resv.Cancel},

		{http.MethodGet, "/api/restaurants/{id}/reviews", Public, rv.List},
		{http.MethodPost, "/api/restaurants/{id}/reviews", Required, rv.Create},
		{http.MethodPut, "/api/reviews/{id}", Required, rv.Update},
		{http.MethodDelete, "/api/reviews/{id}", Required, rv.Delete},
	}
}

func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", Health(d.DB))

	for _, rt := range Routes(d) {
		var h http.Handler = rt.handler
		if rt.Auth == Required {
			h = d.Auth.Required(h)
		} else {
			h = d.Auth.Optional(h)
		}
		mux.Handle(rt.Method+" "+rt.Path, h)
	}

	if d.Uploads != nil {
		mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", d.Uploads))
	}

	// Unknown API paths (and unsupported methods on known ones) get the shared error format.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		apperr.Write(w, r, apperr.ErrNotFound)
	})

	return middleware.Recover(middleware.Logger(middleware.SecurityHeaders(mux)))
}

// ExamplePath fills path parameters with a syntactically valid ID, for tests.
func (rt Route) ExamplePath(id string) string {
	p := rt.Path
	for _, param := range []string{"{id}", "{imageId}"} {
		p = strings.ReplaceAll(p, param, id)
	}
	return p
}
