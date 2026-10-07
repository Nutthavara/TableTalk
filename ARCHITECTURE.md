# TableTalk architecture

This document explains how TableTalk is built and why. Setup steps and the full API reference are in [README.md](README.md). Conventions and gotchas for contributors are in [AGENTS.md](AGENTS.md).

## System overview

```
Browser ──► Next.js 16 (App Router, server-side BFF) ──Bearer JWT──► Go REST API ──► PostgreSQL 18
               │ ▲                                                      │
               ▼ │ OIDC Authorization Code + PKCE                       │ verifies JWT via JWKS
             Keycloak 26 ◄──────────────────────────────────────────────┘
```

| Component | Role | Key technology |
|---|---|---|
| Next.js (`frontend/`) | All pages, both customer and owner. Runs as a BFF: it holds the session and calls the API from the server | Next 16 App Router, React 19, Auth.js v5 (Keycloak provider), Tailwind v4 |
| Go API (`backend/`) | The only authority for business rules and authorization | stdlib `net/http` ServeMux, pgx/v5, goose, go-oidc/v3, slog |
| PostgreSQL 18 | Data, constraints, row locks | `timestamptz`, CHECK and UNIQUE constraints, `uuidv7()` |
| Keycloak 26.4 | Identity: login, registration, tokens | Realm imported from `keycloak/realm-tabletalk.json` |

In development, Docker Compose runs Postgres and Keycloak (Keycloak uses its own `keycloak` database in the same Postgres). The API and web app run on the host for easy debugging, and so the token issuer URL is identical for the browser and the API.

Dependencies are kept small on purpose:
- **Go:** no web framework and no ORM. Validation is explicit code in the services, because most rules need database state.
- **Frontend:** no UI kit (hand-built components on native `<dialog>`) and no form or validation library. HTML constraints handle UX, and the API's `details` handle field errors.

## Authentication and identity

1. **Sign in.** The user clicks Log in, which goes to `/login`. That route handler calls Auth.js `signIn("keycloak")`, which uses the authorization code flow with PKCE S256 and state. `/register` adds `prompt=create`.
2. **Callback.** The Next server exchanges the code and stores the access, refresh and ID tokens in an **encrypted, chunked, httpOnly** cookie (`authjs.session-token.*`). The session object the browser can read has no tokens.
3. **Refresh** happens in two places:
   - The Auth.js `jwt` callback refreshes about 60 s before expiry and persists the result through `proxy.ts`.
   - `getAccessToken()` refreshes in memory when the current request still carries an expired token.
4. **API calls.** `lib/api.ts` (`server-only`) sends `Authorization: Bearer <access token>`.
5. **Verification.** The Go middleware (`auth.Verifier`, go-oidc) checks:
   - the RS256 signature against Keycloak's JWKS (fetched lazily and cached)
   - `iss`
   - `aud` contains `tabletalk-api` (added by an audience mapper in the realm)
   - `exp`
   - `sub` is a UUID

   It then upserts a `users` profile row (`SyncUser`) and puts `auth.Identity` into the request context.
6. **Middleware modes:**
   - `Required` returns 401 without a token.
   - `Optional` (public GETs) lets anonymous requests through, but **any token that is present must be valid**; a bad token is never downgraded to anonymous.
7. **Sign out** clears the local session and redirects to Keycloak's end-session endpoint with `id_token_hint`.

Keycloak clients:
- `tabletalk-web`: confidential, used by Next.js.
- `tabletalk-dev-cli`: public, password grant, for getting tokens with `curl` during development only. Remove it in production.

## Backend structure

```
cmd/api            wiring: config → pool → store → services → handlers → router; graceful shutdown; server timeouts
cmd/migrate        goose runner over embedded migrations
internal/auth      Verifier (go-oidc), Identity in context; authtest = in-memory signer for tests
internal/middleware Required/Optional auth, SecurityHeaders, Logger, Recover
internal/handler   decode/encode, path IDs (malformed → 404), route table (Routes), error mapping
internal/service   business rules: RestaurantService, ReservationService, ReviewService, sweep line, hours
internal/repository all SQL; Store.InTx; Queries usable on pool or tx
internal/storage   Storage interface; Local (os.Root, atomic writes, file server without listings)
internal/apperr    typed errors + shared response writer
internal/model     domain types (Restaurant, Reservation, Review, TimeOfDay)
migrations         goose SQL (embedded)
```

