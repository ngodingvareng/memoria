-- name: CreateActivity :one
INSERT INTO activities(
    user_id, name, description, is_fixed_schedule,
    color_hex, confirmation_timeout_minutes
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name), sqlc.narg(description), sqlc.arg(is_fixed_schedule),
    sqlc.narg(color_hex), sqlc.arg(confirmation_timeout_minutes)
)
RETURNING *;

-- name: GetActivityByID :one
-- user_id in the WHERE clause doubles as an ownership check.
SELECT *
FROM activities
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: ListActivitiesByUserID :many
SELECT *
FROM activities
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: SearchActivities :many
-- Search & filter activities for the authenticated user's activity list.
-- All filters are optional (NULL = "don't filter on this"):
--   - name: case-insensitive partial match (ILIKE) against activities.name
--   - is_fixed_schedule: exact boolean match
-- Ordered newest-first; paginated via page_limit / page_offset.
SELECT *
FROM activities
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    AND (
        sqlc.narg(name)::text IS NULL
        OR name ILIKE '%' || sqlc.narg(name)::text || '%'
    )
    AND (
        sqlc.narg(is_fixed_schedule)::bool IS NULL
        OR is_fixed_schedule = sqlc.narg(is_fixed_schedule)::bool
    )
ORDER BY created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountSearchActivities :one
-- Total matching row count for SearchActivities, ignoring
-- page_limit/page_offset — used to compute pagination metadata.
SELECT COUNT(*)
FROM activities
WHERE user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    AND (
        sqlc.narg(name)::text IS NULL
        OR name ILIKE '%' || sqlc.narg(name)::text || '%'
    )
    AND (
        sqlc.narg(is_fixed_schedule)::bool IS NULL
        OR is_fixed_schedule = sqlc.narg(is_fixed_schedule)::bool
    );

-- name: UpdateActivity :one
UPDATE activities
SET name = sqlc.arg(name),
    description = sqlc.narg(description),
    color_hex = sqlc.narg(color_hex),
    confirmation_timeout_minutes = sqlc.arg(confirmation_timeout_minutes),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    RETURNING *;

-- name: SetActivityFixedSchedule :exec
-- Toggling this does NOT itself touch activity_schedules — if flipping
-- to false, the app should separately retire/archive existing schedules
-- (see activity_schedules.sql) so the generator job actually stops.
UPDATE activities
SET is_fixed_schedule = sqlc.arg(is_fixed_schedule),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);

-- name: SoftDeleteActivity :exec
UPDATE activities
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: RestoreActivity :exec
UPDATE activities
SET deleted_at = NULL
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);

-- name: HardDeleteActivity :exec
DELETE FROM activities
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);