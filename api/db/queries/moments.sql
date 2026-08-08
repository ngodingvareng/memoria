-- name: CreateMoment :one
-- The manual/import capture path — origin distinguishes the two, but
-- both are ordinary, fully-editable Moments with no state (FEATURES.md:
-- "A manually captured Moment has no state"). Commitment-generated
-- Moments are created from commitment_occurrences instead, one
-- migration later, and never go through this query.
--
-- ON CONFLICT mirrors uq_moments_user_client_id: a Moment captured
-- offline and synced twice (e.g. a flaky retry) is silently ignored
-- instead of erroring, so capture never shows a failure state for
-- something already recorded.
INSERT INTO moments(
    user_id, thread_id, origin,
    occurred_at, occurred_local, occurred_utc_offset_minutes,
    recorded_at,
    note, color_hex, place_name, latitude, longitude,
    client_id
) VALUES (
    sqlc.arg(user_id), sqlc.narg(thread_id), sqlc.arg(origin),
    sqlc.arg(occurred_at), sqlc.arg(occurred_local), sqlc.arg(occurred_utc_offset_minutes),
    COALESCE(sqlc.narg(recorded_at), NOW()),
    sqlc.narg(note), sqlc.narg(color_hex), sqlc.narg(place_name), sqlc.narg(latitude), sqlc.narg(longitude),
    sqlc.narg(client_id)
)
ON CONFLICT (user_id, client_id) WHERE client_id IS NOT NULL DO NOTHING
RETURNING *;

-- name: GetMomentByUserAndClientID :one
-- Backs the idempotent-retry path for CreateMoment: when the ON
-- CONFLICT ... DO NOTHING above skips a duplicate offline-sync insert,
-- RETURNING produces no row, and the caller re-fetches the already-
-- persisted Moment through this query instead of surfacing an error.
SELECT *
FROM moments
WHERE user_id = sqlc.arg(user_id)
    AND client_id = sqlc.arg(client_id)
    AND deleted_at IS NULL;

-- name: GetMomentByID :one
-- user_id in the WHERE clause doubles as an ownership check. A Moment
-- reached via a mention or a Circle share is authorized through
-- moment_mentions/moment_circles instead — see those query files.
SELECT *
FROM moments
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: ListMomentsByUserID :many
-- The personal archive timeline: every Moment this user owns, newest
-- occurrence first.
SELECT *
FROM moments
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
ORDER BY occurred_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: ListMomentsByThreadID :many
-- A single Thread's browsable timeline. No user_id filter: a
-- collaborative Thread holds Moments captured by several members, and
-- access to the Thread itself is checked upstream (see
-- threads.GetThreadWithAccess).
SELECT *
FROM moments
WHERE thread_id = sqlc.arg(thread_id)
    AND deleted_at IS NULL
ORDER BY occurred_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: SearchMoments :many
-- Global text Search (FEATURES.md, Looking Back) over one user's own
-- Moments, backed by idx_moments_search_document.
SELECT *
FROM moments
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    AND search_document @@ plainto_tsquery('simple', sqlc.arg(query))
ORDER BY occurred_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: ListMomentsOnDates :many
-- Echoes same-day-and-month flashback. The caller computes the IN-list
-- of prior-year occurred_on dates for "today" (e.g. every August 4th
-- going back); matching against occurred_on lets
-- idx_moments_user_occurred_on serve this directly instead of an
-- expression index on extract(month/day) — see that index's comment.
SELECT *
FROM moments
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    AND occurred_on = ANY(sqlc.arg(occurred_on_dates)::date[])
ORDER BY occurred_on DESC;

-- name: UpdateMoment :one
-- A Moment is never final (FEATURES.md, "Moments are never final") —
-- note, color, location, and even occurred_at may be edited at any
-- point, including years later; occurred_at is the one timestamp the
-- user is the authority on (see the Time table). recorded_at is never
-- touched here: it is stamped once, at capture, and stays fixed.
UPDATE moments
SET thread_id = sqlc.narg(thread_id),
    occurred_at = sqlc.arg(occurred_at),
    occurred_local = sqlc.arg(occurred_local),
    occurred_utc_offset_minutes = sqlc.arg(occurred_utc_offset_minutes),
    note = sqlc.narg(note),
    color_hex = sqlc.narg(color_hex),
    place_name = sqlc.narg(place_name),
    latitude = sqlc.narg(latitude),
    longitude = sqlc.narg(longitude),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    RETURNING *;