- **Request path:** `Recover → Logger → SecurityHeaders → ServeMux → auth (Required/Optional) → handler → service → repository`.
- **Route table:** `handler.Routes()` is the single list of API routes. `NewRouter` wraps each one in auth according to its `AuthMode`, and a test walks the table to prove that every non-GET route requires auth.
- **Security headers:** every response gets `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY` and `Referrer-Policy: no-referrer`. `/api/*` responses also get `Cache-Control: no-store`.
- **CORS:** none configured, because only the Next server calls the API and it authenticates with bearer tokens, not cookies.

## Data model

```sql
users(id uuid PK = Keycloak sub, username, display_name, email, created_at, updated_at)

restaurants(id uuid PK uuidv7, owner_id → users, name 1..120, description ≤2000,
            capacity > 0, opening_time time, closing_time time, timezone text DEFAULT 'Asia/Bangkok',
            cancellation_notice_hours >= 2, created_at, updated_at,
            CHECK closing_time > opening_time)

restaurant_images(id, restaurant_id → restaurants ON DELETE CASCADE, storage_key UNIQUE,
                  content_type, position >= 0, UNIQUE(restaurant_id, position))

reservations(id, restaurant_id → restaurants CASCADE, user_id → users, party_size > 0,
             start_at timestamptz, end_at timestamptz, status ACTIVE|CANCELLED, cancelled_at,
             CHECK end_at > start_at, CHECK (status = CANCELLED) = (cancelled_at IS NOT NULL))
  INDEX (restaurant_id, start_at, end_at) WHERE status = 'ACTIVE'   -- overlap query
  INDEX (user_id, start_at DESC)                                     -- My Reservations

reviews(id, restaurant_id → restaurants CASCADE, user_id → users, rating smallint 1..5,
        comment ≤2000, UNIQUE(user_id, restaurant_id))
  INDEX (restaurant_id, created_at DESC)
```

- **Ratings are not stored on `restaurants`.** Every restaurant query computes `ROUND(AVG(rating),1)` and `COUNT(*)` with a `LEFT JOIN LATERAL` over `reviews`. A restaurant without reviews gets `average_rating = null`, shown as "No reviews yet", never as 0 stars.
- **Rules that span rows can't be DB constraints**, so services enforce them: capacity, at least one image per restaurant, and the owner not reviewing their own restaurant.
- **Time:** all instants are stored as `timestamptz` and returned in UTC. `opening_time` and `closing_time` are wall-clock times in the restaurant's IANA `timezone`.

## Reservations

**Input:** `{party_size, date: YYYY-MM-DD, start_time: HH:MM, end_time: HH:MM}`, as restaurant-local wall time. The server converts it to UTC using the restaurant's timezone, so clients never do timezone math.

**Rules**, applied in order inside one transaction (`ReservationService.checkBookable`):
1. **Party size:** `0 < party_size ≤ capacity` → otherwise 400 `PARTY_SIZE_EXCEEDS_CAPACITY` (or `VALIDATION_ERROR` for ≤ 0).
2. **Time:** `start < end`, and `start_at > now` by the server clock → otherwise 400 `RESERVATION_IN_PAST` (or `VALIDATION_ERROR`).
3. **Opening hours:** the whole interval lies within opening hours on one local day → otherwise 400 `OUTSIDE_OPENING_HOURS`.
4. **Capacity:** at every instant of the requested interval, occupied seats plus `party_size` must not exceed capacity → otherwise 409 `INSUFFICIENT_CAPACITY`.

**Capacity algorithm (sweep line).** Load the ACTIVE reservations that overlap the request (`start_at < req.end AND end_at > req.start AND id <> excludeID`). Clip each to the requested window, emit `(start, +size)` and `(end, −size)` events, and sort by time. **At equal times, departures come before arrivals**, so `[17:00,18:00)` and `[18:00,19:00)` never count together. Track the running sum and accept when `peak + party_size ≤ capacity`. This is O(n log n).

