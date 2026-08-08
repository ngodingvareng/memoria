-- name: CreateUserBlock :exec
-- Idempotent: blocking someone already blocked is a no-op, not an error.
INSERT INTO user_blocks(blocker_user_id, blocked_user_id)
VALUES (sqlc.arg(blocker_user_id), sqlc.arg(blocked_user_id))
ON CONFLICT (blocker_user_id, blocked_user_id) DO NOTHING;

-- name: DeleteUserBlock :exec
DELETE FROM user_blocks
WHERE blocker_user_id = sqlc.arg(blocker_user_id)
    AND blocked_user_id = sqlc.arg(blocked_user_id);

-- name: ListBlockedUsersByBlockerID :many
-- "Blocked users" settings surface.
SELECT *
FROM user_blocks
WHERE blocker_user_id = sqlc.arg(blocker_user_id)
ORDER BY created_at DESC;

-- name: IsBlockedEitherDirection :one
-- Blocking is symmetric in effect (see the migration's header comment
-- on user_blocks): this must gate every view/mention/comment/reaction
-- access check, tested both ways since only the blocker's row exists.
SELECT EXISTS(
    SELECT 1
    FROM user_blocks
    WHERE (blocker_user_id = sqlc.arg(user_a) AND blocked_user_id = sqlc.arg(user_b))
        OR (blocker_user_id = sqlc.arg(user_b) AND blocked_user_id = sqlc.arg(user_a))
) AS blocked;
