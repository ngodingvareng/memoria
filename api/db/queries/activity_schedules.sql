-- name: CreateActivitySchedule :one
INSERT INTO activity_schedules(activity_id, cron_expression, timezone)
VALUES (sqlc.arg(activity_id), sqlc.arg(cron_expression), sqlc.narg(timezone))
RETURNING *;

-- name: GetActivityScheduleByID :one
SELECT *
FROM activity_schedules
WHERE id = sqlc.arg(id) AND activity_id = sqlc.arg(activity_id);

-- name: ListActiveSchedulesByActivityID :many
-- Can return more than one row: an activity may have multiple
-- concurrently active schedules (e.g. "pagi" and "sore").
SELECT *
FROM activity_schedules
WHERE activity_id = sqlc.arg(activity_id)
ORDER BY created_at;

-- name: UpdateActivitySchedule :one
-- Update in place (same id) so existing activity_items.schedule_id links
-- stay valid. The caller must insert a row into
-- activity_schedule_histories with the OLD cron_expression/timezone
-- BEFORE calling this, in the same transaction.
UPDATE activity_schedules
SET cron_expression = sqlc.arg(cron_expression),
    timezone = sqlc.narg(timezone)
WHERE id = sqlc.arg(id) AND activity_id = sqlc.arg(activity_id)
    RETURNING *;

-- name: DeleteActivitySchedule :exec
-- Retire a schedule entirely. The caller should archive it into
-- activity_schedule_histories first. Existing activity_items that
-- referenced it keep activity_id intact and just get schedule_id set to
-- NULL (composite FK, ON DELETE SET NULL (schedule_id)).
DELETE FROM activity_schedules
WHERE id = sqlc.arg(id) AND activity_id = sqlc.arg(activity_id);

-- name: ListActiveSchedulesForGeneration :many
-- Feed for the scheduler worker: every active schedule belonging to a
-- non-deleted, is_fixed_schedule = true activity. The worker evaluates
-- cron_expression/timezone itself and calls
-- activity_items.CreateScheduledActivityItem when a slot fires.
SELECT activity_schedules.* FROM activity_schedules
    JOIN activities ON activities.id = activity_schedules.activity_id
WHERE activities.deleted_at IS NULL AND activities.is_fixed_schedule = TRUE;