# TableTalk — Restaurant Reservation & Review System

Anyone who signs in can open their own restaurants, book tables at other people's restaurants and review them. A single account can be both an owner and a customer. You own a restaurant when its `owner_id` matches your user ID. There are no roles.

```
Browser ──► Next.js 16 (App Router, server-side BFF) ──Bearer JWT──► Go REST API ──► PostgreSQL 18
               │ ▲                                                      │
               ▼ │ OIDC Authorization Code + PKCE                       │ verifies JWT via JWKS
             Keycloak 26 ◄──────────────────────────────────────────────┘
```

- **The Go API is the only authority** for authorization and business rules. Frontend checks exist only to make the UI easier to use.
- **Access tokens stay on the Next.js server** in an encrypted, httpOnly session cookie. Browser JavaScript never receives them.
- **Docker Compose runs Postgres and Keycloak.** The API and the web app run on your machine, so debugging and hot reload work normally.

> **Status:** backend phases 1–7 and 9 are complete. The Next.js UI (phase 8) is waiting for the final design. Authentication (login and logout) already works end to end.

---

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Docker + Compose v2 | recent | Colima or Docker Desktop both work |
| Go | 1.26.5+ | `go.mod` pins `toolchain go1.26.6`, and the `go` command downloads it automatically |
| Node.js | 20.9+ | Needed by Next.js 16 |
| make | any | Optional. Each target is a single command you can run by hand |

Ports used by default: **5432** (Postgres), **8080** (Keycloak), **8081** (Go API), **3000** (Next.js). Postgres and Keycloak ports can be changed in `.env`.

## Quick start

```bash
cp .env.example .env                          # root config: compose + Go API
cp frontend/.env.example frontend/.env.local  # frontend config
# Put a random value in AUTH_SECRET in frontend/.env.local:
#   openssl rand -base64 33

make up          # start Postgres + Keycloak (Keycloak needs ~20–40 s on first start)
make migrate     # apply database migrations
make api         # terminal 1: Go API on http://localhost:8081
cd frontend && npm install && cd ..
make web         # terminal 2: Next.js on http://localhost:3000
```

Open http://localhost:3000 and sign in as **alice / alice** or **bob / bob**. These are development users only. You can also register a new account on the Keycloak login page.

## Environment variables

### Root `.env` (Docker Compose and Go API)

The Makefile loads this file, so every `make` target sees these values.

| Variable | Default | Used by | Purpose |
|---|---|---|---|
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | `tabletalk` | compose | Database credentials (dev only) |
| `POSTGRES_PORT` | `5432` | compose | Host port for Postgres |
| `KEYCLOAK_PORT` | `8080` | compose | Host port for Keycloak. Also used in the token issuer URL |
| `KEYCLOAK_ADMIN` / `KEYCLOAK_ADMIN_PASSWORD` | `admin` | compose | Keycloak admin console login (dev only) |
| `HTTP_ADDR` | `:8081` | API | Listen address |
| `DATABASE_URL` | — | API, migrate | Postgres connection string (**required**) |
| `TEST_DATABASE_URL` | — | tests | Database for backend tests. Tests are skipped when it is unset |
| `OIDC_ISSUER_URL` | — | API | `http://localhost:8080/realms/tabletalk` (**required**). Must match the token's `iss` exactly |
| `OIDC_AUDIENCE` | `tabletalk-api` | API | Required `aud` in access tokens |
| `UPLOAD_DIR` | `./uploads` | API | Where images are stored (relative to `backend/`) |
| `PUBLIC_BASE_URL` | `http://localhost:8081` | API | Base URL used to build image URLs |

If you change `KEYCLOAK_PORT`, also update `OIDC_ISSUER_URL` here and `AUTH_KEYCLOAK_ISSUER` in the frontend.

### `frontend/.env.local`

| Variable | Example | Purpose |
|---|---|---|
| `AUTH_SECRET` | random 32+ bytes | Encrypts the session cookie |
| `AUTH_URL` | `http://localhost:3000` | Public URL of the web app |
| `AUTH_KEYCLOAK_ID` | `tabletalk-web` | OIDC client ID |
| `AUTH_KEYCLOAK_SECRET` | `dev-only-tabletalk-web-secret` | OIDC client secret. Matches the realm import and is **for development only** |
| `AUTH_KEYCLOAK_ISSUER` | `http://localhost:8080/realms/tabletalk` | Keycloak realm URL |
| `API_BASE_URL` | `http://localhost:8081` | Go API. Only the Next.js server calls it |

