-- name: CreateOAuthUserAccount :one
INSERT INTO user_accounts(
    user_id, account_id, provider_id,
    access_token, refresh_token,
    access_token_expires_at, refresh_token_expires_at,
    scope, id_token
) VALUES (
    sqlc.arg(user_id), sqlc.arg(account_id), sqlc.arg(provider_id),
    sqlc.narg(access_token), sqlc.narg(refresh_token),
    sqlc.narg(access_token_expires_at), sqlc.narg(refresh_token_expires_at),
    sqlc.narg(scope), sqlc.narg(id_token)
)
RETURNING *;

-- name: CreateCredentialUserAccount :one
INSERT INTO user_accounts(user_id, account_id, provider_id, password_hash)
VALUES (sqlc.arg(user_id), sqlc.arg(account_id), 'credential', sqlc.arg(password_hash))
RETURNING *;

-- name: GetUserAccountByProvider :one
-- Login lookup: OAuth callback (provider_id, account_id from the
-- provider) or credential login (provider_id = 'credential', account_id
-- = whatever convention the app uses, e.g. the user's id or email).
SELECT *
FROM user_accounts
WHERE provider_id = sqlc.arg(provider_id)
    AND account_id = sqlc.arg(account_id);

-- name: GetCredentialUserAccountByUserID :one
SELECT *
FROM user_accounts
WHERE user_id = sqlc.arg(user_id)
    AND provider_id = 'credential';

-- name: ListUserAccountsByUserID :many
-- "Linked accounts" list in account settings.
SELECT *
FROM user_accounts
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at;

-- name: UpdateOAuthUserAccountTokens :exec
-- Called after a token refresh with the provider.
UPDATE user_accounts
SET access_token = sqlc.arg(access_token),
    refresh_token = sqlc.narg(refresh_token),
    access_token_expires_at = sqlc.narg(access_token_expires_at),
    refresh_token_expires_at = sqlc.narg(refresh_token_expires_at),
    updated_at = NOW()
WHERE id = sqlc.arg(id);

-- name: UpdateUserAccountPasswordHash :exec
-- Password change / reset flow.
UPDATE user_accounts
SET password_hash = sqlc.arg(password_hash),
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
    AND provider_id = 'credential';

-- name: DeleteUserAccount :exec
-- Unlink a provider. user_id in the WHERE clause is a safety check so a
-- caller can't unlink someone else's account by guessing an id.
DELETE FROM user_accounts
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);

-- name: IncrementFailedLoginAttempts :one
-- Called after a wrong password. The usecase decides whether the
-- returned failed_login_attempts count crosses the configured
-- lockout threshold and, if so, calls LockCredentialUserAccount.
UPDATE user_accounts
SET failed_login_attempts = failed_login_attempts + 1,
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
    AND provider_id = 'credential'
RETURNING *;

-- name: LockCredentialUserAccount :exec
UPDATE user_accounts
SET locked_until = sqlc.arg(locked_until),
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
    AND provider_id = 'credential';

-- name: ResetFailedLoginAttempts :exec
-- Called after a successful login to clear any accumulated failures.
UPDATE user_accounts
SET failed_login_attempts = 0,
    locked_until = NULL,
    updated_at = NOW()
WHERE user_id = sqlc.arg(user_id)
    AND provider_id = 'credential';