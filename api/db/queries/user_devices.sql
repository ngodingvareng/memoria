-- name: UpsertUserDevice :one
-- Registering a push token. Re-registering an existing active token
-- (reinstall, re-login on the same device) refreshes ownership and
-- last_seen_at instead of failing uq_user_devices_push_token.
INSERT INTO user_devices(user_id, platform, push_token, timezone)
VALUES (sqlc.arg(user_id), sqlc.arg(platform), sqlc.arg(push_token), sqlc.narg(timezone))
ON CONFLICT (push_token) WHERE revoked_at IS NULL
DO UPDATE SET
    user_id = EXCLUDED.user_id,
    platform = EXCLUDED.platform,
    timezone = EXCLUDED.timezone,
    last_seen_at = NOW()
RETURNING *;

-- name: ListActiveUserDevicesByUserID :many
-- Every live push target for a recipient — notifications queued in
-- `notifications` are delivered to each of these.
SELECT *
FROM user_devices
WHERE user_id = sqlc.arg(user_id)
    AND revoked_at IS NULL;

-- name: TouchUserDeviceLastSeen :exec
UPDATE user_devices
SET last_seen_at = NOW()
WHERE id = sqlc.arg(id);

-- name: RevokeUserDevice :exec
-- Unregister a single device (e.g. explicit logout on that device).
UPDATE user_devices
SET revoked_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND revoked_at IS NULL;

-- name: RevokeAllUserDevicesByUserID :exec
-- Logout-everywhere / account deletion.
UPDATE user_devices
SET revoked_at = NOW()
WHERE user_id = sqlc.arg(user_id)
    AND revoked_at IS NULL;
