// Package service holds the business rules. Handlers only translate HTTP to service calls; the
// authenticated identity is always passed in explicitly and never taken from request input.
package service

import (
	"net/http"

	"tabletalk/backend/internal/apperr"
)

var (
	ErrRestaurantNotFound    = apperr.New(http.StatusNotFound, "RESTAURANT_NOT_FOUND", "Restaurant not found.")
	ErrImageNotFound         = apperr.New(http.StatusNotFound, "IMAGE_NOT_FOUND", "Image not found.")
	ErrNotRestaurantOwner    = apperr.New(http.StatusForbidden, "FORBIDDEN", "Only the restaurant owner can do this.")
	ErrLastImage             = apperr.New(http.StatusConflict, "LAST_IMAGE", "A restaurant must keep at least one image.")
	ErrTooManyImages         = apperr.New(http.StatusConflict, "TOO_MANY_IMAGES", "A restaurant can have at most 10 images.")
	ErrHasActiveReservations = apperr.New(http.StatusConflict, "RESTAURANT_HAS_UPCOMING_RESERVATIONS",
		"This restaurant still has upcoming reservations and cannot be deleted.")
	ErrConflictsWithReservations = apperr.New(http.StatusConflict, "CONFLICTS_WITH_EXISTING_RESERVATIONS",
		"This change conflicts with existing upcoming reservations (capacity or opening hours).")
)
