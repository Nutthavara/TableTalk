// Package testutil provides a real PostgreSQL database for tests.
package testutil

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tabletalk/backend/internal/database"
)

var (
	pools   = map[string]*pgxpool.Pool{}
	poolsMu sync.Mutex
)

// DB returns a pool whose search_path is a schema dedicated to one test package (so packages can run
// in parallel), migrated to the latest version, with all tables emptied. Tests are skipped when
// TEST_DATABASE_URL is not set.
func DB(t testing.TB, schema string) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set; run `make test` or see README")
	}

	poolsMu.Lock()
	defer poolsMu.Unlock()
	pool, ok := pools[schema]
	if !ok {
		pool = newSchemaPool(t, url, "test_"+schema)
		pools[schema] = pool
	}
	truncate(t, pool)
	return pool
}

func newSchemaPool(t testing.TB, url, schema string) *pgxpool.Pool {
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	defer admin.Close(ctx)
	ident := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "DROP SCHEMA IF EXISTS "+ident+" CASCADE; CREATE SCHEMA "+ident); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(ctx, pool, "up"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

func truncate(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`TRUNCATE reviews, reservations, restaurant_images, restaurants, users CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}
