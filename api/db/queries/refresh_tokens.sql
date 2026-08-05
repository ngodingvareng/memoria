-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, family_id, token_hash, expires_at, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRefreshTokenByTokenHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- Rotate is a compare-and-swap: it only succeeds if the row is still
-- active (revoked_at IS NULL) when this statement runs. If two
-- concurrent Refresh requests race on the same token, only one can
-- win; the loser gets 0 affected rows and the usecase rolls back its
-- whole transaction.
-- name: RotateRefreshToken :execrows
UPDATE refresh_tokens SET revoked_at = NOW(), replaced_by_id = $2, updated_at = NOW()
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshTokenFamily :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW()
WHERE family_id = $1 AND revoked_at IS NULL;

-- name: RevokeAllRefreshTokensByUserID :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;