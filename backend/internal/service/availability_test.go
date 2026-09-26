package service

import (
	"fmt"
	"testing"
	"time"

	"tabletalk/backend/internal/model"
)

func at(hhmm string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", "2030-01-15 "+hhmm)
	if err != nil {
		panic(err)
	}
	return t
}

func res(from, to string, size int) model.Reservation {
	return model.Reservation{StartAt: at(from), EndAt: at(to), PartySize: size}
}

func TestPeakOccupancy(t *testing.T) {
	tests := []struct {
		name     string
		existing []model.Reservation
		from, to string
		want     int
	}{
		{"no reservations", nil, "17:00", "19:00", 0},
		{"single overlap", []model.Reservation{res("17:00", "18:00", 7)}, "17:30", "18:30", 7},
		{"adjacent before is not counted", []model.Reservation{res("16:00", "17:00", 7)}, "17:00", "18:00", 0},
		{"adjacent after is not counted", []model.Reservation{res("18:00", "19:00", 7)}, "17:00", "18:00", 0},
		{
			// Prompt example: A and B each overlap the window but never each other.
			"disjoint overlaps are not summed",
			[]model.Reservation{res("17:00", "17:30", 6), res("18:00", "18:30", 6)},
			"17:00", "19:00", 6,
		},
		{
			"back-to-back reservations are not summed",
			[]model.Reservation{res("17:00", "18:00", 5), res("18:00", "19:00", 5)},
			"17:00", "19:00", 5,
		},
		{
			"simultaneous reservations are summed",
			[]model.Reservation{res("17:00", "19:00", 3), res("17:30", "18:30", 4), res("18:00", "18:15", 2)},
			"17:00", "19:00", 9,
		},
		{
			"peak outside the window is ignored",
			[]model.Reservation{res("15:00", "16:00", 9), res("15:30", "17:30", 1)},
			"17:00", "18:00", 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := peakOccupancy(tc.existing, at(tc.from), at(tc.to)); got != tc.want {
				t.Fatalf("peakOccupancy = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestWithinOpeningHours(t *testing.T) {
	bangkok, _ := time.LoadLocation("Asia/Bangkok")
	local := func(day int, hhmm string) time.Time {
		t, _ := time.ParseInLocation("2006-01-02 15:04", fmt.Sprintf("2030-01-%02d %s", day, hhmm), bangkok)
		return t
	}
	open, close := model.TimeOfDay(10*60), model.TimeOfDay(22*60)

	tests := []struct {
		name       string
		start, end time.Time
		want       bool
	}{
		{"inside", local(15, "18:00"), local(15, "20:00"), true},
		{"exactly the opening hours", local(15, "10:00"), local(15, "22:00"), true},
		{"starts before opening", local(15, "09:30"), local(15, "11:00"), false},
		{"ends after closing", local(15, "21:00"), local(15, "23:00"), false},
		{"crosses midnight", local(15, "21:00"), local(16, "01:00"), false},
		// 11:00 UTC is 18:00 in Bangkok: the check must use the restaurant's timezone.
		{"evaluated in restaurant timezone", local(15, "18:00").UTC(), local(15, "19:00").UTC(), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := withinOpeningHours(tc.start, tc.end, bangkok, open, close); got != tc.want {
				t.Fatalf("withinOpeningHours = %v, want %v", got, tc.want)
			}
		})
	}
}
