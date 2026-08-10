-- name: AddMomentImage :one
INSERT INTO moment_images(
    moment_id, image_path, image_alt,
    content_type, byte_size, width, height, metadata_stripped
) VALUES (
    sqlc.arg(moment_id), sqlc.arg(image_path), sqlc.narg(image_alt),
    sqlc.narg(content_type), sqlc.narg(byte_size), sqlc.narg(width), sqlc.narg(height),
    sqlc.arg(metadata_stripped)
)
RETURNING *;

-- name: ListMomentImagesByMomentID :many
SELECT *
FROM moment_images
WHERE moment_id = sqlc.arg(moment_id)
ORDER BY sort_order, created_at;

-- name: DeleteMomentImage :one
-- Returns the deleted row (specifically image_path) so the caller can
-- remove the matching storage object after the DB record is gone — see
-- MomentImageUsecase.DeleteMomentImage for the ordering rationale. No
-- matching row (already deleted / wrong moment) surfaces as
-- pgx.ErrNoRows, which the repository treats as a silent no-op, same as
-- the :exec version this replaces.
DELETE FROM moment_images
WHERE id = sqlc.arg(id)
    AND moment_id = sqlc.arg(moment_id)
RETURNING *;

-- name: ListAlbumImages :many
-- The Personal Album (FEATURES.md, Looking Back): every photo
-- attachment from every Moment this user owns, newest occurrence
-- first — "the Album is exactly this table [moment_images] joined
-- back to moments for ordering and visibility" (see that table's own
-- comment above). occurred_local/occurred_utc_offset_minutes are
-- returned instead of occurred_at so the caller can bucket by the
-- Moment's local calendar day, not a UTC one.
--
-- A missed Commitment occurrence produces no moments row at all, so it
-- can produce no moment_images row either — the Commitment Firewall
-- holds here structurally, no extra predicate needed. Resurfacing
-- exclusions (resurfacing_exclusions) are deliberately NOT filtered
-- here either: FEATURES.md limits that mechanism to *automatic*
-- surfacing (Echoes, Recap, notifications, algorithmic Album
-- groupings) and says excluded content "stays fully in the archive,
-- searchable and browsable" — this is plain chronological browsing,
-- never algorithmic.
SELECT
    moment_images.id,
    moment_images.moment_id,
    moment_images.image_path,
    moment_images.image_alt,
    moment_images.width,
    moment_images.height,
    moments.occurred_local,
    moments.occurred_utc_offset_minutes
FROM moment_images
    JOIN moments ON moments.id = moment_images.moment_id
WHERE moments.user_id = sqlc.arg(user_id)
    AND moments.deleted_at IS NULL
ORDER BY moments.occurred_at DESC, moment_images.sort_order ASC, moment_images.id ASC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: ListCircleAlbumImages :many
-- The Circle's Album (FEATURES.md, Looking Back): every photo
-- attachment from every Moment visible in this Circle — both its own
-- collaborative Threads' Moments and personal Moments shared into it
-- via mentioning. Mirrors ListCircleMoments' two-source union, but
-- additionally carries an is_shared discriminator plus the sharer's
-- user id/username, so the caller can render the "from @username"
-- attribution label the spec requires on shared-in images only.
--
-- native_circle_moments / shared_circle_moments are split into two
-- CTEs (rather than one UNION like ListCircleMoments) specifically so
-- the shared branch can NOT EXISTS-exclude anything already native:
-- ShareToCircle only checks moment ownership + circle membership, not
-- "unless this moment is already native to this circle" — so a Moment
-- captured directly into the Circle's own Thread could also pick up a
-- moment_circles row for that same Circle. A plain UNION would not
-- collapse that duplicate, because is_shared/shared_by_user_id differ
-- between the two candidate rows for the same Moment. Native wins
-- (unattributed) over shared.
--
-- Firewall + resurfacing-exclusion reasoning: identical to
-- ListAlbumImages above. Caller is responsible for the Circle
-- membership check before calling this query.
WITH native_circle_moments AS (
    SELECT
        m.id,
        m.occurred_at,
        m.occurred_local,
        m.occurred_utc_offset_minutes
    FROM moments m
        JOIN threads t ON t.id = m.thread_id
    WHERE t.circle_id = sqlc.arg(circle_id)
        AND m.deleted_at IS NULL
),
shared_circle_moments AS (
    SELECT
        m.id,
        m.occurred_at,
        m.occurred_local,
        m.occurred_utc_offset_minutes,
        mc.shared_by_user_id
    FROM moment_circles mc
        JOIN moments m ON m.id = mc.moment_id
    WHERE mc.circle_id = sqlc.arg(circle_id)
        AND m.deleted_at IS NULL
        AND NOT EXISTS (
            SELECT 1 FROM native_circle_moments ncm WHERE ncm.id = m.id
        )
),
circle_moments AS (
    SELECT
        id, occurred_at, occurred_local, occurred_utc_offset_minutes,
        FALSE AS is_shared, NULL::uuid AS shared_by_user_id
    FROM native_circle_moments
    UNION ALL
    SELECT
        id, occurred_at, occurred_local, occurred_utc_offset_minutes,
        TRUE AS is_shared, shared_by_user_id
    FROM shared_circle_moments
)
SELECT
    moment_images.id,
    moment_images.moment_id,
    moment_images.image_path,
    moment_images.image_alt,
    moment_images.width,
    moment_images.height,
    circle_moments.occurred_local,
    circle_moments.occurred_utc_offset_minutes,
    circle_moments.is_shared,
    circle_moments.shared_by_user_id,
    users.username AS shared_by_username
FROM moment_images
    JOIN circle_moments ON circle_moments.id = moment_images.moment_id
    LEFT JOIN users ON users.id = circle_moments.shared_by_user_id
ORDER BY circle_moments.occurred_at DESC, moment_images.sort_order ASC, moment_images.id ASC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);