-- name: TouchMomentLastViewed :exec
-- Owner-only, per the last_viewed_at column comment: this is not a read
-- receipt and never runs for any viewer other than moments.user_id.
UPDATE moments
SET last_viewed_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: SoftDeleteMoment :exec
-- Per FEATURES.md Lifecycle & Deletion, the caller must also remove
-- this Moment's shares/mentions/comments/reactions in the same
-- transaction — deleting a Moment removes it everywhere, not just from
-- the owner's own archive.
UPDATE moments
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: RestoreMoment :exec
UPDATE moments
SET deleted_at = NULL
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);

-- name: GetHeatmapData :many
-- GitHub-contribution-graph-style data for Rhythms: Moment count per
-- calendar day. Buckets by occurred_on — the local date where each
-- Moment happened — never by the viewer's current timezone, per the
-- occurred_on column comment in the archive schema migration.
SELECT
    occurred_on AS day,
    COUNT(*)::bigint AS moment_count
FROM moments
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    AND occurred_on >= sqlc.arg(from_date)
    AND occurred_on < sqlc.arg(to_date)
GROUP BY occurred_on
ORDER BY occurred_on;

-- name: GetMomentWithAccess :one
-- "Can this viewer reach this Moment at all" — true for the owner, an
-- active (non-removed) mentioned user, an active member of a Circle it
-- was shared into, or an active member of the Circle that natively owns
-- its Thread. Mirrors threads.GetThreadWithAccess. Used by Mention and
-- Response, where the actor is frequently not the owner — GetMomentByID
-- deliberately doesn't cover that (see its own comment).
SELECT moments.*
FROM moments
WHERE moments.id = sqlc.arg(id)
    AND moments.deleted_at IS NULL
    AND (
        moments.user_id = sqlc.arg(user_id)
        OR EXISTS (
            SELECT 1 FROM moment_mentions
            WHERE moment_id = moments.id
                AND mentioned_user_id = sqlc.arg(user_id)
                AND removed_at IS NULL
        )
        OR EXISTS (
            SELECT 1 FROM moment_circles mc
                JOIN circle_members cm ON cm.circle_id = mc.circle_id
            WHERE mc.moment_id = moments.id
                AND cm.user_id = sqlc.arg(user_id)
                AND cm.left_at IS NULL
        )
        OR (
            moments.thread_id IS NOT NULL
            AND EXISTS (
                SELECT 1 FROM threads t
                    JOIN circle_members cm ON cm.circle_id = t.circle_id
                WHERE t.id = moments.thread_id
                    AND cm.user_id = sqlc.arg(user_id)
                    AND cm.left_at IS NULL
            )
        )
    );

-- name: ListMomentVisibleCircleIDs :many
-- Which Circle audiences sqlc.arg(user_id) actually shares with this
-- Moment: the intersection of {its native Thread's Circle, every Circle
-- it's been shared into} with the Circles the viewer is currently an
-- active member of. Backs both authorizing a specific circle_id on
-- Comment/Reaction create, and scoping ListComments/ListReactions.
WITH candidate_circles AS (
    SELECT t.circle_id AS circle_id
    FROM moments m
        JOIN threads t ON t.id = m.thread_id
    WHERE m.id = sqlc.arg(moment_id)
        AND t.circle_id IS NOT NULL
    UNION
    SELECT mc.circle_id AS circle_id
    FROM moment_circles mc
    WHERE mc.moment_id = sqlc.arg(moment_id)
)
SELECT DISTINCT cc.circle_id
FROM candidate_circles cc
    JOIN circle_members cm ON cm.circle_id = cc.circle_id
WHERE cm.user_id = sqlc.arg(user_id)
    AND cm.left_at IS NULL;

-- name: GetSettlingTimeStats :many
-- Raw data for the Time to Tell chart (FEATURES.md, Rhythms): how long
-- a Moment took to settle into a record. Scoped to origin = 'manual'
-- because Time to Tell is explicitly a manual-Moment stat and must
-- never be merged with, or presented alongside, Commitment adherence
-- (a separate punctuality measurement that will live on
-- commitment_occurrences).
SELECT
    id,
    occurred_at,
    recorded_at,
    settling_time
FROM moments
WHERE user_id = sqlc.arg(user_id)
    AND origin = 'manual'
    AND deleted_at IS NULL
ORDER BY occurred_at;
