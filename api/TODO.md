# TODO

Remaining work, ordered by priority (security-critical first, product-blocking next, then completeness, ops, open design decisions, and low-urgency housekeeping last).

## Tier 1 — Security

- [ ] Content-type validation on image uploads is header-only right now (`fileHeader.Header.Get("Content-Type")`, client-supplied and spoofable) — validate actual file content (magic bytes) before trusting it's really an image. Flagged inline in `thread_image_handler.go`.
- [ ] Rate limiting / lockout on `Login` — nothing currently stops repeated failed-password attempts against one account or IP.
- [ ] CORS — no middleware configured yet; will block `web/` from calling the API cross-origin once it starts doing so.

## Tier 2 — Core product features not started

- [ ] **Scheduler worker** — a job that evaluates cron expressions from `commitments` and auto-generates `moments`. The `ListCommitmentsForGeneration` query exists, but there's no worker actually parsing cron and calling `CreateCommittedMoment`.
- [ ] **Timeout worker** — a periodic job calling `MarkOverdueMomentsAsMissed` (see the retroactive-timeout decision under Tier 5 before implementing this).
- [ ] **Statistics endpoints** — heatmap & Time to Tell chart. The `GetHeatmapData` / `GetSettlingTimeStats` queries exist, but there's no usecase/handler yet.
- [ ] `RestoreThread` — usecase + handler. The sqlc query already exists (`internal/db/threads.sql.go`), nothing calls it yet.
- [ ] `moment_images` upload — needs the identical repository/usecase/handler pattern already used for `thread_images`.

## Tier 3 — Auth feature completeness

- [ ] **OAuth (Google/GitHub)** — not implemented, only the `credential` provider path exists. `user_accounts.provider_id` and the schema already support it; no usecase/handler yet.
- [ ] **Email verification** — not implemented. `SetUserEmailVerified` / `user_verifications` queries exist but nothing calls them yet.
- [ ] **Password reset** — not implemented.

## Tier 4 — Operational readiness

- [ ] Health check endpoint (`/healthz`) for container orchestration.
- [ ] `Dockerfile` + a deployment `docker-compose.yml` (app + Postgres) — `api/docker-compose.dev.yaml` only covers local Postgres/RustFS for `make dev`, not shipping the app itself.
- [ ] CI: check mocks aren't stale — run `mockery` then `git diff --exit-code internal/usecase/mocks/` as a step in `.github/workflows/api-test.yaml`, failing if someone forgot to regenerate after changing an interface.
- [ ] CI: `PREPARE` every generated query against a real Postgres. `sqlc generate` validates table/column names but **not operators**, which is how a broken `||` (written as `| |`) survived in `ListDueMomentsPastDeadline` and `MarkOverdueMomentsAsMissed` — both would have failed at runtime, and neither is covered by a test because the workers that call them don't exist yet. Extracting the `const` SQL from `internal/db/*.sql.go` and running `PREPARE` on each against the migrated schema catches this class outright.

## Tier 5 — Open design decisions (full detail in `SCHEMA_REVIEW.md`)

### Schema/spec gaps opened by the vocabulary rename

Migration `000003` renamed the schema to the `FEATURES.md` vocabulary but
deliberately changed no semantics. These are the three places where the schema
and the spec now disagree, and each needs a model change rather than a rename:

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

## Tier 6 — Code tidiness (not urgent yet, watch for these thresholds)

- [ ] Split `router.go` per domain (`thread_routes.go`, etc.) once it outgrows a handful of domains — currently small enough to stay as one file.
- [ ] `thread_usecase.go` — once it exceeds ~5-6 methods or ~250-300 lines, consider splitting per operation (`thread_usecase_create.go`, etc.), keeping the interface in one main file.
- [ ] `thread_dto.go` — keep one file per entity (not per operation) as long as its size stays reasonable.
