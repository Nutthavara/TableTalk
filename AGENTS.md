# AGENTS.md — working on TableTalk

Guide for coding agents and developers continuing work on this repo. For how the system is built and why, read [ARCHITECTURE.md](ARCHITECTURE.md). For setup, environment variables and the API reference, read [README.md](README.md).

TableTalk is a restaurant reservation and review system:
- **Backend:** Go REST API (`backend/`)
- **Frontend:** Next.js 16 App Router (`frontend/`)
- **Auth:** Keycloak
- **Database:** PostgreSQL 18

One account can both own restaurants and book or review other people's restaurants. There are no roles. You own a restaurant when `restaurant.owner_id` equals your user ID.

## Status (2026-10)

| Area | State |
|---|---|
| Phases 1–10 (infra, auth, restaurants, listing, reservations, edit/cancel, reviews, UI, security, docs/verification) | Done |
| Search `?q=` on `GET /api/restaurants` | Done (the only backend change made while building the UI) |
| Keycloak login/register theme matching the design | Done: `keycloak/themes/tabletalk` (extends `keycloak.v2`), selected by `loginTheme` in the realm |
| Known v1 limitations | No rate limiting, no audit log, no overnight opening hours, images on local disk |

The UI was built from the Claude Design canvas at https://claude.ai/artifact/W6dGecSXgfpVon2kWXVzee (24 boards: desktop, mobile, states, dialogs). The UI is in English.

## Run, test and verify

```bash
make up && make migrate        # Postgres :5432 + Keycloak :8080 + Mailpit :8025 (Keycloak needs ~20–40 s)
make api                       # Go API :8081
make web                       # Next.js :3000   (dev users alice/alice, bob/bob)
make test                      # go test -race ./... — needs Postgres running
make lint                      # go vet + eslint
cd frontend && npx tsc --noEmit && npm run build
```

- **Backend tests need `TEST_DATABASE_URL`.** `make test` exports it from `.env`. If it is unset, DB tests are **skipped silently**. Always confirm they ran: `go test -v ./... | grep -c -- '--- SKIP'` should print 0.
- **Each test package gets its own Postgres schema** (`internal/testutil.DB(t, "<pkg>")`), so packages can run in parallel. Tables are truncated per test.
- **Frontend route types are generated.** If `PageProps<"/new/route">` fails to typecheck, run `npx next typegen`.
- **Browser E2E** is not committed. During development it lived in the session scratchpad: `e2e/seed.mjs` seeds data through the API, and `e2e/e2e.mjs` runs 20 flows. It uses `playwright-core` with the installed Google Chrome (`channel: "chrome"`), installed outside the repo. Recreate it when you change UI flows.
- **Check that new tests can fail.** Briefly break the code under test (remove a lock, an ownership check, escaping, a tie-breaker) and confirm the test fails, then restore it. Earlier mutation checks once "passed" only because the mutated code didn't compile. Make sure a mutation still builds.

## Working agreement with the project owner

