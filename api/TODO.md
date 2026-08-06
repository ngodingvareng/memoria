# TODO

Remaining work, ordered by priority (security-critical first, product-blocking next, then completeness, ops, open design decisions, and low-urgency housekeeping last).

## Tier 2 — Core product features not started

- [ ] **Scheduler worker** — a job that evaluates cron expressions from `activity_schedules` and auto-generates `activity_captures`. The `ListActiveSchedulesForGeneration` query exists, but there's no worker actually parsing cron and calling `CreateScheduledActivityCapture`.
- [ ] **Timeout worker** — a periodic job calling `MarkOverdueItemsAsMissed` (see the retroactive-timeout decision under Tier 5 before implementing this).
- [ ] **Statistics endpoints** — heatmap & confirmation-delay chart. The `GetHeatmapData` / `GetConfirmationDelayStats` queries exist, but there's no usecase/handler yet.
- [ ] `RestoreActivity` — usecase + handler. The sqlc query already exists (`internal/db/activities.sql.go`), nothing calls it yet.
- [ ] `activity_capture_images` upload — needs the identical repository/usecase/handler pattern already used for `activity_images`.

## Tier 3 — Auth feature completeness

- [ ] **OAuth (Google/GitHub)** — not implemented, only the `credential` provider path exists. `user_accounts.provider_id` and the schema already support it; no usecase/handler yet.
- [ ] **Email verification** — not implemented. `SetUserEmailVerified` / `user_verifications` queries exist but nothing calls them yet.
- [ ] **Password reset** — not implemented.

## Tier 4 — Operational readiness

- [ ] Health check endpoint (`/healthz`) for container orchestration.
- [ ] `Dockerfile` + a deployment `docker-compose.yml` (app + Postgres) — `api/docker-compose.dev.yaml` only covers local Postgres/RustFS for `make dev`, not shipping the app itself.
- [ ] CI: check mocks aren't stale — run `mockery` then `git diff --exit-code internal/usecase/mocks/` as a step in `.github/workflows/api-test.yaml`, failing if someone forgot to regenerate after changing an interface.

## Tier 5 — Open design decisions (full detail in `SCHEMA_REVIEW.md`)

- [ ] Race condition: concurrent schedule edits (no locking/versioning yet).
- [ ] Decide whether changing `confirmation_timeout_minutes` should apply retroactively to old `awaiting` items — needed before implementing the Timeout worker (Tier 2).
- [ ] `is_fixed_schedule` can drift out of sync with the contents of `activity_schedules` (no enforcement, purely an app-layer discipline).
- [ ] `activity_captures.status` transitions aren't restricted (a `missed` item could go back to `captured`) — decide whether late confirmation should actually be supported.
- [ ] Image files (`activity_images`/`activity_capture_images`) become orphaned after a hard delete — no cleanup job yet.
- [ ] Soft-deleting a user doesn't revoke their refresh tokens — moot until a `SoftDeleteUser` usecase actually exists (currently only the sqlc query does); wire in a `RevokeAllByUserID` call when that flow gets built.
- [ ] `user_verifications` is vulnerable to replay (no `consumed_at` / row-lock during validation) — relevant once email verification (Tier 3) is actually implemented.
- [ ] Login lockout (`IncrementFailedLoginAttempts` / `LockCredentialUserAccount`) isn't atomic against concurrent requests — a burst of parallel wrong-password attempts can overshoot `LOGIN_MAX_FAILED_ATTEMPTS` before the lock takes effect. Same class of race already accepted for schedule edits above; worth a single `UPDATE ... RETURNING` if it becomes a problem in practice.

## Tier 6 — Code tidiness (not urgent yet, watch for these thresholds)

- [ ] Split `router.go` per domain (`activity_routes.go`, etc.) once it outgrows a handful of domains — currently small enough to stay as one file.
- [ ] `activity_usecase.go` — once it exceeds ~5-6 methods or ~250-300 lines, consider splitting per operation (`activity_usecase_create.go`, etc.), keeping the interface in one main file.
- [ ] `activity_dto.go` — keep one file per entity (not per operation) as long as its size stays reasonable.
