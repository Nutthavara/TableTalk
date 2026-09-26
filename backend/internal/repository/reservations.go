package repository

import (
	"context"
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
