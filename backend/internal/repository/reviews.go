package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"tabletalk/backend/internal/model"
)

// IsUniqueViolation reports whether err is a PostgreSQL unique constraint violation.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

const reviewSelect = `
	SELECT rv.id, rv.restaurant_id, rv.user_id, u.display_name, rv.rating, rv.comment, rv.created_at, rv.updated_at
	FROM reviews rv
	JOIN users u ON u.id = rv.user_id`

func scanReview(row pgx.Row) (model.Review, error) {
	var r model.Review
	err := row.Scan(&r.ID, &r.RestaurantID, &r.UserID, &r.AuthorName, &r.Rating, &r.Comment, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func (q *Queries) GetReview(ctx context.Context, id uuid.UUID) (model.Review, error) {
	r, err := scanReview(q.db.QueryRow(ctx, reviewSelect+` WHERE rv.id = $1`, id))
	return r, notFound(err)
}

// InsertReview fails with a unique violation when the user already reviewed the restaurant.
func (q *Queries) InsertReview(ctx context.Context, r model.Review) (uuid.UUID, error) {
	var id uuid.UUID
	err := q.db.QueryRow(ctx, `
		INSERT INTO reviews (restaurant_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		r.RestaurantID, r.UserID, r.Rating, r.Comment).Scan(&id)
	return id, err
}

func (q *Queries) UpdateReview(ctx context.Context, id uuid.UUID, rating int, comment string) error {
	_, err := q.db.Exec(ctx, `UPDATE reviews SET rating = $2, comment = $3, updated_at = now() WHERE id = $1`, id, rating, comment)
	return err
}

func (q *Queries) DeleteReview(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, `DELETE FROM reviews WHERE id = $1`, id)
	return err
}

// ListReviews returns one page of a restaurant's reviews, newest first, and the total count.
func (q *Queries) ListReviews(ctx context.Context, restaurantID uuid.UUID, limit, offset int) ([]model.Review, int, error) {
	var total int
	if err := q.db.QueryRow(ctx, `SELECT COUNT(*) FROM reviews WHERE restaurant_id = $1`, restaurantID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := q.db.Query(ctx, reviewSelect+`
		WHERE rv.restaurant_id = $1
		ORDER BY rv.created_at DESC, rv.id DESC
		LIMIT $2 OFFSET $3`, restaurantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Review{}
	for rows.Next() {
		r, err := scanReview(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, rows.Err()
}

// ReviewIDByUser returns the user's review of the restaurant, or ErrNotFound.
func (q *Queries) ReviewIDByUser(ctx context.Context, restaurantID, userID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := q.db.QueryRow(ctx, `SELECT id FROM reviews WHERE restaurant_id = $1 AND user_id = $2`, restaurantID, userID).Scan(&id)
	return id, notFound(err)
}

// RestaurantOwner returns the owner of a restaurant, or ErrNotFound.
func (q *Queries) RestaurantOwner(ctx context.Context, restaurantID uuid.UUID) (uuid.UUID, error) {
	var owner uuid.UUID
	err := q.db.QueryRow(ctx, `SELECT owner_id FROM restaurants WHERE id = $1`, restaurantID).Scan(&owner)
	return owner, notFound(err)
}
