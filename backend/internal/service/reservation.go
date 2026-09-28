package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/model"
	"tabletalk/backend/internal/repository"
)

var (
	ErrPartySizeExceedsCapacity = apperr.New(http.StatusBadRequest, "PARTY_SIZE_EXCEEDS_CAPACITY",
		"The party is larger than the restaurant's capacity.")
	ErrReservationInPast = apperr.New(http.StatusBadRequest, "RESERVATION_IN_PAST",
		"Reservations must start in the future.")
	ErrOutsideOpeningHours = apperr.New(http.StatusBadRequest, "OUTSIDE_OPENING_HOURS",
		"The whole reservation must be within the restaurant's opening hours.")
	ErrInsufficientCapacity = apperr.New(http.StatusConflict, "INSUFFICIENT_CAPACITY",
		"Not enough seats are available for the selected time.")
)

// ReservationInput uses the restaurant's local date and wall-clock times; the service converts them
// to UTC instants using the restaurant's timezone.
type ReservationInput struct {
	PartySize int
	Date      string // YYYY-MM-DD
	StartTime string // HH:MM
	EndTime   string // HH:MM
}

type ReservationService struct {
	store *repository.Store
	now   func() time.Time
}

func NewReservationService(store *repository.Store, now func() time.Time) *ReservationService {
	return &ReservationService{store: store, now: now}
}

// Create books a table for the authenticated user. The restaurant row lock serialises concurrent
// bookings for the same restaurant, so two requests can never both see the same free seats.
func (s *ReservationService) Create(ctx context.Context, user auth.Identity, restaurantID uuid.UUID, in ReservationInput) (model.Reservation, error) {
	if details := validateReservationInput(in); len(details) > 0 {
		return model.Reservation{}, apperr.Validation(details)
	}

	var created model.Reservation
	err := s.store.InTx(ctx, func(q *repository.Queries) error {
		restaurant, err := q.LockRestaurant(ctx, restaurantID)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRestaurantNotFound
		}
		if err != nil {
			return err
		}

		start, end, err := s.checkBookable(ctx, q, restaurant, in, uuid.Nil)
		if err != nil {
			return err
		}

		created, err = q.InsertReservation(ctx, model.Reservation{
			RestaurantID: restaurant.ID,
			UserID:       user.UserID,
			PartySize:    in.PartySize,
			StartAt:      start,
			EndAt:        end,
		})
		created.RestaurantName, created.RestaurantTimezone = restaurant.Name, restaurant.Timezone
		return err
	})
	return created, err
}

// checkBookable applies every reservation rule against a locked restaurant and returns the UTC
// interval. excludeID is the reservation being edited, which must not count against itself.
func (s *ReservationService) checkBookable(ctx context.Context, q *repository.Queries, r model.Restaurant, in ReservationInput, excludeID uuid.UUID) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation(r.Timezone)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start, end := localInterval(in, loc)

	// Rule 1: party size (> 0 is checked in validateReservationInput).
	if in.PartySize > r.Capacity {
		return time.Time{}, time.Time{}, ErrPartySizeExceedsCapacity
	}
	// Rule 2: start < end (validated earlier) and not in the past, by server time.
	if !start.After(s.now()) {
		return time.Time{}, time.Time{}, ErrReservationInPast
	}
	// Rule 3: entirely within opening hours, in the restaurant's timezone.
	if !withinOpeningHours(start, end, loc, r.OpeningTime, r.ClosingTime) {
		return time.Time{}, time.Time{}, ErrOutsideOpeningHours
	}
	// Rule 4: at every instant of [start, end), seats taken + party size <= capacity.
	overlapping, err := q.ListOverlappingActiveReservations(ctx, r.ID, start, end, excludeID)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if peakOccupancy(overlapping, start, end)+in.PartySize > r.Capacity {
		return time.Time{}, time.Time{}, ErrInsufficientCapacity
	}
	return start.UTC(), end.UTC(), nil
}

// validateReservationInput checks formats and the rules that need no restaurant data.
func validateReservationInput(in ReservationInput) map[string]string {
	details := map[string]string{}
	if in.PartySize < 1 {
		details["party_size"] = "Party size must be at least 1."
	}
	if _, err := time.Parse(time.DateOnly, in.Date); err != nil {
		details["date"] = "Date must be in YYYY-MM-DD format."
	}
	start, errStart := model.ParseTimeOfDay(in.StartTime)
	if errStart != nil {
		details["start_time"] = "Start time must be in HH:MM format."
	}
	end, errEnd := model.ParseTimeOfDay(in.EndTime)
	if errEnd != nil {
		details["end_time"] = "End time must be in HH:MM format."
	}
	if errStart == nil && errEnd == nil && end <= start {
		details["end_time"] = "End time must be after start time."
	}
	return details
}

// localInterval interprets the validated date and times as wall-clock time in loc.
func localInterval(in ReservationInput, loc *time.Location) (time.Time, time.Time) {
	day, _ := time.Parse(time.DateOnly, in.Date)
	startTOD, _ := model.ParseTimeOfDay(in.StartTime)
	endTOD, _ := model.ParseTimeOfDay(in.EndTime)
	at := func(t model.TimeOfDay) time.Time {
		return time.Date(day.Year(), day.Month(), day.Day(), int(t)/60, int(t)%60, 0, 0, loc)
	}
	return at(startTOD), at(endTOD)
}
