# Loads .env (if present) so every target sees the same configuration.
-include .env
export

.PHONY: up down reset-infra migrate migrate-down api test web lint

up: ## Start Postgres and Keycloak
	docker compose up -d

down:
	docker compose down

reset-infra: ## Destroy and recreate Postgres/Keycloak data (dev only)
	docker compose down -v

migrate:
	cd backend && go run ./cmd/migrate up

migrate-down:
	cd backend && go run ./cmd/migrate down

api:
	cd backend && go run ./cmd/api

test: ## Needs Postgres running (make up); uses TEST_DATABASE_URL from .env
	cd backend && go test -race ./...

web:
	cd frontend && npm run dev

lint:
	cd backend && go vet ./...
	cd frontend && npm run lint