- **Work in phases.** After each phase, run build, test and lint, report the results honestly, and wait for approval before the next phase.
- **Commit messages:** write them in **Thai**, keep technical terms in English, **no `Co-Authored-By` trailer**, and **show the draft and wait for approval before `git commit`**.
- **Business rules and backend behaviour:** do not change them unless the owner approves. Ambiguities are recorded as assumptions A1–A8 in [ARCHITECTURE.md](ARCHITECTURE.md#assumptions-and-decisions).
- **Don't run destructive commands without asking.** This includes `rm -rf` of repo content, `make reset-infra`, and stopping other Docker containers. The owner's machine runs Colima, and other projects' containers may use ports 5432 and 8080.

## Non-negotiable invariants

Breaking any of these is a bug, even if the UI hides it.

1. **The Go API is the only authority** for authorization and business rules. Frontend checks are for UX only.
2. **Identity comes only from the verified JWT.** `owner_id` and `user_id` are never read from request input. JSON decoding uses `DisallowUnknownFields`, so smuggled fields return 400. Multipart create forwards only known fields.
3. **Every write that affects a restaurant locks the restaurant row first** (`Queries.LockRestaurant`, `SELECT … FOR UPDATE`). This covers restaurant edit/delete/images and reservation create/edit/cancel. When a reservation is involved, lock it **after** the restaurant. Never reverse this order.
4. **Capacity uses the sweep line** (`service/peakOccupancy`) over half-open `[start, end)` intervals, with departures processed before arrivals at the same instant. Never use `SUM(party_size)` over overlapping rows.
5. **Editing a reservation excludes the reservation itself** from the capacity check (`checkBookable(..., excludeID)`).
6. **Cancel and edit are allowed while `now <= start_at − cancellation_notice_hours`**, inclusive (A3). Cancellation is logical: `status = CANCELLED` plus `cancelled_at`.
7. **One review per (user, restaurant)** is enforced by a unique constraint, mapped to 409. Owners cannot review their own restaurant (403).
8. **Every error uses one format**: `{"error":{"code","message","details?"}}`. Unexpected errors become a generic 500 and are logged. `details` is present only for `VALIDATION_ERROR`.
9. **Timestamps are `timestamptz` and the API returns UTC.** Opening hours and booking inputs are wall-clock times in the restaurant's IANA `timezone`.
10. **Every non-GET route is `Required` auth.** `handler/security_test.go` enforces this through the route table. Add new routes to `handler.Routes()`.

## Backend conventions (`backend/`)

- **Layers:**
  - `handler` decodes and encodes HTTP and maps errors.
  - `service` holds all business rules and takes `auth.Identity` explicitly.
  - `repository` holds all SQL (pgx, hand-written, no ORM).
  - `service` decides transaction boundaries with `store.InTx(ctx, func(q *repository.Queries) error)`.
- **Errors:** services return `*apperr.Error` values (`apperr.New(status, CODE, message)` or `apperr.Validation(details)`), and handlers call `apperr.Write`. Repositories return `repository.ErrNotFound` or raw errors. Use `repository.IsUniqueViolation` for 23505.
- **Clock injection:** services take `now func() time.Time`, and tests pin it (see `service/fixture_test.go`, whose fixed "now" is 2030-01-15 09:00 Bangkok).
- **Test layout:**
  - Service tests use the real DB plus a fixed clock.
  - HTTP tests go through `handler.NewRouter` with tokens from `auth/authtest` (an in-memory RSA signer).
  - `newTestAPI(t)` in `handler/api_test.go` provides users alice and bob plus helpers.
- **Migrations:** goose SQL in `backend/migrations` is embedded in the binary. Add `0000N_name.sql` with Up and Down sections. Keep DB constraints as a second line of defence.
- **Storage:** images go through `storage.Storage`. `Local` uses `os.Root` (no path traversal) and writes atomically via temp file plus rename. Image type is detected from content (jpeg/png/webp), ≤5 MB each, ≤10 per restaurant.
- **Toolchain:** `go.mod` pins `toolchain go1.26.6` for stdlib security fixes. Run `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` after upgrading dependencies.

## Frontend conventions (`frontend/`)

**Next.js 16 differs from older versions.** Read `node_modules/next/dist/docs/` before using an API (see `frontend/AGENTS.md`). Points that matter here:
- `middleware` is now `proxy.ts`.
- `params` and `searchParams` are Promises.
- `PageProps<"/route">` is generated by typegen.
- `unauthorized()` and `forbidden()` are experimental (not used).
- After a mutation, call `refresh()` from `next/cache`.

**Layout**
- `app/` holds routes only.
- `features/<domain>/queries.ts` holds server-only reads.
- `features/<domain>/actions.ts` holds Server Actions.
- `features/<domain>/components/` holds UI.
- `components/ui` holds primitives; `components/layout` holds the shell.
- `lib/` holds API client, auth, routes, flash and formatting.

**BFF pattern:** only the Next server calls the Go API, through `lib/api.ts` (`server-only`), which attaches the bearer token from `lib/session.ts`. Tokens never reach browser JS, and `/api/auth/session` contains none.

**Server Actions**
- They return `ActionState` (`{status: "idle" | "success" | "error", code, message, fieldErrors}`) built with `toActionError`.
- Actions that navigate call `setFlash(...)` then `redirect(...)`.
- Actions that stay on the same page return success and let the client call `toast(...)`. A flash cookie is only picked up when the pathname or query changes, so redirecting to the current URL shows no toast.

**Gotchas**
- **Links to `/login` and `/register` must be plain `<a>`, not `next/link`.** They are route handlers that redirect to Keycloak, and an RSC prefetch fails with CORS.
- **API images are rendered with `next/image` `unoptimized`.** The optimizer refuses private hosts (an SSRF guard). Don't enable `dangerouslyAllowLocalIP`.
- **The flash cookie value is base64url JSON.** `cookies().set` already URL-encodes values, so encoding again breaks parsing.
- **Uploads go through Server Actions.** `next.config.ts` raises `serverActions.bodySizeLimit` and `proxyClientMaxBodySize` to 52 MB (10 × 5 MB). Keep both in sync with backend limits.
- **The client-side booking checks in `features/reservations/booking.ts` mirror the API rules for fast feedback only.** Availability is decided only by the API (`INSUFFICIENT_CAPACITY` takes the panel to its "Not enough seats" step).
- **The ESLint React rules are strict:** no `setState` directly inside effects and no ref access during render.

**Design tokens** are defined in `app/globals.css` (`@theme`): linen, paprika (CTA), herb (owner/success), saffron (stars only), chili (danger) and the ink scale.
- Fonts: Bricolage Grotesque (display), Figtree (body), Noto Sans Thai.
- Buttons: use `buttonClass(variant, size)`.
- Dialogs: use the `Dialog` component (native `<dialog>`, which becomes a bottom sheet on mobile).

## Keycloak theme (`keycloak/themes/tabletalk/login`)

- **It extends `keycloak.v2`.** Override only what's needed: `template.ftl` (TableTalk brand panel plus form column; the copy changes by `pageId`, e.g. `"register"`), `messages/messages_en.properties` (wording) and `resources/css/tabletalk.css` (design tokens over PatternFly v5 classes). Keep the v2 page templates, so reset password, verify email and similar flows keep working.
- **Setting changes on a running realm:** `realm-tabletalk.json` is imported only when the realm doesn't exist yet. Apply changes such as `loginTheme` or the client's `baseUrl` with `kcadm.sh` inside the container. Call `kcadm.sh` directly with `docker compose exec`; a `sh -c` wrapper may be blocked by a safety check. Avoid `make reset-infra` unless the owner approves.
- **Checking it in a browser:** run `kc-theme.mjs` in the scratchpad; it covers the login page, a wrong password, a successful login, registration and mobile. Delete any test user it creates (Keycloak user plus `users` row).
- **Users defined in the realm JSON must list `"realmRoles": ["default-roles-tabletalk"]`.** The import does not assign default roles to them, and without `manage-account` the Profile link (Keycloak Account Console) fails with 401. Users who sign up through the form get the roles automatically.
- **Forgot password** emails go to Mailpit (inbox http://localhost:8025, API `GET /api/v1/messages`). The browser check is `forgot.mjs` in the scratchpad; it resets bob's password back to `bob`.
- **Message files use Java MessageFormat:** write an apostrophe as `''`, or it disappears.
- **The account menu has no Profile item on purpose** (owner's decision): Keycloak's Account Console isn't themed, and account editing is out of scope.
- **Dev users:** alice/alice and bob/bob come from the realm import. alice's password was changed on this machine around 2026-10, so use bob for scripted logins or ask the owner.

## Environment gotchas seen on this machine

- **Colima's SSH mux may already be listening on :5432 and :8080** for other projects' containers. Check with `docker ps` before assuming the ports are free.
- **Keycloak's token `iss` must match `OIDC_ISSUER_URL` exactly.** This is why the API and web app run on the host, not in Docker.
- **When scripting a login flow with Python's `cookiejar`:** Keycloak cookies are marked `Secure` (browsers still send them to `http://localhost`), and the jar applies cookie deletions before additions. Prefer the Playwright E2E for anything beyond simple checks.
- **Next's Server Actions reject POSTs without an `Origin` header** (CSRF protection). Scripted calls must send it.
