package handler

import (
	"net/http"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/middleware"
)

type Deps struct {
	DB           Pinger
	Auth         *middleware.Authenticator
	Restaurants  *RestaurantHandler
	Reservations *ReservationHandler
	// Uploads serves stored images in development (local storage).
	Uploads http.Handler
}

func NewRouter(d Deps) http.Handler {
	mux := http.NewServeMux()
	required := func(h http.HandlerFunc) http.Handler { return d.Auth.Required(h) }
	optional := func(h http.HandlerFunc) http.Handler { return d.Auth.Optional(h) }

	mux.HandleFunc("GET /healthz", Health(d.DB))
	mux.Handle("GET /api/me", required(Me))

	rh := d.Restaurants
	mux.Handle("GET /api/restaurants", optional(rh.List))
	mux.Handle("GET /api/restaurants/{id}", optional(rh.Get))
	mux.Handle("POST /api/restaurants", required(rh.Create))
	mux.Handle("PUT /api/restaurants/{id}", required(rh.Update))
	mux.Handle("DELETE /api/restaurants/{id}", required(rh.Delete))
	mux.Handle("POST /api/restaurants/{id}/images", required(rh.AddImages))
	mux.Handle("DELETE /api/restaurants/{id}/images/{imageId}", required(rh.DeleteImage))
	mux.Handle("GET /api/me/restaurants", required(rh.ListMine))

	mux.Handle("POST /api/restaurants/{id}/reservations", required(d.Reservations.Create))

	if d.Uploads != nil {
		mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", d.Uploads))
	}

	// Unknown API paths get the shared error format instead of the default text 404.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		apperr.Write(w, r, apperr.ErrNotFound)
	})

	return middleware.Recover(middleware.Logger(mux))
}
