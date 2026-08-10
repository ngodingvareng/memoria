-- name: CreateNotification :one
-- One row per deliverable alert. All subject columns are nullable and the
-- caller only ever populates the one(s) relevant to kind — see
-- entity.Notification's own field comments for which subject goes with
-- which kind. deliver_after defaults to NOW() unless the caller is
-- honoring quiet hours, in which case it's pushed forward by the usecase
-- before this query runs.
INSERT INTO notifications(
    user_id, kind, actor_user_id,
    moment_id, thread_id, commitment_id, commitment_occurrence_id,
    circle_id, circle_invite_id, circle_join_request_id,
    recap_id, resurfacing_id,
    payload, deliver_after
) VALUES (
    sqlc.arg(user_id), sqlc.arg(kind), sqlc.narg(actor_user_id),
    sqlc.narg(moment_id), sqlc.narg(thread_id), sqlc.narg(commitment_id), sqlc.narg(commitment_occurrence_id),
    sqlc.narg(circle_id), sqlc.narg(circle_invite_id), sqlc.narg(circle_join_request_id),
    sqlc.narg(recap_id), sqlc.narg(resurfacing_id),
    sqlc.arg(payload), sqlc.arg(deliver_after)
)
RETURNING *;

-- name: ListNotificationsByUserID :many
-- The in-app notification feed: everything past its deliver_after (quiet
-- hours held rows are simply not yet visible), newest first. Uses
-- idx_notifications_user_created_at.
SELECT *
FROM notifications
WHERE user_id = sqlc.arg(user_id)
    AND deliver_after <= NOW()
ORDER BY created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountUnreadNotifications :one
-- Backs the header bell badge.
SELECT COUNT(*)::bigint
FROM notifications
WHERE user_id = sqlc.arg(user_id)
    AND deliver_after <= NOW()
    AND read_at IS NULL;

-- name: CountNotifications :one
-- Total visible rows, for ListNotifications' pagination footer.
SELECT COUNT(*)::bigint
FROM notifications
WHERE user_id = sqlc.arg(user_id)
    AND deliver_after <= NOW();

-- name: MarkNotificationRead :exec
-- Owner-scoped; a WHERE clause matching zero rows (already read, wrong
-- owner, or not found) is a silent no-op.
UPDATE notifications
SET read_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND read_at IS NULL;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = NOW()
WHERE user_id = sqlc.arg(user_id)
    AND deliver_after <= NOW()
    AND read_at IS NULL;