`.env` and `.env.local` are git-ignored. Never commit real secrets. Every secret in this repo is a development placeholder.

## Keycloak setup

No manual setup is needed. On first start, Compose imports [`keycloak/realm-tabletalk.json`](keycloak/realm-tabletalk.json), which contains:

- **Realm `tabletalk`**, with self-registration enabled and no application roles.
- **Client `tabletalk-web`**, used by Next.js. It is a confidential client using the authorization code flow with PKCE (S256), with redirect URI `http://localhost:3000/*`.
- **Client `tabletalk-dev-cli`**, a public client that allows the password grant. It exists so you can get tokens with `curl` during development. Do **not** keep it in production.
- **An audience mapper on both clients**, which adds `aud: tabletalk-api` to access tokens. The API rejects any token without it.
- **Users `alice` and `bob`**, whose passwords equal their usernames.

Admin console: http://localhost:8080/admin (`admin` / `admin`).

Keycloak imports the realm only when it does not exist yet. To apply changes you made to the JSON, recreate the containers. This also **deletes all app data**:

```bash
make reset-infra && make up && make migrate
```

### Calling the API with curl

```bash
TOKEN=$(curl -s -d grant_type=password -d client_id=tabletalk-dev-cli \
  -d username=alice -d password=alice \
  http://localhost:8080/realms/tabletalk/protocol/openid-connect/token | jq -r .access_token)

curl -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/me
```

### How authentication works

1. **Sign in:** Next.js (Auth.js v5) redirects to Keycloak using the authorization code flow with PKCE and state. It exchanges the code on the server and stores the tokens in an encrypted, chunked httpOnly cookie. `/api/auth/session` exposes no tokens.
2. **Calling the API:** Server Components and Server Actions call the Go API with `Authorization: Bearer <access token>`. Tokens are refreshed shortly before they expire.
3. **Verifying the token:** The Go API checks the RS256 signature against Keycloak's JWKS, then checks `iss`, `aud = tabletalk-api` and `exp`. It reads the user ID from `sub` and upserts the `users` profile row.
4. **Identity:** `owner_id` and `user_id` always come from the verified token. JSON bodies reject unknown fields, so a client cannot send identity fields.
5. **Sign out:** clears the session and ends the Keycloak SSO session.

## Database and migrations

