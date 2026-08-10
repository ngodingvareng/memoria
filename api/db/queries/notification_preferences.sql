-- name: GetNotificationPreferences :one
SELECT *
FROM notification_preferences
WHERE user_id = sqlc.arg(user_id);

-- name: CreateDefaultNotificationPreferences :one
-- Lazily inserts the all-defaults row the first time a user's
-- preferences are read or updated, so "no row yet" and "row with every
-- default value" are the same state and a missing row can never silently
-- mean "off" (see the migration's own header comment on this table).
INSERT INTO notification_preferences(user_id)
VALUES (sqlc.arg(user_id))
ON CONFLICT (user_id) DO NOTHING
RETURNING *;

-- name: UpdateNotificationPreferences :one
UPDATE notification_preferences
SET commitment_upcoming_enabled = sqlc.arg(commitment_upcoming_enabled),
    moment_due_enabled = sqlc.arg(moment_due_enabled),
    moment_missed_enabled = sqlc.arg(moment_missed_enabled),
    echo_ready_enabled = sqlc.arg(echo_ready_enabled),
    recap_generated_enabled = sqlc.arg(recap_generated_enabled),
    circle_invite_received_enabled = sqlc.arg(circle_invite_received_enabled),
    circle_join_request_received_enabled = sqlc.arg(circle_join_request_received_enabled),
    mentioned_in_moment_enabled = sqlc.arg(mentioned_in_moment_enabled),
    added_to_circle_enabled = sqlc.arg(added_to_circle_enabled),
    circle_join_request_approved_enabled = sqlc.arg(circle_join_request_approved_enabled),
    response_digest_enabled = sqlc.arg(response_digest_enabled),
    response_digest_hour = sqlc.arg(response_digest_hour),
    quiet_hours_start = sqlc.narg(quiet_hours_start),
    quiet_hours_end = sqlc.narg(quiet_hours_end),
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
RETURNING *;
