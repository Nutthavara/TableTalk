package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"tabletalk/backend/internal/model"
)

// restaurantSelect includes the rating aggregate computed from reviews (not stored on restaurants).
const restaurantSelect = `
	SELECT r.id, r.owner_id, u.display_name, r.name, r.description, r.capacity,
	       r.opening_time, r.closing_time, r.timezone, r.cancellation_notice_hours,
	       r.created_at, r.updated_at,
	       ROUND(s.avg_rating, 1)::float8, s.review_count
	FROM restaurants r
	JOIN users u ON u.id = r.owner_id
	LEFT JOIN LATERAL (
	    SELECT AVG(rating) AS avg_rating, COUNT(*) AS review_count
	    FROM reviews WHERE restaurant_id = r.id
	) s ON true`

func scanRestaurant(row pgx.Row) (model.Restaurant, error) {
	var (
		r           model.Restaurant
		open, close pgtype.Time
	)
	err := row.Scan(&r.ID, &r.OwnerID, &r.OwnerName, &r.Name, &r.Description, &r.Capacity,
		&open, &close, &r.Timezone, &r.CancellationNoticeHours,
		&r.CreatedAt, &r.UpdatedAt, &r.AverageRating, &r.ReviewCount)
	r.OpeningTime, r.ClosingTime = fromPgTime(open), fromPgTime(close)
	return r, err
}

func fromPgTime(t pgtype.Time) model.TimeOfDay {
	return model.TimeOfDay(t.Microseconds / 60_000_000)
}

func toPgTime(t model.TimeOfDay) pgtype.Time {
	return pgtype.Time{Microseconds: int64(t) * 60_000_000, Valid: true}
}

func (q *Queries) GetRestaurant(ctx context.Context, id uuid.UUID) (model.Restaurant, error) {
	r, err := scanRestaurant(q.db.QueryRow(ctx, restaurantSelect+` WHERE r.id = $1`, id))
	if err != nil {
		return model.Restaurant{}, notFound(err)
	}
	if err := q.attachImages(ctx, []*model.Restaurant{&r}); err != nil {
		return model.Restaurant{}, err
	}
	return r, nil
}

func (q *Queries) ListRestaurantsByOwner(ctx context.Context, ownerID uuid.UUID) ([]model.Restaurant, error) {
	return q.listRestaurants(ctx, restaurantSelect+` WHERE r.owner_id = $1 ORDER BY r.created_at DESC, r.id`, ownerID)
}

