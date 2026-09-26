package service

import (
	"cmp"
	"slices"
	"time"

	"tabletalk/backend/internal/model"
)

type occupancyEvent struct {
	at    time.Time
	delta int
}

// peakOccupancy returns the largest number of seats held at the same instant within [from, to),
// using a sweep line over the reservations' half-open intervals [start_at, end_at).
//
// Summing every overlapping reservation would over-count when those reservations overlap the
// window but not each other; the sweep only adds up parties that are present at the same time.
func peakOccupancy(reservations []model.Reservation, from, to time.Time) int {
	events := make([]occupancyEvent, 0, 2*len(reservations))
	for _, r := range reservations {
		start, end := later(r.StartAt, from), earlier(r.EndAt, to)
		if !start.Before(end) {
			continue // no overlap with the window
		}
		events = append(events, occupancyEvent{start, r.PartySize}, occupancyEvent{end, -r.PartySize})
	}

	// At equal instants, departures are processed before arrivals: [17:00,18:00) and [18:00,19:00)
	// never hold seats at the same time.
	slices.SortFunc(events, func(a, b occupancyEvent) int {
		if c := a.at.Compare(b.at); c != 0 {
			return c
		}
		return cmp.Compare(a.delta, b.delta)
	})

	current, peak := 0, 0
	for _, e := range events {
		current += e.delta
		peak = max(peak, current)
	}
	return peak
}

func later(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func earlier(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
