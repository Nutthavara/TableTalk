package service

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"testing"

	"github.com/google/uuid"

	"tabletalk/backend/internal/model"
)

// Test #1: a reservation inside opening hours succeeds; local times are interpreted in the
// restaurant's timezone.
func TestCreateReservation_InsideOpeningHours(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)

	got := f.mustBook(f.bob, r, 4, "18:00", "20:00")

	if got.UserID != f.bob.UserID || got.Status != model.ReservationActive || got.PartySize != 4 {
		t.Fatalf("unexpected reservation %+v", got)
	}
	if want := localTime(testDay, "18:00"); !got.StartAt.Equal(want) {
		t.Fatalf("start_at = %v, want %v", got.StartAt, want.UTC())
	}
	if !got.EndAt.Equal(localTime(testDay, "20:00")) {
		t.Fatalf("end_at = %v", got.EndAt)
	}

	// Boundaries: exactly opening to closing is allowed.
	f.mustBook(f.carol, f.restaurant(10), 2, "10:00", "22:00")
}

// Test #2: any part outside opening hours is rejected.
func TestCreateReservation_OutsideOpeningHours(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	for _, tc := range []struct{ start, end string }{
		{"21:00", "23:00"}, // prompt example: ends after closing
		{"09:30", "11:00"}, // starts before opening
		{"22:00", "23:00"}, // starts at closing
		{"23:00", "23:30"},
	} {
		if _, err := f.book(f.bob, r, 2, tc.start, tc.end); errCode(err) != "OUTSIDE_OPENING_HOURS" {
			t.Errorf("%s–%s: got %v, want OUTSIDE_OPENING_HOURS", tc.start, tc.end, err)
		}
	}
}

// Test #3: reservations must start in the future according to the server clock.
func TestCreateReservation_InPast(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	f.now = localTime(testDay, "18:30")

	if _, err := f.book(f.bob, r, 2, "18:00", "20:00"); errCode(err) != "RESERVATION_IN_PAST" {
		t.Fatalf("started 30 minutes ago: got %v", err)
	}
	if _, err := f.book(f.bob, r, 2, "18:30", "20:00"); errCode(err) != "RESERVATION_IN_PAST" {
		t.Fatalf("starting exactly now: got %v", err)
	}
	if _, err := f.reservations.Create(f.ctx, f.bob, r, ReservationInput{PartySize: 2, Date: "2030-01-14", StartTime: "19:00", EndTime: "20:00"}); errCode(err) != "RESERVATION_IN_PAST" {
		t.Fatalf("yesterday: got %v", err)
	}
	f.mustBook(f.bob, r, 2, "18:31", "20:00")
}

// Test #4 plus the other Rule 1/Rule 2 input checks.
func TestCreateReservation_InvalidPartyAndTimes(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)

	if _, err := f.book(f.bob, r, 11, "18:00", "19:00"); errCode(err) != "PARTY_SIZE_EXCEEDS_CAPACITY" {
		t.Errorf("party > capacity: got %v", err)
	}
	f.mustBook(f.bob, r, 10, "18:00", "19:00") // party == capacity is allowed

	for _, tc := range []struct {
		name string
		in   ReservationInput
	}{
		{"zero party", ReservationInput{0, testDay, "18:00", "19:00"}},
		{"negative party", ReservationInput{-2, testDay, "18:00", "19:00"}},
		{"end before start", ReservationInput{2, testDay, "19:00", "18:00"}},
		{"end equals start", ReservationInput{2, testDay, "18:00", "18:00"}},
		{"bad date", ReservationInput{2, "15/01/2030", "18:00", "19:00"}},
		{"bad time", ReservationInput{2, testDay, "6pm", "19:00"}},
	} {
		if _, err := f.reservations.Create(f.ctx, f.bob, r, tc.in); errCode(err) != "VALIDATION_ERROR" {
			t.Errorf("%s: got %v, want VALIDATION_ERROR", tc.name, err)
		}
	}

	if _, err := f.book(f.bob, uuid.New(), 2, "18:00", "19:00"); !errors.Is(err, ErrRestaurantNotFound) {
		t.Errorf("unknown restaurant: got %v", err)
	}
}

