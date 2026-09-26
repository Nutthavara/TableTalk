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
}
