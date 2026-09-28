package model

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID           uuid.UUID
	RestaurantID uuid.UUID
	UserID       uuid.UUID
	AuthorName   string
	Rating       int
	Comment      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
