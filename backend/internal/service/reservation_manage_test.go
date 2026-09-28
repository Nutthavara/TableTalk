package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"tabletalk/backend/internal/auth"

	"tabletalk/backend/internal/model"
)

func input(party int, start, end string) ReservationInput {
	return ReservationInput{PartySize: party, Date: testDay, StartTime: start, EndTime: end}
}

// Test #9: editing excludes the reservation itself from the capacity check.
func TestUpdateReservation_ExcludesItself(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	mine := f.mustBook(f.bob, r, 8, "17:00", "18:00")

	// Counting itself would give 8 + 10 = 18 > 10.
	got, err := f.reservations.Update(f.ctx, f.bob, mine.ID, input(10, "17:30", "18:30"))
	if err != nil {
		t.Fatalf("edit into its own seats: %v", err)
	}
	if got.PartySize != 10 || !got.StartAt.Equal(localTime(testDay, "17:30")) || got.ID != mine.ID {
		t.Fatalf("unexpected %+v", got)
	}

	// Other reservations still count: 2 (carol) + 9 > 10.
	f.mustBook(f.carol, r, 2, "19:00", "20:00")
	if _, err := f.reservations.Update(f.ctx, f.bob, mine.ID, input(9, "19:00", "20:00")); errCode(err) != "INSUFFICIENT_CAPACITY" {
		t.Fatalf("edit over another booking: got %v", err)
	}
	if _, err := f.reservations.Update(f.ctx, f.bob, mine.ID, input(8, "19:00", "20:00")); err != nil {
		t.Fatalf("edit that fits: %v", err)
	}
}

func TestUpdateReservation_AppliesCreateRules(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	mine := f.mustBook(f.bob, r, 2, "18:00", "19:00")

	for _, tc := range []struct {
		name string
		in   ReservationInput
		code string
	}{
		{"outside hours", input(2, "21:00", "23:00"), "OUTSIDE_OPENING_HOURS"},
		{"party > capacity", input(11, "18:00", "19:00"), "PARTY_SIZE_EXCEEDS_CAPACITY"},
		{"in the past", ReservationInput{2, "2030-01-14", "18:00", "19:00"}, "RESERVATION_IN_PAST"},
		{"end before start", input(2, "19:00", "18:00"), "VALIDATION_ERROR"},
		{"zero party", input(0, "18:00", "19:00"), "VALIDATION_ERROR"},
	} {
		if _, err := f.reservations.Update(f.ctx, f.bob, mine.ID, tc.in); errCode(err) != tc.code {
			t.Errorf("%s: got %v, want %s", tc.name, err, tc.code)
		}
	}
}

// A3: edits follow the same deadline as cancellation, based on the current start time.
func TestUpdateReservation_Deadline(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	mine := f.mustBook(f.bob, r, 2, "18:00", "19:00")

	f.now = localTime(testDay, "16:01")
	if _, err := f.reservations.Update(f.ctx, f.bob, mine.ID, input(2, "20:00", "21:00")); errCode(err) != "EDIT_DEADLINE_PASSED" {
		t.Fatalf("after deadline: got %v", err)
	}
	f.now = localTime(testDay, "16:00")
	if _, err := f.reservations.Update(f.ctx, f.bob, mine.ID, input(2, "20:00", "21:00")); err != nil {
		t.Fatalf("at deadline: %v", err)
	}
}

