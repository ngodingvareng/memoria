-- name: CreateThread :one
INSERT INTO threads(
    user_id, name, description, has_commitment,
    color_hex, confirmation_timeout_minutes
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name), sqlc.narg(description), sqlc.arg(has_commitment),
    sqlc.narg(color_hex), sqlc.arg(confirmation_timeout_minutes)
)
RETURNING *;

-- name: GetThreadByID :one
-- user_id in the WHERE clause doubles as an ownership check.
SELECT *
FROM threads
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: ListThreadsByUserID :many
SELECT *
FROM threads
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: SearchThreads :many
-- Search & filter threads for the authenticated user's thread list.
-- All filters are optional (NULL = "don't filter on this"):
--   - name: case-insensitive partial match (ILIKE) against threads.name
--   - has_commitment: exact boolean match
-- Ordered newest-first; paginated via page_limit / page_offset.
SELECT *
FROM threads
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    AND (
        sqlc.narg(name)::text IS NULL
        OR name ILIKE '%' || sqlc.narg(name)::text || '%'
    )
    AND (
        sqlc.narg(has_commitment)::bool IS NULL
        OR has_commitment = sqlc.narg(has_commitment)::bool
    )
ORDER BY created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountSearchThreads :one
-- Total matching row count for SearchThreads, ignoring
-- page_limit/page_offset — used to compute pagination metadata.
SELECT COUNT(*)
FROM threads
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    AND (
        sqlc.narg(name)::text IS NULL
        OR name ILIKE '%' || sqlc.narg(name)::text || '%'
    )
    AND (
        sqlc.narg(has_commitment)::bool IS NULL
        OR has_commitment = sqlc.narg(has_commitment)::bool
    );

-- name: UpdateThread :one
UPDATE threads
SET name = sqlc.arg(name),
    description = sqlc.narg(description),
    color_hex = sqlc.narg(color_hex),
    confirmation_timeout_minutes = sqlc.arg(confirmation_timeout_minutes),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    RETURNING *;

-- name: SetThreadHasCommitment :exec
-- Toggling this does NOT itself touch commitments — if flipping to
-- false, the app should separately retire/archive existing commitments
-- (see commitments.sql) so the generator job actually stops.
UPDATE threads
SET has_commitment = sqlc.arg(has_commitment),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);

-- name: SoftDeleteThread :exec
UPDATE threads
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: RestoreThread :exec
UPDATE threads
SET deleted_at = NULL
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);

-- name: HardDeleteThread :exec
DELETE FROM threads
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);
