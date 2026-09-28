package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"tabletalk/backend/internal/model"
)

const reservationColumns = `id, restaurant_id, user_id, party_size, start_at, end_at, status, created_at, updated_at, cancelled_at`

func scanReservation(row pgx.Row) (model.Reservation, error) {
	var r model.Reservation
	err := row.Scan(&r.ID, &r.RestaurantID, &r.UserID, &r.PartySize, &r.StartAt, &r.EndAt, &r.Status,
		&r.CreatedAt, &r.UpdatedAt, &r.CancelledAt)
	return r, err
}

func (q *Queries) collectReservations(ctx context.Context, sql string, args ...any) ([]model.Reservation, error) {
	rows, err := q.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Reservation
	for rows.Next() {
		r, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// ListActiveReservationsEndingAfter returns ACTIVE reservations that have not finished by t
// (upcoming and in-progress ones).
func (q *Queries) ListActiveReservationsEndingAfter(ctx context.Context, restaurantID uuid.UUID, t time.Time) ([]model.Reservation, error) {
	return q.collectReservations(ctx, `
		SELECT `+reservationColumns+` FROM reservations
		WHERE restaurant_id = $1 AND status = 'ACTIVE' AND end_at > $2
		ORDER BY start_at`, restaurantID, t)
}

// ListOverlappingActiveReservations returns ACTIVE reservations whose [start_at, end_at) overlaps
// [start, end), excluding excludeID (the reservation being edited; uuid.Nil when creating).
// Call it after LockRestaurant in the same transaction so the result cannot change before insert.
func (q *Queries) ListOverlappingActiveReservations(ctx context.Context, restaurantID uuid.UUID, start, end time.Time, excludeID uuid.UUID) ([]model.Reservation, error) {
	return q.collectReservations(ctx, `
		SELECT `+reservationColumns+` FROM reservations
		WHERE restaurant_id = $1 AND status = 'ACTIVE'
		  AND start_at < $3 AND end_at > $2
		  AND id <> $4`, restaurantID, start, end, excludeID)
}

func (q *Queries) InsertReservation(ctx context.Context, r model.Reservation) (model.Reservation, error) {
	return scanReservation(q.db.QueryRow(ctx, `
		INSERT INTO reservations (restaurant_id, user_id, party_size, start_at, end_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+reservationColumns,
		r.RestaurantID, r.UserID, r.PartySize, r.StartAt, r.EndAt))
}

const reservationWithRestaurantSelect = `
	SELECT r.id, r.restaurant_id, r.user_id, r.party_size, r.start_at, r.end_at, r.status,
	       r.created_at, r.updated_at, r.cancelled_at,
	       rs.name, rs.timezone, rs.cancellation_notice_hours
	FROM reservations r
	JOIN restaurants rs ON rs.id = r.restaurant_id`

func scanReservationWithRestaurant(row pgx.Row) (model.Reservation, error) {
	var r model.Reservation
	err := row.Scan(&r.ID, &r.RestaurantID, &r.UserID, &r.PartySize, &r.StartAt, &r.EndAt, &r.Status,
		&r.CreatedAt, &r.UpdatedAt, &r.CancelledAt,
		&r.RestaurantName, &r.RestaurantTimezone, &r.CancellationNoticeHours)
	return r, err
}

func (q *Queries) GetReservation(ctx context.Context, id uuid.UUID) (model.Reservation, error) {
	r, err := scanReservationWithRestaurant(q.db.QueryRow(ctx, reservationWithRestaurantSelect+` WHERE r.id = $1`, id))
	return r, notFound(err)
}

// LockReservation locks the reservation row. Callers lock the restaurant first (LockRestaurant) so
// every transaction takes locks in the same order.
func (q *Queries) LockReservation(ctx context.Context, id uuid.UUID) (model.Reservation, error) {
	r, err := scanReservation(q.db.QueryRow(ctx, `SELECT `+reservationColumns+` FROM reservations WHERE id = $1 FOR UPDATE`, id))
	return r, notFound(err)
}

func (q *Queries) UpdateReservation(ctx context.Context, id uuid.UUID, partySize int, start, end time.Time) error {
	_, err := q.db.Exec(ctx, `
		UPDATE reservations SET party_size = $2, start_at = $3, end_at = $4, updated_at = now()
		WHERE id = $1`, id, partySize, start, end)
	return err
}

// CancelReservation is a logical cancellation; the row is kept as history.
func (q *Queries) CancelReservation(ctx context.Context, id uuid.UUID, at time.Time) error {
	_, err := q.db.Exec(ctx, `
		UPDATE reservations SET status = 'CANCELLED', cancelled_at = $2, updated_at = now()
		WHERE id = $1`, id, at)
	return err
}

type ReservationFilter string

const (
	FilterUpcoming  ReservationFilter = "upcoming"  // ACTIVE and not finished (includes in progress)
	FilterPast      ReservationFilter = "past"      // ACTIVE and finished
	FilterCancelled ReservationFilter = "cancelled" // CANCELLED
)

// Every condition references $2 (now) so all filters share one parameter list.
var reservationFilters = map[ReservationFilter]struct{ where, orderBy string }{
	FilterUpcoming:  {`r.status = 'ACTIVE' AND r.end_at > $2`, `r.start_at ASC, r.id ASC`},
	FilterPast:      {`r.status = 'ACTIVE' AND r.end_at <= $2`, `r.start_at DESC, r.id ASC`},
	FilterCancelled: {`r.status = 'CANCELLED' AND $2::timestamptz IS NOT NULL`, `r.start_at DESC, r.id ASC`},
}

func ValidReservationFilter(f ReservationFilter) bool {
	_, ok := reservationFilters[f]
	return ok
}

// ListUserReservations returns one page of the user's reservations and the total count.
func (q *Queries) ListUserReservations(ctx context.Context, userID uuid.UUID, filter ReservationFilter, now time.Time, limit, offset int) ([]model.Reservation, int, error) {
	f, ok := reservationFilters[filter]
	if !ok {
		return nil, 0, fmt.Errorf("unknown filter %q", filter)
	}
	where := ` WHERE r.user_id = $1 AND ` + f.where

	var total int
	if err := q.db.QueryRow(ctx, `SELECT COUNT(*) FROM reservations r`+where, userID, now).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := q.db.Query(ctx, reservationWithRestaurantSelect+where+` ORDER BY `+f.orderBy+` LIMIT $3 OFFSET $4`,
		userID, now, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Reservation{}
	for rows.Next() {
		r, err := scanReservationWithRestaurant(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, rows.Err()
}