func (q *Queries) listRestaurants(ctx context.Context, sql string, args ...any) ([]model.Restaurant, error) {
	rows, err := q.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Restaurant
	for rows.Next() {
		r, err := scanRestaurant(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ptrs := make([]*model.Restaurant, len(list))
	for i := range list {
		ptrs[i] = &list[i]
	}
	return list, q.attachImages(ctx, ptrs)
}

// LockRestaurant takes a row lock that serialises every write affecting the restaurant
// (restaurant edits and reservation create/edit/cancel). Must be called inside a transaction.
// The returned value has no owner name, rating or images.
func (q *Queries) LockRestaurant(ctx context.Context, id uuid.UUID) (model.Restaurant, error) {
	var (
		r           model.Restaurant
		open, close pgtype.Time
	)
	err := q.db.QueryRow(ctx, `
		SELECT id, owner_id, name, description, capacity, opening_time, closing_time, timezone,
		       cancellation_notice_hours, created_at, updated_at
		FROM restaurants WHERE id = $1
		FOR UPDATE`, id).
		Scan(&r.ID, &r.OwnerID, &r.Name, &r.Description, &r.Capacity, &open, &close, &r.Timezone,
			&r.CancellationNoticeHours, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return model.Restaurant{}, notFound(err)
	}
	r.OpeningTime, r.ClosingTime = fromPgTime(open), fromPgTime(close)
	return r, nil
}

func (q *Queries) InsertRestaurant(ctx context.Context, r model.Restaurant) (uuid.UUID, error) {
	var id uuid.UUID
	err := q.db.QueryRow(ctx, `
		INSERT INTO restaurants (owner_id, name, description, capacity, opening_time, closing_time,
		                         timezone, cancellation_notice_hours)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		r.OwnerID, r.Name, r.Description, r.Capacity, toPgTime(r.OpeningTime), toPgTime(r.ClosingTime),
		r.Timezone, r.CancellationNoticeHours).Scan(&id)
	return id, err
}

// UpdateRestaurant replaces the editable fields. owner_id is never updated.
func (q *Queries) UpdateRestaurant(ctx context.Context, r model.Restaurant) error {
	tag, err := q.db.Exec(ctx, `
		UPDATE restaurants
		SET name = $2, description = $3, capacity = $4, opening_time = $5, closing_time = $6,
		    timezone = $7, cancellation_notice_hours = $8, updated_at = now()
		WHERE id = $1`,
		r.ID, r.Name, r.Description, r.Capacity, toPgTime(r.OpeningTime), toPgTime(r.ClosingTime),
		r.Timezone, r.CancellationNoticeHours)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (q *Queries) DeleteRestaurant(ctx context.Context, id uuid.UUID) error {
	tag, err := q.db.Exec(ctx, `DELETE FROM restaurants WHERE id = $1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

func (q *Queries) attachImages(ctx context.Context, restaurants []*model.Restaurant) error {
	if len(restaurants) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(restaurants))
	byID := make(map[uuid.UUID]*model.Restaurant, len(restaurants))
	for i, r := range restaurants {
		ids[i] = r.ID
		byID[r.ID] = r
		r.Images = []model.Image{}
	}

	rows, err := q.db.Query(ctx, `
		SELECT restaurant_id, id, storage_key, content_type, position
		FROM restaurant_images WHERE restaurant_id = ANY($1)
		ORDER BY position`, ids)
	if err != nil {
		return fmt.Errorf("list images: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			restaurantID uuid.UUID
			img          model.Image
		)
		if err := rows.Scan(&restaurantID, &img.ID, &img.StorageKey, &img.ContentType, &img.Position); err != nil {
			return err
		}
		byID[restaurantID].Images = append(byID[restaurantID].Images, img)
	}
	return rows.Err()
}

func (q *Queries) ListImages(ctx context.Context, restaurantID uuid.UUID) ([]model.Image, error) {
	r := &model.Restaurant{ID: restaurantID}
	err := q.attachImages(ctx, []*model.Restaurant{r})
	return r.Images, err
}

func (q *Queries) InsertImage(ctx context.Context, restaurantID uuid.UUID, img model.Image) error {
	_, err := q.db.Exec(ctx, `
		INSERT INTO restaurant_images (restaurant_id, storage_key, content_type, position)
		VALUES ($1, $2, $3, $4)`, restaurantID, img.StorageKey, img.ContentType, img.Position)
	return err
}

func (q *Queries) DeleteImage(ctx context.Context, restaurantID, imageID uuid.UUID) error {
	tag, err := q.db.Exec(ctx, `DELETE FROM restaurant_images WHERE id = $1 AND restaurant_id = $2`, imageID, restaurantID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return err
}

type RestaurantSort string

const (
	SortHighestRating RestaurantSort = "rating"
	SortMostReviewed  RestaurantSort = "reviews"
)

// Deterministic orderings; restaurants without reviews sort after rated ones. Ratings are compared
// at the displayed precision (1 decimal) so equal-looking ratings fall through to the tie-breakers.
var restaurantOrderBy = map[RestaurantSort]string{
	SortHighestRating: `ROUND(s.avg_rating, 1) DESC NULLS LAST, s.review_count DESC, r.id ASC`,
	SortMostReviewed:  `s.review_count DESC, ROUND(s.avg_rating, 1) DESC NULLS LAST, r.id ASC`,
}

// ListRestaurants returns one page of restaurants. A non-empty search matches names containing it,
// ignoring case; LIKE wildcards in the search are matched literally.
func (q *Queries) ListRestaurants(ctx context.Context, sort RestaurantSort, search string, limit, offset int) ([]model.Restaurant, error) {
	orderBy, ok := restaurantOrderBy[sort]
	if !ok {
		return nil, fmt.Errorf("unknown sort %q", sort)
	}
	return q.listRestaurants(ctx, restaurantSelect+` WHERE `+nameMatches+` ORDER BY `+orderBy+` LIMIT $2 OFFSET $3`,
		likePattern(search), limit, offset)
}

func (q *Queries) CountRestaurants(ctx context.Context, search string) (int, error) {
	var n int
	err := q.db.QueryRow(ctx, `SELECT COUNT(*) FROM restaurants r WHERE `+nameMatches, likePattern(search)).Scan(&n)
	return n, err
}

// nameMatches filters by $1, a pattern from likePattern ('%' matches every restaurant).
const nameMatches = `r.name ILIKE $1 ESCAPE '\'`

func likePattern(search string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(search)
	return "%" + escaped + "%"
}
