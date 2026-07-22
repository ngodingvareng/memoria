-- name: CreateUserVerification :one
INSERT INTO user_verifications(identifier, value, expires_at)
VALUES (sqlc.arg(identifier), sqlc.arg(value), sqlc.arg(expires_at))
RETURNING *;

-- name: GetValidUserVerification :one
-- Validate an incoming OTP/token. The app should still delete it (see
-- DeleteUserVerification) in the same transaction to prevent reuse.
SELECT *
FROM user_verifications
WHERE identifier = sqlc.arg(identifier)
    AND value = sqlc.arg(value)
    AND expires_at > NOW();

-- name: DeleteUserVerification :exec
DELETE FROM user_verifications
WHERE id = sqlc.arg(id);

-- name: DeleteUserVerificationsByIdentifier :exec
-- Invalidate any previously-issued pending codes before issuing a new
-- one for the same identifier (e.g. "resend OTP").
DELETE FROM user_verifications
WHERE identifier = sqlc.arg(identifier);

-- name: DeleteExpiredUserVerifications :execrows
-- Periodic cleanup job.
DELETE FROM user_verifications
WHERE expires_at <= NOW();