Why not `SUM(party_size)` over the overlapping rows: two bookings can each overlap the request without overlapping each other. With A = 17:00–17:30 for 6 and B = 18:00–18:30 for 6, a request for 17:00–19:00 for 4 fits under capacity 10, but a plain SUM would reject it.

**Concurrency.**
- Every write that affects a restaurant starts with `SELECT … FROM restaurants WHERE id = $1 FOR UPDATE` (READ COMMITTED). Competing bookings for one restaurant therefore run one after another, and the second sees the first's insert.
- Edit and cancel read the reservation's `restaurant_id` (which never changes), lock the restaurant, then lock the reservation, then re-check owner and status. The lock order is always the same, so there are no deadlocks.
- **Alternatives considered:**
  - SERIALIZABLE with retries: correct, but adds retry logic and harder errors.
  - Advisory locks: less visible in the code.
  - An exclusion constraint: can't express a sum.
- **Cost:** writes to one restaurant are serialized. Transactions are short, and other restaurants are unaffected.
- **Proof:** tests run 20 to 60 concurrent bookings or edits. With the lock removed, overbooking actually happened (up to 23 seats against a capacity of 10).

**Edit and cancel.**
- Edit uses the same rules as create, excluding the reservation itself.
- Both edit and cancel are allowed only while the reservation is ACTIVE and `now ≤ start_at − cancellation_notice_hours`, boundary included.
- Cancel is logical. `DELETE /api/reservations/{id}` returns the reservation with status CANCELLED.
- Responses include `cancellation_deadline` and `can_modify` for the UI. The API still enforces the rule itself.

**My Reservations filters** are computed from the server clock:
- `upcoming`: ACTIVE and `end_at > now`, which includes reservations in progress.
- `past`: ACTIVE and finished.
- `cancelled`: CANCELLED.

## Restaurants, listing and reviews

