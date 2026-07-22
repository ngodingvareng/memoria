-- name: CreateUser :one
INSERT INTO users(name, email, timezone)
VALUES (sqlc.arg(name), sqlc.arg(email), sqlc.arg(timezone))
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = sqlc.arg(id)
    AND deleted_at IS NULL;

-- name: GetUserByEmail :one
-- Case-insensitive lookup, matches uq_users_email_lower.
SELECT *
FROM users
WHERE lower(email) = lower(sqlc.arg(email))
    AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
UPDATE users
SET name = sqlc.arg(name),
    image_path = sqlc.narg(image_path),
    timezone = sqlc.arg(timezone),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND deleted_at IS NULL
    RETURNING *;

-- name: SetUserEmailVerified :exec
UPDATE users
SET email_verified = TRUE,
    updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: SoftDeleteUser :exec
-- Starts the recovery grace period. Caller should also revoke all
-- sessions (see user_sessions.DeleteAllSessionsByUserID) in the same
-- transaction.
UPDATE users
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
    AND deleted_at IS NULL;

-- name: RestoreUser :exec
UPDATE users
SET deleted_at = NULL
WHERE id = sqlc.arg(id);

-- name: ListUsersPendingPurge :many
-- For a purge job: pass purge_before = NOW() - INTERVAL '30 days' (or
-- whatever grace period is chosen) from the application.
SELECT *
FROM users
WHERE deleted_at IS NOT NULL
    AND deleted_at < sqlc.arg(purge_before);

-- name: HardDeleteUser :exec
-- Actually cascades into user_accounts/user_sessions/activities/... Only
-- call this from the purge job, on rows already past their grace period.
DELETE FROM users
WHERE id = sqlc.arg(id);