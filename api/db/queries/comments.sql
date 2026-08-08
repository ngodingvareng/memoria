-- name: CreateComment :one
-- circle_id NULL is the mention context; set, it names which Circle
-- audience the comment was posted in (FEATURES.md, Response). The
-- caller (CommentUsecase, via ResponseAccessChecker) has already
-- verified the requester may act in that specific context before this
-- runs.
INSERT INTO comments(moment_id, user_id, circle_id, body)
VALUES (sqlc.arg(moment_id), sqlc.arg(user_id), sqlc.narg(circle_id), sqlc.arg(body))
RETURNING *;

-- name: GetCommentByID :one
SELECT *
FROM comments
WHERE id = sqlc.arg(id)
    AND deleted_at IS NULL;

-- name: ListCommentsForMoment :many
-- Visibility is resolved by the caller (ResponseAccessChecker) and
-- passed in as mention_allowed/visible_circle_ids — this query only
-- applies it, plus the Mute view filter (never an access gate, per
-- FEATURES.md Response Terms: "Muted users aren't restricted at all").
SELECT comments.*
FROM comments
WHERE comments.moment_id = sqlc.arg(moment_id)
    AND comments.deleted_at IS NULL
    AND (
        (sqlc.arg(mention_allowed)::bool AND comments.circle_id IS NULL)
        OR comments.circle_id = ANY(sqlc.arg(visible_circle_ids)::uuid[])
    )
    AND (
        comments.user_id IS NULL
        OR NOT EXISTS (
            SELECT 1 FROM user_mutes
            WHERE muter_user_id = sqlc.arg(viewer_id)
                AND muted_user_id = comments.user_id
        )
    )
ORDER BY comments.created_at;

-- name: UpdateComment :one
UPDATE comments
SET body = sqlc.arg(body),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteComment :exec
UPDATE comments
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;
