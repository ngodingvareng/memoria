-- name: CreateActivity :one
INSERT INTO activities(
    user_id, name, description, is_fixed_schedule,
    color_palette, confirmation_timeout_minutes
) VALUES (
    sqlc.arg(user_id), sqlc.arg(name), sqlc.narg(description), sqlc.arg(is_fixed_schedule),
    sqlc.narg(color_palette), sqlc.arg(confirmation_timeout_minutes)
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

-- name: UpdateActivity :one
UPDATE activities
SET name = sqlc.arg(name),
    description = sqlc.narg(description),
    color_palette = sqlc.narg(color_palette),
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