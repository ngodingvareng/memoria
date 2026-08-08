-- name: CreateCircle :one
INSERT INTO circles(name, description, color_hex, image_path, created_by_user_id)
VALUES (
    sqlc.arg(name), sqlc.narg(description), sqlc.narg(color_hex), sqlc.narg(image_path),
    sqlc.arg(created_by_user_id)
)
RETURNING *;

-- name: GetCircleWithAccess :one
-- "Can this viewer reach this Circle at all" — true for any active
-- member, mirrors threads.GetThreadWithAccess.
SELECT circles.*
FROM circles
WHERE circles.id = sqlc.arg(id)
    AND circles.dissolved_at IS NULL
    AND EXISTS (
        SELECT 1 FROM circle_members
        WHERE circle_id = circles.id
            AND user_id = sqlc.arg(user_id)
            AND left_at IS NULL
    );

-- name: ListCirclesByUserID :many
-- The "My Circles" list: every Circle this user is currently an active
-- member of.
SELECT circles.*
FROM circles
    JOIN circle_members ON circle_members.circle_id = circles.id
WHERE circle_members.user_id = sqlc.arg(user_id)
    AND circle_members.left_at IS NULL
    AND circles.dissolved_at IS NULL
ORDER BY circles.created_at DESC;

-- name: UpdateCircle :one
-- Admin-only, enforced in the WHERE clause rather than the usecase layer
-- so a non-admin's attempt and a wrong/foreign id are indistinguishable
-- from here — same "no row = 404" reasoning as ThreadRepository.Update.
UPDATE circles
SET name = sqlc.arg(name),
    description = sqlc.narg(description),
    color_hex = sqlc.narg(color_hex),
    image_path = sqlc.narg(image_path),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND dissolved_at IS NULL
    AND EXISTS (
        SELECT 1 FROM circle_members
        WHERE circle_id = circles.id
            AND user_id = sqlc.arg(user_id)
            AND left_at IS NULL
            AND role = 'admin'
    )
RETURNING *;

-- name: DissolveCircle :exec
-- Admin-only (see UpdateCircle). :exec — a mismatched WHERE (wrong id,
-- not an admin, already dissolved) is a silent no-op, not
-- errs.ErrNotFound; callers needing a real 404 look it up via
-- GetCircleWithAccess first.
UPDATE circles
SET dissolved_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND dissolved_at IS NULL
    AND EXISTS (
        SELECT 1 FROM circle_members
        WHERE circle_id = circles.id
            AND user_id = sqlc.arg(user_id)
            AND left_at IS NULL
            AND role = 'admin'
    );
