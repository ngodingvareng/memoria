# TODO

Remaining work, ordered by priority (build-breaking first, core product features next, then auth/session completeness, ops, open design decisions, and low-urgency housekeeping last).

## Tier 1 — Build-breaking bugs

- [x] ~~`bun run build` failed outright (`tsc -b` errors before `vite build` ran)~~ — done: `next-themes`/`sonner` installed (`src/components/ui/sonner.tsx` is a fully-built Toaster wrapper, not scaffolding — kept it rather than deleting); `MomentList`/`MomentCardParam`/`MomentInput` imports in `$username.tsx` and `thread.$id.index.tsx` now point at `@/features/moments` instead of `@/features/threads`; `_circle.tsx`'s tab links fixed from `/g/$id...` to `/c/$id...`; unused `FavouriteIcon`/`LinkForwardIcon` imports dropped from `moment-card.tsx`; `thread-header.tsx`'s `onShare` prop was genuinely wired by its caller (`thread.$id.index.tsx` opens a real `ShareDialog` through it) but had no button rendering it — added one instead of just silencing the unused-var error.

## Tier 2 — Core product wired to real API

`web/` now has the pattern for this (Orval-generated types/hooks + TanStack Query, see the auth flow: `src/lib/api/generated/auth`, `src/lib/api/mutator.ts`) — every other domain is still rendering `src/lib/dummies.tsx` or inline mock data (`dummyThreadStories` in `$username.tsx`, the hardcoded "336k threads", "Confirmations 19", "Moment Today" cards on the home route, etc.). The corresponding `api/` endpoints exist for Thread/Moment CRUD per `api/TODO.md`; Circles and the rest need checking against `api/TODO.md` per-endpoint before wiring.

- [ ] **Threads** — `thread.index.tsx`, `thread.new.tsx`, `thread.$id.index.tsx`, `thread.$id.info.tsx`.
- [ ] **Moments** — `moment.new.tsx`, `thread.$id.moments.$momentId.tsx`.
- [ ] **Circles** — `circle.index.tsx`, `circle.new.tsx`, `_circle/c.$id.*.tsx` (index, thread, member, settings, invite, join).
- [ ] **Home / profile** — `_app/index.tsx` and `_app/$username.tsx`: heatmap, Confirmations, Threads, Moment Today, Last Moments — all hardcoded right now.
- [ ] **Mentions, Notifications, Album, Search, Recap, Rhythms (stats)** — still scaffolding/dummy per their route files; check `api/TODO.md` first since the statistics endpoints (heatmap/Time to Tell) aren't implemented backend-side yet either.
- [ ] Delete `src/lib/dummies.tsx` once nothing imports from it.

## Tier 3 — Auth/session completeness

- [ ] `signin-form.tsx`'s "Login with Apple"/"Login with Google" buttons are decorative (no handler) — matches `api/TODO.md` Tier 3, OAuth isn't implemented backend-side either.
- [ ] "Forgot password?" link on `signin-form.tsx` points to `/signup` — no forgot-password flow exists on either side yet.
- [ ] Proactive silent refresh — `LoginResponse.expires_in` exists specifically so the frontend can schedule a refresh shortly before the access token expires (see the Go doc comment on `dto.LoginResponse.ExpiresIn`). Right now the app only reacts to a 401 after the fact (`src/lib/api/mutator.ts`). Deliberately deferred when the TanStack Query/Orval work landed — reactive refresh + bootstrap-refresh-on-reload already cover correctness; this is a latency nicety, not a bug.
- [ ] No avatar upload/display — `UserResponse` has no avatar field on the backend, and `user-header-menu.tsx` shows the same hardcoded placeholder image regardless of which user is signed in.

## Tier 4 — Operational readiness

- [ ] No test runner configured at all (per root `CLAUDE.md`) — `web/` has zero automated tests. `web-test.yaml` only does a generated-API-client drift check + build; nothing exercises actual app behavior.
- [ ] No lint/format CI check — `bun run lint` (oxlint) and `bun run format:check` (prettier) exist as scripts but aren't wired into any workflow yet.
- [ ] No centralized error/toast strategy — every form repeats the same `ApiError` → message mapping in its own `submitError` state (`signin-form.tsx`, `signup-form.tsx`, `welcome-form.tsx`). Fine at three forms; worth a shared helper once Tier 2 adds more mutations.
- [ ] `api/`'s `swagger.json` freshness isn't checked anywhere — nothing catches a Go handler's swag annotations drifting from the committed `docs/swagger.json` (a `check-swagger` job mirroring `check-mocks`, on the `api/` side, regenerating via `swag init` and diffing `docs/`). `web-test.yaml`'s drift check only verifies generated TS is in sync with whatever `swagger.json` currently says, not that `swagger.json` itself is accurate.

## Tier 5 — Open UX/design decisions

- [ ] "Display Language: English" submenu in `user-header-menu.tsx` is fully decorative — the English/Indonesia options don't do anything (no i18n library chosen yet, unlike the adjacent theme submenu which actually calls `setTheme`).
- [ ] `mobile/` (Flutter) has its own hand-rolled `core/network/api_client.dart`, entirely separate from the Orval-generated client here — decide whether/how to keep the two API contracts in sync, or accept the duplication given the different toolchains.

## Tier 6 — Code tidiness

- [ ] `src/lib/api/generated/` is ~90 files from a single `bun run generate:api` run, almost entirely for domains (circles, moments, threads, comments, reactions, mentions) not wired to any UI yet (Tier 2). Expect this directory to keep growing — not something to prune, just don't be surprised by diff size.
- [ ] Several `githubComNgodingvarengMemoria...WebResponse...` files under `src/lib/api/generated/models/` are dead/unused — `src/lib/api/transformer.ts` collapses each `WebResponse<T>` definition down to `T` in place, but Orval still emits one model file per original definition name. Harmless and tree-shaken; not worth fighting further (see the comment in `orval.config.ts`).