// Tests #5–#8: capacity is checked at every instant of the requested interval.
func TestCreateReservation_Capacity(t *testing.T) {
	tests := []struct {
		name     string
		existing [][3]any // party, start, end
		request  [3]any
		wantCode string
	}{
		{
			name:     "#5 overlapping with enough capacity",
			existing: [][3]any{{7, "17:00", "18:00"}},
			request:  [3]any{3, "17:30", "18:30"},
		},
		{
			name:     "#6 overlapping exceeding capacity (prompt: 7 + 5 > 10)",
			existing: [][3]any{{7, "17:00", "18:00"}},
			request:  [3]any{5, "17:30", "18:30"},
			wantCode: "INSUFFICIENT_CAPACITY",
		},
		{
			name:     "#7 adjacent reservations do not overlap",
			existing: [][3]any{{10, "17:00", "18:00"}, {10, "19:00", "20:00"}},
			request:  [3]any{10, "18:00", "19:00"},
		},
		{
			name:     "#8 overlaps that do not overlap each other (prompt: peak 6 + 4 = 10)",
			existing: [][3]any{{6, "17:00", "17:30"}, {6, "18:00", "18:30"}},
			request:  [3]any{4, "17:00", "19:00"},
		},
		{
			name:     "#8 same shape but one seat too many",
			existing: [][3]any{{6, "17:00", "17:30"}, {6, "18:00", "18:30"}},
			request:  [3]any{5, "17:00", "19:00"},
			wantCode: "INSUFFICIENT_CAPACITY",
		},
		{
			name:     "three reservations stacked at one instant",
			existing: [][3]any{{3, "17:00", "19:00"}, {3, "18:00", "20:00"}, {3, "18:30", "18:45"}},
			request:  [3]any{2, "18:40", "21:00"},
			wantCode: "INSUFFICIENT_CAPACITY", // 3+3+3+2 = 11 during 18:40–18:45
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			r := f.restaurant(10)
			for _, e := range tc.existing {
				f.mustBook(f.carol, r, e[0].(int), e[1].(string), e[2].(string))
			}
			_, err := f.book(f.bob, r, tc.request[0].(int), tc.request[1].(string), tc.request[2].(string))
			if errCode(err) != tc.wantCode {
				t.Fatalf("got %q (%v), want %q", errCode(err), err, tc.wantCode)
			}
		})
	}
}

func TestCreateReservation_CancelledReservationsFreeSeats(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	existing := f.mustBook(f.carol, r, 10, "17:00", "18:00")
	if _, err := f.book(f.bob, r, 1, "17:00", "18:00"); errCode(err) != "INSUFFICIENT_CAPACITY" {
		t.Fatalf("full: got %v", err)
	}
	f.exec(`UPDATE reservations SET status = 'CANCELLED', cancelled_at = now() WHERE id = $1`, existing.ID)
	f.mustBook(f.bob, r, 10, "17:00", "18:00")
}

// Test #16: concurrent requests cannot oversell. With 3 seats left, many simultaneous requests for
// 3 seats must produce exactly one booking.
func TestCreateReservation_ConcurrentRequestsCannotExceedCapacity(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	f.mustBook(f.carol, r, 7, "17:00", "18:00")

	const workers = 20
	results := runConcurrently(workers, func(int) error {
		_, err := f.book(f.bob, r, 3, "17:00", "18:00")
		return err
	})

	succeeded := 0
	for _, err := range results {
		switch errCode(err) {
		case "":
			succeeded++
		case "INSUFFICIENT_CAPACITY":
		default:
			t.Errorf("unexpected error: %v", err)
		}
	}
	if succeeded != 1 {
		t.Fatalf("%d of %d concurrent bookings succeeded, want exactly 1", succeeded, workers)
	}
	f.assertNeverOverCapacity(r, 10)
}

// Random overlapping bookings raced against each other must never exceed capacity at any instant.
func TestCreateReservation_ConcurrentRandomBookingsStayWithinCapacity(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	rng := rand.New(rand.NewPCG(1, 2))
	type req struct {
		party      int
		start, end string
	}
	requests := make([]req, 60)
	for i := range requests {
		startSlot := rng.IntN(8) // 17:00 .. 20:30 in 30-minute slots
		length := 1 + rng.IntN(4)
		slot := func(n int) string { return fmt.Sprintf("%02d:%02d", 17+n/2, 30*(n%2)) }
		requests[i] = req{1 + rng.IntN(4), slot(startSlot), slot(min(startSlot+length, 10))}
	}

	results := runConcurrently(len(requests), func(i int) error {
		_, err := f.book(f.bob, r, requests[i].party, requests[i].start, requests[i].end)
		return err
	})
	accepted := 0
	for _, err := range results {
		if code := errCode(err); code != "" && code != "INSUFFICIENT_CAPACITY" {
			t.Fatalf("unexpected error: %v", err)
		} else if code == "" {
			accepted++
		}
	}
	if accepted == 0 || accepted == len(requests) {
		t.Fatalf("expected a mix of accepted and rejected bookings, got %d accepted", accepted)
	}
	f.assertNeverOverCapacity(r, 10)
}

// runConcurrently starts n goroutines that call fn at the same moment and returns their errors.
func runConcurrently(n int, fn func(i int) error) []error {
	var (
		wg      sync.WaitGroup
		start   = make(chan struct{})
		results = make([]error, n)
	)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results[i] = fn(i)
		}()
	}
	close(start)
	wg.Wait()
	return results
}

func (f *fixture) assertNeverOverCapacity(restaurantID uuid.UUID, capacity int) {
	f.t.Helper()
	all, err := f.store.ListActiveReservationsEndingAfter(f.ctx, restaurantID, f.now)
	if err != nil {
		f.t.Fatal(err)
	}
	if peak := peakOccupancy(all, localTime(testDay, "00:00"), localTime(testDay, "23:59")); peak > capacity {
		f.t.Fatalf("stored reservations reach %d seats at once, capacity is %d", peak, capacity)
	}
}
