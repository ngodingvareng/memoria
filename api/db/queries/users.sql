-- name: CreateUser :one
-- username is nullable here: registration creates the account before the
-- onboarding step (SetUsername) claims one.
INSERT INTO users(name, username, email, timezone)
VALUES (sqlc.arg(name), sqlc.narg(username), sqlc.arg(email), sqlc.arg(timezone))
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

-- name: GetUserByUsername :one
-- Case-insensitive lookup, matches uq_users_username_lower. Backs
-- @mention resolution and "invite by entering their username" (Circle
-- Invite) — callers must additionally check discoverable_by_username
-- for the invite path, since a mentionable user may still opt out of
-- being found by handle.
SELECT *
FROM users
WHERE lower(username) = lower(sqlc.arg(username))
    AND deleted_at IS NULL;

-- name: UpdateUserImagePath :one
-- Sets the profile photo (or clears it, if narg is NULL). Separate from
-- UpdateUserProfile so an upload doesn't require resending name/
-- username/timezone alongside it.
UPDATE users
SET image_path = sqlc.narg(image_path),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND deleted_at IS NULL
RETURNING *;

-- name: SearchUsersByUsername :many
-- Prefix search over discoverable usernames, backing the mention-search
-- typeahead (FEATURES.md, Mention). Same discoverable_by_username gate
-- as the Circle Invite path (see GetUserByUsername) — mention_policy
-- and blocking are filtered in the usecase, same as CreateMention. The
-- caller is responsible for escaping any %/_ in query before binding.
SELECT *
FROM users
WHERE discoverable_by_username = TRUE
    AND deleted_at IS NULL
    AND username IS NOT NULL
    AND id != sqlc.arg(exclude_user_id)
    AND username ILIKE sqlc.arg(query) || '%'
ORDER BY username ASC
LIMIT sqlc.arg(page_limit);

-- name: UpdateUserProfile :one
UPDATE users
SET name = sqlc.arg(name),
    username = sqlc.arg(username),
    image_path = sqlc.narg(image_path),
    bio = sqlc.narg(bio),
    timezone = sqlc.arg(timezone),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND deleted_at IS NULL
    RETURNING *;

-- name: UpdateUserPrivacySettings :one
-- Social Interaction Controls + Data Controls (FEATURES.md, Privacy &
-- Control) — the account-level privacy toggles, separate from profile
-- content so a settings screen can save them independently.
UPDATE users
SET mention_policy = sqlc.arg(mention_policy),
    circle_invite_policy = sqlc.arg(circle_invite_policy),
    discoverable_by_username = sqlc.arg(discoverable_by_username),
    strip_photo_metadata = sqlc.arg(strip_photo_metadata),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND deleted_at IS NULL
    RETURNING *;

-- name: SetUsername :one
-- Claims a username for an account that doesn't have one yet (the
-- post-register onboarding step). uq_users_username_lower/
-- chk_users_username_format still guard uniqueness/format at the DB
-- level regardless of caller-side validation.
UPDATE users
SET username = sqlc.arg(username),
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
-- refresh tokens (see refresh_tokens.RevokeAllRefreshTokensByUserID) in
-- the same transaction.
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
-- Actually cascades into user_accounts/refresh_tokens/threads/moments/...
-- Only call this from the purge job, on rows already past their grace
-- period.
DELETE FROM users
WHERE id = sqlc.arg(id);