// Test #10 and #11: cancel at or before start_at - cancellation_notice_hours.
func TestCancelReservation_Deadline(t *testing.T) {
	for _, tc := range []struct {
		notice int
		now    string
		ok     bool
	}{
		{2, "15:59", true},
		{2, "16:00", true}, // prompt: 18:00 with 2h notice → cancellable up to and including 16:00
		{2, "16:01", false},
		{6, "12:00", true}, // prompt: 6h notice → up to and including 12:00
		{6, "12:01", false},
		{2, "18:30", false}, // already started
	} {
		t.Run(tc.now, func(t *testing.T) {
			f := newFixture(t)
			res := f.mustBook(f.bob, f.restaurantWithNotice(10, tc.notice), 2, "18:00", "20:00")
			f.now = localTime(testDay, tc.now)

			got, err := f.reservations.Cancel(f.ctx, f.bob, res.ID)
			if !tc.ok {
				if errCode(err) != "CANCELLATION_DEADLINE_PASSED" {
					t.Fatalf("notice %dh at %s: got %v, want CANCELLATION_DEADLINE_PASSED", tc.notice, tc.now, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("notice %dh at %s: %v", tc.notice, tc.now, err)
			}
			if got.Status != model.ReservationCancelled || got.CancelledAt == nil || !got.CancelledAt.Equal(f.now) {
				t.Fatalf("not cancelled: %+v", got)
			}
		})
	}
}

func TestCancelReservation_KeepsHistoryAndCannotRepeat(t *testing.T) {
	f := newFixture(t)
	res := f.mustBook(f.bob, f.restaurant(10), 2, "18:00", "20:00")

	if _, err := f.reservations.Cancel(f.ctx, f.bob, res.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.reservations.Cancel(f.ctx, f.bob, res.ID); errCode(err) != "RESERVATION_NOT_ACTIVE" {
		t.Fatalf("second cancel: got %v", err)
	}
	if _, err := f.reservations.Update(f.ctx, f.bob, res.ID, input(2, "18:00", "20:00")); errCode(err) != "RESERVATION_NOT_ACTIVE" {
		t.Fatalf("edit cancelled: got %v", err)
	}
	got, err := f.reservations.Get(f.ctx, f.bob, res.ID)
	if err != nil || got.Status != model.ReservationCancelled {
		t.Fatalf("cancelled reservation must remain readable: %+v %v", got, err)
	}
}

// Test #14: a user cannot see, modify or cancel another user's reservation.
func TestReservation_OtherUsersForbidden(t *testing.T) {
	f := newFixture(t)
	res := f.mustBook(f.bob, f.restaurant(10), 2, "18:00", "20:00")

	if _, err := f.reservations.Get(f.ctx, f.carol, res.ID); !errors.Is(err, ErrNotReservationOwner) {
		t.Errorf("get: got %v", err)
	}
	if _, err := f.reservations.Update(f.ctx, f.carol, res.ID, input(1, "18:00", "19:00")); !errors.Is(err, ErrNotReservationOwner) {
		t.Errorf("update: got %v", err)
	}
	if _, err := f.reservations.Cancel(f.ctx, f.carol, res.ID); !errors.Is(err, ErrNotReservationOwner) {
		t.Errorf("cancel: got %v", err)
	}
	// The restaurant owner is not the reservation owner either.
	if _, err := f.reservations.Cancel(f.ctx, f.alice, res.ID); !errors.Is(err, ErrNotReservationOwner) {
		t.Errorf("restaurant owner cancel: got %v", err)
	}

	got, _ := f.reservations.Get(f.ctx, f.bob, res.ID)
	if got.Status != model.ReservationActive || got.PartySize != 2 {
		t.Fatalf("reservation changed by another user: %+v", got)
	}
	if _, err := f.reservations.Get(f.ctx, f.bob, uuid.New()); !errors.Is(err, ErrReservationNotFound) {
		t.Errorf("unknown id: got %v", err)
	}
}

func TestListMyReservations(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	f.now = localTime("2030-01-14", "09:00")
	past := f.mustBookOn("2030-01-14", f.bob, r, "12:00", "13:00")
	cancelled := f.mustBookOn("2030-01-14", f.bob, r, "18:00", "19:00")
	if _, err := f.reservations.Cancel(f.ctx, f.bob, cancelled.ID); err != nil {
		t.Fatal(err)
	}
	later := f.mustBook(f.bob, r, 2, "19:00", "20:00")
	sooner := f.mustBook(f.bob, r, 2, "12:00", "13:00")
	f.mustBook(f.carol, r, 2, "12:00", "13:00") // someone else's
	f.now = localTime(testDay, "09:00")

	for _, tc := range []struct {
		filter string
		want   []uuid.UUID
	}{
		{"", []uuid.UUID{sooner.ID, later.ID}},
		{"upcoming", []uuid.UUID{sooner.ID, later.ID}},
		{"past", []uuid.UUID{past.ID}},
		{"cancelled", []uuid.UUID{cancelled.ID}},
	} {
		page, err := f.reservations.ListMine(f.ctx, f.bob, ListReservationsParams{Filter: tc.filter, Page: 1, PageSize: 10})
		if err != nil {
			t.Fatal(err)
		}
		var got []uuid.UUID
		for _, r := range page.Items {
			got = append(got, r.ID)
			if r.RestaurantName != "Test" {
				t.Errorf("restaurant name missing: %+v", r)
			}
		}
		if len(got) != len(tc.want) || page.Total != len(tc.want) {
			t.Fatalf("filter %q: got %v (total %d), want %v", tc.filter, got, page.Total, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("filter %q: got %v, want %v", tc.filter, got, tc.want)
			}
		}
	}

	if _, err := f.reservations.ListMine(f.ctx, f.bob, ListReservationsParams{Filter: "all", Page: 1, PageSize: 10}); errCode(err) != "VALIDATION_ERROR" {
		t.Fatalf("unknown filter: got %v", err)
	}
}

// Concurrent edits competing for the last seats must not oversell either: ten 1-person bookings
// all try to become 3 people in a slot with only 3 seats left.
func TestUpdateReservation_ConcurrentEditsCannotExceedCapacity(t *testing.T) {
	f := newFixture(t)
	r := f.restaurant(10)
	f.mustBook(f.carol, r, 7, "17:00", "18:00")
	ids := make([]uuid.UUID, 10)
	for i := range ids {
		ids[i] = f.mustBook(f.bob, r, 1, "12:00", "13:00").ID
	}

	results := runConcurrently(len(ids), func(i int) error {
		_, err := f.reservations.Update(f.ctx, f.bob, ids[i], input(3, "17:00", "18:00"))
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
		t.Fatalf("%d concurrent edits succeeded, want exactly 1", succeeded)
	}
	f.assertNeverOverCapacity(r, 10)
}

func (f *fixture) mustBookOn(date string, user auth.Identity, restaurantID uuid.UUID, start, end string) model.Reservation {
	f.t.Helper()
	r, err := f.reservations.Create(f.ctx, user, restaurantID, ReservationInput{PartySize: 2, Date: date, StartTime: start, EndTime: end})
	if err != nil {
		f.t.Fatalf("book %s %s–%s: %v", date, start, end, err)
	}
	return r
}
