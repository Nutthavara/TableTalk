package repository

import (
	"context"

	"tabletalk/backend/internal/auth"
)

// SyncUser upserts the profile from a verified token; the row is only rewritten when a field changed.
func (q *Queries) SyncUser(ctx context.Context, id auth.Identity) error {
	_, err := q.db.Exec(ctx, `
		INSERT INTO users (id, username, display_name, email)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		ON CONFLICT (id) DO UPDATE
		SET username = EXCLUDED.username,
		    display_name = EXCLUDED.display_name,
		    email = EXCLUDED.email,
		    updated_at = now()
		WHERE (users.username, users.display_name, users.email)
		      IS DISTINCT FROM (EXCLUDED.username, EXCLUDED.display_name, EXCLUDED.email)`,
		id.UserID, id.Username, id.DisplayName, id.Email)
	return err
}
