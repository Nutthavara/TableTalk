package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"tabletalk/backend/internal/apperr"
	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/model"
	"tabletalk/backend/internal/repository"
	"tabletalk/backend/internal/testutil"
)

// testDay is the fixed "today" for service tests; the clock reads 09:00 Bangkok on this day.
const testDay = "2030-01-15"

var bangkok = mustLoad("Asia/Bangkok")

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

type fixture struct {
	t            *testing.T
	ctx          context.Context
	db           *pgxpool.Pool
	store        *repository.Store
	now          time.Time
	reservations *ReservationService
	reviews      *ReviewService
	alice, bob   auth.Identity
	carol        auth.Identity
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db := testutil.DB(t, "service")
	f := &fixture{
		t:     t,
		ctx:   context.Background(),
		db:    db,
		store: repository.NewStore(db),
		now:   localTime(testDay, "09:00"),
		alice: auth.Identity{UserID: uuid.New(), Username: "alice", DisplayName: "Alice"},
		bob:   auth.Identity{UserID: uuid.New(), Username: "bob", DisplayName: "Bob"},
		carol: auth.Identity{UserID: uuid.New(), Username: "carol", DisplayName: "Carol"},
	}
	clock := func() time.Time { return f.now }
	f.reservations = NewReservationService(f.store, clock)
	f.reviews = NewReviewService(f.store)
	for _, u := range []auth.Identity{f.alice, f.bob, f.carol} {
		if err := f.store.SyncUser(f.ctx, u); err != nil {
			t.Fatal(err)
		}
	}
	return f
}

func localTime(date, hhmm string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", date+" "+hhmm, bangkok)
	if err != nil {
		panic(err)
	}
	return t
}

// restaurant inserts a restaurant owned by alice, open 10:00–22:00 Bangkok, 2h cancellation notice.
func (f *fixture) restaurant(capacity int) uuid.UUID {
	f.t.Helper()
	return f.restaurantWithNotice(capacity, 2)
}

func (f *fixture) restaurantWithNotice(capacity, noticeHours int) uuid.UUID {
	f.t.Helper()
	id, err := f.store.InsertRestaurant(f.ctx, model.Restaurant{
		OwnerID: f.alice.UserID, Name: "Test", Capacity: capacity,
		OpeningTime: 10 * 60, ClosingTime: 22 * 60, Timezone: "Asia/Bangkok", CancellationNoticeHours: noticeHours,
	})
	if err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *fixture) book(user auth.Identity, restaurantID uuid.UUID, party int, start, end string) (model.Reservation, error) {
	return f.reservations.Create(f.ctx, user, restaurantID, ReservationInput{
		PartySize: party, Date: testDay, StartTime: start, EndTime: end,
	})
}

// mustBook fails the test unless the booking succeeds.
func (f *fixture) mustBook(user auth.Identity, restaurantID uuid.UUID, party int, start, end string) model.Reservation {
	f.t.Helper()
	r, err := f.book(user, restaurantID, party, start, end)
	if err != nil {
		f.t.Fatalf("book %d people %s–%s: %v", party, start, end, err)
	}
	return r
}

func (f *fixture) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.db.Exec(f.ctx, sql, args...); err != nil {
		f.t.Fatalf("exec %q: %v", sql, err)
	}
}

// errCode returns the application error code, or "" for nil.
func errCode(err error) string {
	if err == nil {
		return ""
	}
	var e *apperr.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return "unexpected: " + err.Error()
}
