# TODO

Remaining work collected from the project review session (schema, architecture, `create activity`, testing). Grouped by urgency.

## High priority — direct continuation of `create activity`

- [ ] **Real auth middleware.** `ActivityHandler.CreateActivity` still uses `userID := uuid.New()` as a placeholder — not yet wired to a real session/token (`c.Locals("user_id")` or similar). Without this, every created activity gets assigned to a random user.
- [ ] `GetActivity`, `ListActivities`, `UpdateActivity`, `SoftDeleteActivity`, `RestoreActivity` — usecase + handler, following the same pattern as `CreateActivity`. The sqlc queries already exist in `queries/activities.sql`.
- [ ] Split `schema.sql` / `schema_down.sql` into golang-migrate-formatted migration files (`NNNN_name.up.sql` / `.down.sql`) under `db/migrations/`.
- [ ] Run `swag init` to actually generate the contents of `docs/` (the swaggo annotations on the handler haven't been generated from yet).
- [ ] Check `internal/config/config.go` — never touched in this session; confirm `GetDSN()` and `ServerPort` match what's actually needed.
- [ ] Replace `log.Printf`/`log.Fatalf` in `main.go`, `app.go` with `log/slog`. Add `github.com/samber/slog-fiber` for per-request access logging. Replace the logging in `CustomErrorHandler` with `slog.ErrorContext(c, ...)` — see the note above about "log at one single point only."

## Major features not started at all

- [ ] **Full auth flow** — register, login, session refresh, email verification. Tables & queries already exist (`user_accounts`, `user_sessions`, `user_verifications`), but there isn't a single usecase/handler yet.
- [ ] **Scheduler worker** — a job that evaluates cron expressions from `activity_schedules` and auto-generates `activity_items`. The `ListActiveSchedulesForGeneration` query exists, but there's no worker actually parsing cron and calling `CreateScheduledActivityItem`.
- [ ] **Timeout worker** — a periodic job calling `MarkOverdueItemsAsMissed` (see the note on retroactive timeout below before implementing this).
- [ ] **Statistics endpoints** — heatmap & confirmation-delay chart. The `GetHeatmapData` / `GetConfirmationDelayStats` queries exist, but there's no usecase/handler yet.
- [ ] **Image upload** — `activity_images` / `activity_item_images` need multipart file upload middleware, not discussed at all yet.

## Cross-cutting / infrastructure

- [ ] Structured logging via `log/slog` (see details above) — **already decided, just needs implementing.**
- [ ] Health check endpoint (`/healthz`) for container orchestration.
- [ ] CORS, rate limiting.
- [ ] `Dockerfile` + `docker-compose.yml` (app + Postgres) for deployment/dev environment.
- [ ] CI (GitHub Actions): run unit tests on every PR; ideally integration tests too (needs Docker-in-Docker on the runner, or at least `docker` available on the executor).
- [ ] CI: check mocks aren't stale — run `mockery` then `git diff --exit-code internal/usecase/mocks/`, failing if someone forgot to regenerate after changing an interface.

## Items still needing reconciliation (temporary assumptions on my end, verify against your actual code)

- [ ] `pkg/util/validation.go` — the `ErrorResponse` there was my assumption (your own comment said "assuming this is your util return type"). If a different definition already exists elsewhere in `pkg/util`, delete my struct, keep only `FormatValidationErrors`, and adjust field names.
- [ ] `pkg/ptr.To` etc. — confirm you're adopting these, or rename to match your own existing convention if you have one.
- [ ] `sqlc.yaml` override `uuid -> github.com/google/uuid.UUID` — make sure it's been regenerated (`sqlc generate`), and confirm `internal/db` actually uses `uuid.UUID`, not `pgtype.UUID` (the mapper in `activity_mapper.go` assumes the former).

## Open design issues (full detail in `SCHEMA_REVIEW.md`)

- [ ] Race condition: concurrent schedule edits (no locking/versioning yet).
- [ ] `is_fixed_schedule` can drift out of sync with the contents of `activity_schedules` (no enforcement, purely an app-layer discipline).
- [ ] Changing `confirmation_timeout_minutes` applies retroactively to old `awaiting` items — decide whether this is the intended behavior before implementing the timeout worker.
- [ ] Soft-deleting a user doesn't automatically revoke their `user_sessions` — needs `DeleteAllSessionsByUserID` called explicitly as part of the `SoftDeleteUser` flow.
- [ ] Image files (`activity_images`/`activity_item_images`) become orphaned after a hard delete — no cleanup job yet.
- [ ] `user_verifications` is vulnerable to replay (no `consumed_at` / row-lock during validation).
- [ ] `activity_items.status` transitions aren't restricted (a `missed` item could go back to `captured`) — decide whether late confirmation should actually be supported.

## Code tidiness (not urgent yet, but worth watching for these thresholds)

- [x] Split `router.go` per domain (`activity_routes.go`, etc.) — in progress.
- [ ] `activity_usecase.go` — once it exceeds ~5-6 methods or ~250-300 lines, consider splitting per operation (`activity_usecase_create.go`, etc.), keeping the interface in one main file.
- [ ] `activity_dto.go` — keep one file per entity (not per operation) as long as its size stays reasonable.
