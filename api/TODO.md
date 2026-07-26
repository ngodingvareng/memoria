# TODO

Remaining work collected from the project review session (schema, architecture, `create activity`, testing). Grouped by urgency.

## High priority — direct continuation of `create activity`

- [x] **Real auth middleware.** Implemented — `RequireAuth` (see `internal/delivery/rest/middleware/auth_middleware.go`) now populates the real authenticated user id; `ActivityHandler.CreateActivity`'s `uuid.New()` placeholder is gone.
- [ ] `GetActivity`, `ListActivities`, `UpdateActivity`, `SoftDeleteActivity`, `RestoreActivity` — usecase + handler, following the same pattern as `CreateActivity`. The sqlc queries already exist in `queries/activities.sql`.
- [x] Split `schema.sql` / `schema_down.sql` into golang-migrate-formatted migration files (`NNNN_name.up.sql` / `.down.sql`) under `db/migrations/`.
- [x] Run `swag init` to actually generate the contents of `docs/` (the swaggo annotations on the handler haven't been generated from yet).
- [ ] Check `internal/config/config.go` — never touched in this session; confirm `GetDSN()` and `ServerPort` match what's actually needed.
- [ ] Replace `log.Printf`/`log.Fatalf` in `main.go`, `app.go` with `log/slog`. Add `github.com/samber/slog-fiber` for per-request access logging. Replace the logging in `CustomErrorHandler` with `slog.ErrorContext(c, ...)` — see the note above about "log at one single point only."

## Major features not started at all

- [x] **Credential auth (register/login/logout)** — implemented: `internal/usecase/auth_usecase.go` + `internal/repository/{user,user_account,session}_repository.go` + `internal/security/{password,token}.go` + `internal/delivery/rest/handler/auth_handler.go`. Uses bcrypt for password hashing, SHA-256-hashed session tokens in an HttpOnly cookie, and a new `AuthUnitOfWork` (`internal/repository/auth_unit_of_work.go`) to insert the `users` + `user_accounts` rows atomically — the multi-repository transaction case flagged back when `activity_usecase.go`'s `WithTransaction` was built for a single repository only.
- [ ] **OAuth (google/github)** — not implemented, only the `credential` provider path exists. `user_accounts.provider_id` and the schema already support it; no usecase/handler yet.
- [ ] **Email verification** — not implemented. `SetUserEmailVerified` / `user_verifications` queries exist but nothing calls them yet.
- [ ] **Password reset** — not implemented.
- [ ] Session refresh / "remember me" — currently a single flat 30-day session (`sessionTTL` in `auth_usecase.go`), no separate short-lived access + long-lived refresh token split.
- [ ] Rate limiting / lockout on `Login` — nothing currently stops repeated failed-password attempts against one account or IP.
- [x] `config.go` needs two more fields: `SecureCookies bool` (see `AuthHandler` — must be `false` for local http:// dev, `true` in production) and confirm `cfg.GetDSN()` etc. still line up.
- [ ] `internal/repository/session_mapper.go` guesses the sqlc-generated field name for `ip_address` as `IpAddress` — verify against your actual generated code (could be `IPAddress` depending on sqlc's initialism handling).
- [ ] Ownership check still missing on activity image routes (`UploadActivityImage`/`ListActivityImages`/`DeleteActivityImage` don't verify the activity belongs to the now-real authenticated user) — was flagged before auth existed at all; now that `RequireAuth` is wired in, this is finally checkable but not yet done. Needs a `GetActivityByID` usecase/repository method (query already exists) to check `activity.UserID == authenticatedUserID`.
- [ ] No tests yet for the auth feature (usecase unit tests with mocks, repository integration tests) — same pattern as `activity`/`activity_image`, just not asked for yet.
- [ ] **Scheduler worker** — a job that evaluates cron expressions from `activity_schedules` and auto-generates `activity_captures`. The `ListActiveSchedulesForGeneration` query exists, but there's no worker actually parsing cron and calling `CreateScheduledActivityCapture`.
- [ ] **Timeout worker** — a periodic job calling `MarkOverdueItemsAsMissed` (see the note on retroactive timeout below before implementing this).
- [ ] **Statistics endpoints** — heatmap & confirmation-delay chart. The `GetHeatmapData` / `GetConfirmationDelayStats` queries exist, but there's no usecase/handler yet.
- [x] **Image upload** — implemented for `activity_images` using RustFS (S3-compatible). `activity_capture_images` needs the identical pattern (repository/usecase/handler), not yet done.

## Follow-ups from the image upload implementation

- [x] Add to `internal/config/config.go`: `StorageEndpoint`, `StorageRegion`, `StorageAccessKey`, `StorageSecretKey`, `StorageBucket`, `StorageUsePathStyle` (see `internal/app/app.go`'s `storage.NewS3Storage` call for exactly what's needed).
- [x] `go get` these: `github.com/aws/aws-sdk-go-v2`, `github.com/aws/aws-sdk-go-v2/config`, `github.com/aws/aws-sdk-go-v2/credentials`, `github.com/aws/aws-sdk-go-v2/service/s3`.
- [x] Merge `deploy/docker-compose.rustfs.yml` into the real `docker-compose.yml` once that exists.
- [ ] Ownership check missing: `UploadActivityImage`/`ListActivityImages`/`DeleteActivityImage` don't verify the activity actually belongs to the requesting user — same placeholder-auth gap as `CreateActivity`.
- [ ] Content-type validation is header-only right now (`fileHeader.Header.Get("Content-Type")`, client-supplied and spoofable) — validate actual file content (magic bytes) before trusting it's really an image. Flagged inline in `activity_image_handler.go`.
- [ ] RustFS itself is a young project (Beta as of mid-2026) — track its maturity before relying on it in production; the storage abstraction (`internal/storage.Storage`) makes swapping to MinIO/S3/R2 a config change, not a code change, if needed.

## Cross-cutting / infrastructure

- [x] Structured logging via `log/slog` (see details above) — **already decided, just needs implementing.**
- [ ] Health check endpoint (`/healthz`) for container orchestration.
- [ ] CORS, rate limiting.
- [ ] `Dockerfile` + `docker-compose.yml` (app + Postgres) for deployment/dev environment.
- [ ] CI (GitHub Actions): run unit tests on every PR; ideally integration tests too (needs Docker-in-Docker on the runner, or at least `docker` available on the executor).
- [ ] CI: check mocks aren't stale — run `mockery` then `git diff --exit-code internal/usecase/mocks/`, failing if someone forgot to regenerate after changing an interface.

## Items still needing reconciliation (temporary assumptions on my end, verify against your actual code)

- [ ] `pkg/util/validation.go` — the `ErrorResponse` there was my assumption (your own comment said "assuming this is your util return type"). If a different definition already exists elsewhere in `pkg/util`, delete my struct, keep only `FormatValidationErrors`, and adjust field names.
- [ ] `pkg/ptr.To` etc. — confirm you're adopting these, or rename to match your own existing convention if you have one.
- [x] `sqlc.yaml` override `uuid -> github.com/google/uuid.UUID` — make sure it's been regenerated (`sqlc generate`), and confirm `internal/db` actually uses `uuid.UUID`, not `pgtype.UUID` (the mapper in `activity_mapper.go` assumes the former).

## Open design issues (full detail in `SCHEMA_REVIEW.md`)

- [ ] Race condition: concurrent schedule edits (no locking/versioning yet).
- [ ] `is_fixed_schedule` can drift out of sync with the contents of `activity_schedules` (no enforcement, purely an app-layer discipline).
- [ ] Changing `confirmation_timeout_minutes` applies retroactively to old `awaiting` items — decide whether this is the intended behavior before implementing the timeout worker.
- [ ] Soft-deleting a user doesn't automatically revoke their `user_sessions` — needs `DeleteAllSessionsByUserID` called explicitly as part of the `SoftDeleteUser` flow.
- [ ] Image files (`activity_images`/`activity_capture_images`) become orphaned after a hard delete — no cleanup job yet.
- [ ] `user_verifications` is vulnerable to replay (no `consumed_at` / row-lock during validation).
- [ ] `activity_captures.status` transitions aren't restricted (a `missed` item could go back to `captured`) — decide whether late confirmation should actually be supported.

## Code tidiness (not urgent yet, but worth watching for these thresholds)

- [ ] Split `router.go` per domain (`activity_routes.go`, etc.) — in progress.
- [ ] `activity_usecase.go` — once it exceeds ~5-6 methods or ~250-300 lines, consider splitting per operation (`activity_usecase_create.go`, etc.), keeping the interface in one main file.
- [ ] `activity_dto.go` — keep one file per entity (not per operation) as long as its size stays reasonable.
