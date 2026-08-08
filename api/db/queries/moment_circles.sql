-- name: ShareMomentToCircle :one
-- The "Share to circle too?" step of the mention flow (FEATURES.md,
-- Mention) — never a standalone "share" action. Idempotent: sharing
-- into a Circle it's already shared to is a no-op.
INSERT INTO moment_circles(moment_id, circle_id, shared_by_user_id)
VALUES (sqlc.arg(moment_id), sqlc.arg(circle_id), sqlc.arg(shared_by_user_id))
ON CONFLICT (moment_id, circle_id) DO NOTHING
RETURNING *;

-- name: ListCircleIDsByMomentID :many
-- Which Circles a personal Moment is currently shared into, for its
-- own edit/unshare UI.
SELECT circle_id
FROM moment_circles
WHERE moment_id = sqlc.arg(moment_id);

-- name: ListMomentsSharedToCircle :many
-- Personal Moments shared into this Circle via the mention flow.
-- Combined with the Circle's own collaborative-Thread Moments
-- (threads.circle_id), this is the Circle's Album (FEATURES.md,
-- Looking Back).
SELECT moments.*, moment_circles.shared_by_user_id, moment_circles.shared_at
FROM moment_circles
    JOIN moments ON moments.id = moment_circles.moment_id
WHERE moment_circles.circle_id = sqlc.arg(circle_id)
    AND moments.deleted_at IS NULL
ORDER BY moment_circles.shared_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: UnshareMomentFromCircle :exec
DELETE FROM moment_circles
WHERE moment_id = sqlc.arg(moment_id)
    AND circle_id = sqlc.arg(circle_id);
