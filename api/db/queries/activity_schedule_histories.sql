-- name: CreateActivityScheduleHistory :one
-- Archive a schedule's values before it's changed or retired.
INSERT INTO activity_schedule_histories(
    activity_id,
    schedule_id,
    cron_expression,
    timezone,
    active_from,
    active_until
)
VALUES (
    sqlc.arg(activity_id),
    sqlc.narg(schedule_id),
    sqlc.arg(cron_expression),
    sqlc.narg(timezone),
    sqlc.arg(active_from),
    sqlc.arg(active_until)
)
RETURNING *;

-- name: ListScheduleHistoryByActivityID :many
-- "How has this activity's schedule changed over time" view.
SELECT * FROM activity_schedule_histories
WHERE activity_id = sqlc.arg(activity_id)
ORDER BY active_from DESC;