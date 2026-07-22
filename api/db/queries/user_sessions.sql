-- name: CreateUserSession :one
-- token_hash must already be hashed by the app (e.g. SHA-256 of the raw
-- token); the raw token itself is never passed to the DB.
INSERT INTO user_sessions(user_id, token_hash, expires_at, ip_address, user_agent)
VALUES (sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(expires_at), sqlc.narg(ip_address), sqlc.narg(user_agent))
RETURNING *;

-- name: GetUserSessionByTokenHash :one
SELECT *
FROM user_sessions
WHERE token_hash = sqlc.arg(token_hash)
    AND expires_at > NOW();

-- name: GetUserSessionWithUserByTokenHash :one
-- Convenience for auth middleware: session + owning (non-deleted) user
-- in a single round trip.
SELECT sqlc.embed(user_sessions), sqlc.embed(users)
FROM user_sessions
    JOIN users ON users.id = user_sessions.user_id
WHERE user_sessions.token_hash = sqlc.arg(token_hash)
    AND user_sessions.expires_at > NOW()
    AND users.deleted_at IS NULL;

-- name: ListUserSessionsByUserID :many
-- "Active devices/sessions" list in account settings.
SELECT *
FROM user_sessions
WHERE user_id = sqlc.arg(user_id)
    AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: DeleteUserSessionByTokenHash :exec
-- Logout from the current device.
DELETE FROM user_sessions
WHERE token_hash = sqlc.arg(token_hash);

-- name: DeleteUserSessionByID :exec
-- Revoke one specific session from the "active devices" list.
DELETE FROM user_sessions
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);

-- name: DeleteAllUserSessionsByUserID :exec
-- Logout everywhere: on password change, or when the account is
-- soft-deleted (see users.SoftDeleteUser).
DELETE FROM user_sessions
WHERE user_id = sqlc.arg(user_id);

-- name: DeleteExpiredUserSessions :execrows
-- Periodic cleanup job (Postgres has no TTL of its own).
DELETE FROM user_sessions
WHERE expires_at <= NOW();