-- name: CreateUserMute :exec
-- Idempotent: muting someone already muted is a no-op, not an error.
INSERT INTO user_mutes(muter_user_id, muted_user_id)
VALUES (sqlc.arg(muter_user_id), sqlc.arg(muted_user_id))
ON CONFLICT (muter_user_id, muted_user_id) DO NOTHING;

-- name: DeleteUserMute :exec
DELETE FROM user_mutes
WHERE muter_user_id = sqlc.arg(muter_user_id)
    AND muted_user_id = sqlc.arg(muted_user_id);

-- name: ListMutedUsersByMuterID :many
-- "Muted users" settings surface.
SELECT *
FROM user_mutes
WHERE muter_user_id = sqlc.arg(muter_user_id)
ORDER BY created_at DESC;

-- name: IsUserMuted :one
-- Muting is one-directional and purely a view filter — only the
-- muter's own rendering changes, so this is never checked both ways
-- (compare user_blocks.IsBlockedEitherDirection).
SELECT EXISTS(
    SELECT 1
    FROM user_mutes
    WHERE muter_user_id = sqlc.arg(muter_user_id)
        AND muted_user_id = sqlc.arg(muted_user_id)
) AS muted;
