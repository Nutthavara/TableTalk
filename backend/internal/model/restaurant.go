// Package model holds the domain types shared by the service and repository layers.
package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TimeOfDay is a wall-clock time as minutes since midnight (0–1439).
type TimeOfDay int

// ParseTimeOfDay accepts "HH:MM" in 24-hour format.
func ParseTimeOfDay(s string) (TimeOfDay, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, fmt.Errorf("time must be in HH:MM format")
	}
	return TimeOfDay(t.Hour()*60 + t.Minute()), nil
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", int(t)/60, int(t)%60)
}

type Restaurant struct {
	ID                      uuid.UUID
	OwnerID                 uuid.UUID
	OwnerName               string
	Name                    string
	Description             string
	Capacity                int
	OpeningTime             TimeOfDay
	ClosingTime             TimeOfDay
	Timezone                string
	CancellationNoticeHours int
	Images                  []Image
	// AverageRating is nil when the restaurant has no reviews.
	AverageRating *float64
	ReviewCount   int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Image struct {
	ID          uuid.UUID
	StorageKey  string
	ContentType string
	Position    int
}
