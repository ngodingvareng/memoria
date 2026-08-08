# TODO

Remaining work, ordered by priority (security-critical first, product-blocking next, then completeness, ops, open design decisions, and low-urgency housekeeping last).

## Tier 2 — Core product features not started

Both Commitment workers below are **deferred on purpose** — see `COMMITMENT.md`. Commitment is a long-horizon feature whose design isn't settled, and C5 there (a manual capture landing on a Thread with an open **due** Moment) would be a first-week bug for every user if the scheduler shipped before it's decided. Don't start either worker until those questions are answered.

- [ ] **Scheduler worker** — a job that evaluates cron expressions from `commitments` and auto-generates `moments`. The `ListCommitmentsForGeneration` query exists, but there's no worker actually parsing cron and calling `CreateCommittedMoment`. Blocked on `COMMITMENT.md` C1, C4, C5, C6.
- [ ] **Timeout worker** — a periodic job calling `MarkOverdueMomentsAsMissed` (see the retroactive-timeout decision under Tier 5 before implementing this). Blocked on `COMMITMENT.md` C2 and C3.
- [ ] **Statistics endpoints** — heatmap & Time to Tell chart. The `GetHeatmapData` / `GetSettlingTimeStats` queries exist, but there's no usecase/handler yet.
- [ ] `RestoreThread` — usecase + handler. The sqlc query already exists (`internal/db/threads.sql.go`), nothing calls it yet.
- [x] ~~`moment_images` upload~~ — done: `Moment`/`MomentImage` now have the full repository/usecase/handler stack (create/update/soft-delete/get/list/search for Moments; upload/list/delete for MomentImage), mirroring `Thread`/`ThreadImage`. `RestoreMoment`, `TouchMomentLastViewed`, `ListMomentsOnDates` (Echoes), and the heatmap/Time-to-Tell queries above remain unwired, same as their Thread-side counterparts.

## Tier 3 — Auth feature completeness

- [ ] **OAuth (Google/GitHub)** — not implemented, only the `credential` provider path exists. `user_accounts.provider_id` and the schema already support it; no usecase/handler yet.
- [ ] **Email verification** — not implemented. `SetUserEmailVerified` / `user_verifications` queries exist but nothing calls them yet.
- [ ] **Password reset** — not implemented.

## Tier 4 — Operational readiness

- [x] ~~Health check endpoint (`/healthz`) for container orchestration.~~ — done: `HealthHandler` pings the DB pool and is registered unauthenticated at `GET /healthz`.
- [x] ~~`Dockerfile` + a deployment `docker-compose.yml` (app + Postgres)~~ — done: multi-stage `api/Dockerfile` plus `api/docker-compose.yml` (postgres + a one-shot `migrate` service + the app, healthchecked). `docker-compose.dev.yaml` is unchanged and still covers local Postgres/RustFS for `make dev`. Added `DATABASE_HOST` (`internal/config/config.go`, defaults to `localhost`) since `GetDSN` could no longer hardcode it once the app itself runs in a container and must reach Postgres by service name.
- [x] ~~CI: check mocks aren't stale~~ — done: `check-mocks` job in `.github/workflows/api-test.yaml` runs `mockery` then `git diff --exit-code internal/usecase/mocks/`.
- [x] ~~CI: `PREPARE` every generated query against a real Postgres.~~ — done: `api/cmd/checksql` extracts every sqlc `const` from `internal/db/*.sql.go` via `go/parser` and `PREPARE`s each against a migrated Postgres (also runnable locally as `make check-sql`); wired into a `check-sql` job in `.github/workflows/api-test.yaml`. Running it surfaced one real breakage — `ListCommitmentsForGeneration` still referenced `threads.has_commitment`, dropped by migration 000004 in favor of `commitments.paused_at`/`archived_at` — now fixed in `db/queries/commitments.sql` and regenerated.

## Tier 5 — Open design decisions (full detail in `SCHEMA_REVIEW.md`)

### Schema/spec gaps against FEATURES.md

The schema already uses the `FEATURES.md` vocabulary (Thread/Moment/
Commitment), but three places still disagree on semantics, and each
needs a model change:

- [ ] **`moments.status` is `NOT NULL`**, so a manually captured Moment is forced to carry a state it shouldn't have — `FEATURES.md` is explicit that only Commitment-generated Moments have one ("a manually captured Moment has no state"). Either make the column nullable, or take Open Decision 5 in `FEATURES.md` and move Commitment outcomes into their own table, which would make the Commitment Firewall structural instead of a rule handlers have to remember.
- [ ] **`confirmation_timeout_minutes` lives on `threads`**, but `FEATURES.md` puts the confirmation window and strictness on each Commitment — two Commitments under one Thread (a strict morning run, a gentle evening walk) currently cannot differ. Move the column to `commitments` and add `strictness`.
- [ ] **`GetSettlingTimeStats` measures the wrong thing for its name.** It computes `due_at → captured_at`, which is Commitment punctuality. Time to Tell in `FEATURES.md` is `occurred_at → recorded_at` over *manual* Moments, and the spec says the two must never be merged or presented together. Needs a `recorded_at` column and a separate query; until then the existing one should be treated as adherence data, not Time to Tell.

### Pre-existing

- [ ] Race condition: concurrent Commitment edits (no locking/versioning yet).
- [ ] Decide whether changing `confirmation_timeout_minutes` should apply retroactively to old `due` Moments — needed before implementing the Timeout worker (Tier 2).
- [ ] `has_commitment` can drift out of sync with the contents of `commitments` (no enforcement, purely an app-layer discipline).
- [ ] `moments.status` transitions aren't restricted (a `missed` Moment could go back to `kept`) — decide whether late confirmation should actually be supported.
- [ ] Image files (`thread_images`/`moment_images`) become orphaned after a hard delete — no cleanup job yet.
- [ ] Soft-deleting a user doesn't revoke their refresh tokens — moot until a `SoftDeleteUser` usecase actually exists (currently only the sqlc query does); wire in a `RevokeAllByUserID` call when that flow gets built.
- [ ] `user_verifications` is vulnerable to replay (no `consumed_at` / row-lock during validation) — relevant once email verification (Tier 3) is actually implemented.
- [ ] Login lockout (`IncrementFailedLoginAttempts` / `LockCredentialUserAccount`) isn't atomic against concurrent requests — a burst of parallel wrong-password attempts can overshoot `LOGIN_MAX_FAILED_ATTEMPTS` before the lock takes effect. Same class of race already accepted for schedule edits above; worth a single `UPDATE ... RETURNING` if it becomes a problem in practice.

## Tier 6 — Code tidiness (not urgent yet, watch for these thresholds)

- [ ] Split `router.go` per domain (`thread_routes.go`, etc.) once it outgrows a handful of domains — currently small enough to stay as one file.
- [ ] `thread_usecase.go` — once it exceeds ~5-6 methods or ~250-300 lines, consider splitting per operation (`thread_usecase_create.go`, etc.), keeping the interface in one main file.
- [ ] `thread_dto.go` — keep one file per entity (not per operation) as long as its size stays reasonable.