The database is PostgreSQL 18 and the SQL migrations live in [`backend/migrations`](backend/migrations). They are embedded in the binary and run with [goose](https://github.com/pressly/goose), so you do not need to install goose.

```bash
make migrate                                   # apply all pending migrations
make migrate-down                              # roll back the latest migration
cd backend && go run ./cmd/migrate status      # show applied migrations
```

To add a migration, create `backend/migrations/0000N_description.sql` with `-- +goose Up` and `-- +goose Down` sections.

All timestamps are stored as `timestamptz` and returned by the API in UTC. Opening hours are local wall-clock times in the restaurant's IANA `timezone` (default `Asia/Bangkok`).

## Running

```bash
make api    # Go API    → http://localhost:8081 (GET /healthz checks DB connectivity)
make web    # Next.js   → http://localhost:3000
```

Uploaded images are written to `backend/uploads/` and served by the API at `/uploads/...`. Storage sits behind the `storage.Storage` interface, so you can replace the local implementation with S3, MinIO or GCS without changing the service code.

## Tests

The backend tests need Postgres running (`make up`). They use the `tabletalk_test` database, which Compose creates on first start. Each test package migrates its own schema, so no manual setup is needed.

```bash
make test                                   # go test -race ./... with TEST_DATABASE_URL from .env
cd backend && go test ./internal/service/   # a single package (export TEST_DATABASE_URL first)
make lint                                   # go vet + eslint
```

If `TEST_DATABASE_URL` is unset, tests that need the database are **skipped** rather than failing. Use `make test` so they always run.

| Required test (spec §20) | Location |
|---|---|
| 1–8 reservation rules (opening hours, past, capacity, adjacency, sweep line) | `backend/internal/service/reservation_test.go` |
| 9–11 edit excludes itself; cancel before/after the deadline | `backend/internal/service/reservation_manage_test.go` |
| 12, 15 owner cannot review own restaurant; rating must be 1–5 | `backend/internal/service/review_test.go`, `backend/internal/handler/reviews_test.go` |
| 13 cannot modify another owner's restaurant | `backend/internal/handler/restaurants_test.go` |
| 14 cannot modify another user's reservation | `backend/internal/service/reservation_manage_test.go`, `backend/internal/handler/reservations_test.go` |
| 16 concurrent bookings and edits cannot exceed capacity | `backend/internal/service/reservation_test.go`, `reservation_manage_test.go` |
| Every write route requires auth; forged identity fields are rejected | `backend/internal/handler/security_test.go` |

---

## API reference

Every error uses one format:

```json
{ "error": { "code": "INSUFFICIENT_CAPACITY", "message": "Not enough seats are available for the selected time.", "details": { "field": "message" } } }
```

`details` is present only for `VALIDATION_ERROR`.

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/me` | required | Current user |
| GET | `/api/restaurants?sort=rating\|reviews&page=&page_size=` | optional | List restaurants (page size 1–50, default 12) |
| GET | `/api/restaurants/{id}` | optional | Detail, with `is_owner` and `my_review_id` for signed-in users |
| POST | `/api/restaurants` | required | Create (`multipart/form-data` with fields and one or more `images`) |
| PUT | `/api/restaurants/{id}` | owner | Replace the editable fields (JSON, all fields required) |
| DELETE | `/api/restaurants/{id}` | owner | Delete (refused while unfinished ACTIVE reservations exist) |
| POST | `/api/restaurants/{id}/images` | owner | Add images (multipart `images`) |
| DELETE | `/api/restaurants/{id}/images/{imageId}` | owner | Remove an image (the last image cannot be removed) |
| GET | `/api/me/restaurants` | required | Restaurants you own |
| POST | `/api/restaurants/{id}/reservations` | required | Book: `{party_size, date: "YYYY-MM-DD", start_time: "HH:MM", end_time: "HH:MM"}` in the restaurant's local time |
| GET | `/api/me/reservations?filter=upcoming\|past\|cancelled&page=&page_size=` | required | Your reservations |
| GET | `/api/reservations/{id}` | booker | Reservation detail |
| PUT | `/api/reservations/{id}` | booker | Edit (same body and rules as create) |
| DELETE | `/api/reservations/{id}` | booker | **Cancels** the reservation (logical; the row is kept) and returns it |
| GET | `/api/restaurants/{id}/reviews?page=&page_size=` | optional | Reviews, newest first |
| POST | `/api/restaurants/{id}/reviews` | required, not owner | `{rating: 1–5, comment?}` |
| PUT | `/api/reviews/{id}` | author | Edit review |
| DELETE | `/api/reviews/{id}` | author | Delete review |

| Status | Codes |
|---|---|
| 400 | `VALIDATION_ERROR`, `PARTY_SIZE_EXCEEDS_CAPACITY`, `RESERVATION_IN_PAST`, `OUTSIDE_OPENING_HOURS` |
| 401 | `UNAUTHORIZED` |
| 403 | `FORBIDDEN`, `CANNOT_REVIEW_OWN_RESTAURANT` |
| 404 | `NOT_FOUND`, `RESTAURANT_NOT_FOUND`, `IMAGE_NOT_FOUND`, `RESERVATION_NOT_FOUND`, `REVIEW_NOT_FOUND` |
| 409 | `INSUFFICIENT_CAPACITY`, `CANCELLATION_DEADLINE_PASSED`, `EDIT_DEADLINE_PASSED`, `RESERVATION_NOT_ACTIVE`, `REVIEW_ALREADY_EXISTS`, `LAST_IMAGE`, `TOO_MANY_IMAGES`, `RESTAURANT_HAS_UPCOMING_RESERVATIONS`, `CONFLICTS_WITH_EXISTING_RESERVATIONS` |
| 500 | `INTERNAL_ERROR` (details are logged, never returned) |

## Business rules and design decisions

### Reservations
- **Capacity** is the number of people the restaurant can seat at once. There are no tables.
- **Intervals** are half-open `[start_at, end_at)`, so 17:00–18:00 and 18:00–19:00 do not overlap.
- **Capacity check.** The API loads the ACTIVE reservations that overlap the request and runs a sweep line over their start and end events. At equal times, departures are processed before arrivals. The request is accepted when `peak + party_size <= capacity`. This never adds up reservations that overlap the request but not each other.
- **Concurrency.** Every write that affects a restaurant (create, edit or cancel a reservation, or edit the restaurant) first takes `SELECT … FOR UPDATE` on the restaurant row, inside a READ COMMITTED transaction. Two requests competing for the same seats therefore run one after the other, and the second sees the first's booking. Locks are always taken restaurant first, then reservation, so they cannot deadlock.
- **Time.** The server clock decides what counts as the past. Bookings must start after now and lie entirely within opening hours on one local day.
- **Cancellation** is allowed until and including `start_at − cancellation_notice_hours` (minimum 2 hours). It sets `status = CANCELLED` and `cancelled_at`. Nothing is deleted.

### Reviews and ratings
- One review per user per restaurant, enforced by a unique constraint that also covers concurrent submissions. Owners cannot review their own restaurant.
- `average_rating` (1 decimal place) and `review_count` are computed live with `AVG` and `COUNT`. A restaurant without reviews returns `average_rating: null` ("No reviews yet").
- **Sorting.**
  - Highest Rating: `ROUND(avg, 1) DESC NULLS LAST, review_count DESC, id ASC`.
  - Most Reviewed: `review_count DESC, ROUND(avg, 1) DESC NULLS LAST, id ASC`.

  Ratings are compared at the displayed precision, so restaurants that both show 4.3 are ordered by the tie-breakers.

### Assumptions (decided before implementation)
| | Decision |
|---|---|
| A1 | Each restaurant has an IANA `timezone` (default `Asia/Bangkok`). Opening hours and booking times are interpreted in it |
| A2 | No overnight opening hours in v1 (`closing_time > opening_time`). A booking must fall on one local day |
| A3 | A reservation can be edited only while ACTIVE and only until the same deadline as cancellation. This prevents bypassing the policy by moving a booking and then cancelling it |
| A4 | Changing capacity, hours or timezone is rejected (409) if an unfinished ACTIVE reservation would no longer fit. Customer bookings are never cancelled automatically |
| A5 | Deleting a restaurant is rejected (409) while it has unfinished ACTIVE reservations. Otherwise it is hard-deleted along with its images, reviews and past reservations |
| A6 | Any signed-in user who is not the owner may review. No completed reservation is required |
| A7 | Owners may book their own restaurant |
| A8 | Reading someone else's reservation returns 403 |
| — | The review comment is optional (max 2000 characters). Capacity is capped at 10,000 and notice hours at 1 year, purely as sanity limits |
| — | Images must be JPEG, PNG or WebP (detected from the content), at most 5 MB each and at most 10 per restaurant |

## Project structure

```
backend/
  cmd/api            HTTP server entry point
  cmd/migrate        migration runner
  internal/auth      JWT verification (go-oidc) and request identity
  internal/middleware  auth (required and optional), security headers, logging, recovery
  internal/handler   HTTP request and response mapping and the route table
  internal/service   business rules: restaurants, reservations (sweep line), reviews
  internal/repository  all SQL (pgx) and transactions
  internal/storage   image storage interface and local-disk implementation
  internal/apperr    typed errors and the shared error response
  migrations         SQL migrations (embedded)
frontend/            Next.js 16 App Router (Auth.js v5, Tailwind v4)
keycloak/            realm import for development
docker/postgres/     database init script (creates keycloak and tabletalk_test DBs)
```

## Security notes

- Ownership and authorization are checked in the service layer on every write. Hiding a button in the UI is never treated as authorization.
- API responses send `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY` and `Cache-Control: no-store`. Uploaded images are validated by content and served with `nosniff`.
- The API has no CORS configuration because only the Next.js server calls it, and it authenticates with bearer tokens rather than cookies.
- Known v1 limitations: no rate limiting and no audit log.
