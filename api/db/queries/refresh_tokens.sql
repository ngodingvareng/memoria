-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, family_id, token_hash, expires_at, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetRefreshTokenByTokenHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- Rotate itu compare-and-swap: cuma berhasil kalau baris masih aktif
-- (revoked_at IS NULL) di saat statement ini jalan. Kalau dua request
-- Refresh bersamaan lomba pakai token yang sama, hanya satu yang bisa
-- menang; yang kalah affected rows-nya 0 dan seluruh transaksinya
-- di-rollback usecase.
-- name: RotateRefreshToken :execrows
UPDATE refresh_tokens SET revoked_at = NOW(), replaced_by_id = $2, updated_at = NOW()
WHERE id = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshTokenFamily :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW()
WHERE family_id = $1 AND revoked_at IS NULL;

-- name: RevokeAllRefreshTokensByUserID :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL;