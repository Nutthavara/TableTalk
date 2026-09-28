package model

import (
	"time"

	"github.com/google/uuid"
)

type ReservationStatus string

const (
	ReservationActive    ReservationStatus = "ACTIVE"
	ReservationCancelled ReservationStatus = "CANCELLED"
)

type Reservation struct {
	ID           uuid.UUID
	RestaurantID uuid.UUID
	UserID       uuid.UUID
	PartySize    int
	StartAt      time.Time
	EndAt        time.Time
	Status       ReservationStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CancelledAt  *time.Time

	// Denormalised from the restaurant for display; filled by queries that join restaurants.
	RestaurantName          string
	RestaurantTimezone      string
	CancellationNoticeHours int
}

// ModificationDeadline is the last instant at which the reservation may be edited or cancelled.
func (r Reservation) ModificationDeadline() time.Time {
	return r.StartAt.Add(-time.Duration(r.CancellationNoticeHours) * time.Hour)
}