- **Create** is multipart: the fields plus at least one image, in one request, so a restaurant never exists without a photo.
- **Update** is a JSON full replacement. A change to capacity, hours or timezone is rejected (409 `CONFLICTS_WITH_EXISTING_RESERVATIONS`) if any unfinished ACTIVE reservation would no longer fit. The check uses the same sweep line and hours rules.
- **Delete** is rejected while unfinished ACTIVE reservations exist. Otherwise it cascades (images, reviews, past reservations), and image files are removed after commit.
- **Listing** is `GET /api/restaurants?q=&sort=rating|reviews&page=&page_size=` with offset pagination (page size ≤ 50) and deterministic ordering:
  - Highest Rating: `ROUND(avg,1) DESC NULLS LAST, review_count DESC, id ASC`
  - Most Reviewed: `review_count DESC, ROUND(avg,1) DESC NULLS LAST, id ASC`

  Ratings are compared at the displayed precision, so restaurants that both show "4.3" fall through to the tie-breakers. `q` is an ILIKE substring match, and `%`, `_` and `\` are escaped.
- **Reviews** are listed newest first. Restaurant detail includes `my_review_id` for the viewer, and each listed review includes `is_mine`. There is no single-review GET, so the frontend pages through the list to find the viewer's review.

## Frontend architecture

**Routes** (from the design blueprint):

| Route | Purpose | Access |
|---|---|---|
| `/` | Redirects to `/restaurants` | Public |
| `/restaurants` | List, search, sort, pagination | Public |
| `/restaurants/[id]` | Gallery, info, booking panel, reviews. Owner mode when it's your own restaurant | Public; actions need login |
| `/my-reservations?tab=` | Upcoming / Past / Cancelled with counts | Login |
| `/my-reservations/[id]/edit` | Edit Booking with current vs new comparison | Login, booker |
| `/my-restaurants` | Owned restaurants; delete with typed-name confirmation | Login |
| `/my-restaurants/new`, `/my-restaurants/[id]/edit` | 4-section restaurant form | Login, owner |
| `/login`, `/register` | Route handlers that redirect to Keycloak (`callbackUrl` limited to same-origin paths) | Public |

**Layering:**
- `features/*/queries.ts` contains server-only reads that call the API.
- `features/*/actions.ts` contains Server Actions that call the API and return `ActionState`.
- `features/*/components` holds the client and server UI.
- Pages compose these.
- `lib/auth-guard.ts` provides `requireSignedIn` (a UX redirect) and `handlePageError`, which turns 401 into a login redirect and 403 or 404 into the not-found page.

**Key flows:**
- **Booking.** The panel checks the booking in the browser for fast feedback (`booking.ts`, using the restaurant's timezone), then shows a confirmation summary with the cancellation deadline. Confirm calls `createReservation`. The API's `INSUFFICIENT_CAPACITY` leads to the "Not enough seats" step, and success shows "You're booked!" with a toast.
- **Guests** can fill in the booking form. "Log in to reserve" carries the draft in the return URL, so the form is restored after login.
- **Toasts.** A small client store powers `toast()`. Server Actions that redirect leave a base64url flash cookie, which `<Toaster/>` reads after navigation.
- **Images** are loaded directly from the API (`unoptimized`). Reservation cards fetch restaurant covers separately, because reservations carry no images.
- **Responsive:**
  - Desktop has a top nav and the profile menu.
  - On mobile (< 640 px):
    - a compact header and a bottom tab bar
    - on the detail page, a sticky "Reserve a Table" bar that opens the booking panel as a bottom sheet
    - dialogs become bottom sheets

## Assumptions and decisions

| | Decision |
|---|---|
| A1 | Each restaurant has an IANA `timezone` (default Asia/Bangkok). The form doesn't show it; new restaurants use the default and edits keep the existing value |
| A2 | No overnight opening hours. A booking must fall on one local day |
| A3 | Edit is allowed only until the cancellation deadline, so the policy can't be bypassed by moving a booking and then cancelling it |
| A4 | Restaurant changes that would invalidate unfinished ACTIVE reservations are rejected (409). Bookings are never cancelled automatically |
| A5 | Deleting a restaurant is rejected while unfinished ACTIVE reservations exist; otherwise it is hard-deleted with cascade |
| A6 | Any signed-in non-owner may review. No completed reservation is required |
| A7 | The API allows owners to book their own restaurant. **The UI hides booking for owners**, as the design specifies |
| A8 | Reading someone else's reservation returns 403 (the frontend shows the 404 page) |
| — | Review comment is optional (≤2000). Sanity caps: capacity ≤10,000, notice ≤1 year |
| — | Design vs API: the design says editing stays possible after the cancellation deadline, but the API (A3) forbids it, so the UI locks both actions and explains why |
| — | The upload progress bar is indeterminate, because Server Actions don't report progress |
| — | Search was the only backend change made for the design, with the owner's approval |

## Testing strategy

| Layer | What | Where |
|---|---|---|
| Pure unit | Sweep line, opening hours (timezone-aware) | `service/availability_test.go` |
| Service + real DB | Spec tests 1–16, the A3 deadline, filters, concurrent create/edit, review rules | `service/*_test.go` (fixed clock) |
| HTTP + real DB | Authorization, forged identity fields, validation, sorting/pagination/search, status codes | `handler/*_test.go` |
| Security | Every non-GET route requires auth (via the route table); panics don't leak details; security headers | `handler/security_test.go`, `middleware/*_test.go` |
| Browser E2E | 20 flows across guest, customer, owner and mobile with real Keycloak and API | Session scratchpad (not committed), playwright-core with installed Chrome |

Final verification on a fresh stack (2026-09-29):
- 51 Go tests and 82 subtests passed under `-race`, with 0 skipped.
- 20/20 E2E flows passed.
- govulncheck and npm audit found no vulnerabilities.

## Future work

- A Keycloak login and registration theme matching the design (FreeMarker and CSS, mounted in docker-compose).
- Rate limiting and an audit log.
- Object storage (S3/MinIO) behind `storage.Storage`, plus image resizing or a CDN.
- Denormalized rating columns or a materialized view if listing becomes slow. Keyset pagination for large lists.
- Overnight opening hours (would change A2 and the hours check).
- A single-review endpoint to avoid paging when looking up the viewer's review.
- Committing the browser E2E suite (with `playwright` as a dev dependency) to the repo and running it in CI.
