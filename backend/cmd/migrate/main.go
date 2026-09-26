// Command migrate applies database migrations: go run ./cmd/migrate [up|down|status|reset]
package main

import (
	"context"
	"log"
	"os"

	"tabletalk/backend/internal/database"
)

func main() {
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool, command); err != nil {
		log.Fatalf("migrate %s: %v", command, err)
	}
}
