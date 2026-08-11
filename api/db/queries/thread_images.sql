-- name: AddThreadImage :one
INSERT INTO thread_images(thread_id, image_path, image_alt)
VALUES (sqlc.arg(thread_id), sqlc.arg(image_path), sqlc.narg(image_alt))
RETURNING *;

-- name: ListThreadImagesByThreadID :many
SELECT *
FROM thread_images
WHERE thread_id = sqlc.arg(thread_id)
ORDER BY sort_order, created_at;

-- name: ListThreadImagePathsByOwnerID :many
-- Storage cleanup targets for account deletion — every image path
-- belonging to a Thread this user owns, gathered before
-- SoftDeleteThreadsByUserID runs (see UserUsecase.DeleteAccount).
SELECT thread_images.image_path
FROM thread_images
    JOIN threads ON threads.id = thread_images.thread_id
WHERE threads.user_id = sqlc.arg(owner_user_id)
    AND threads.deleted_at IS NULL;

-- name: DeleteThreadImage :one
-- Returns the deleted row (specifically image_path) so the caller can
-- remove the matching storage object after the DB record is gone — see
-- ThreadImageUsecase.DeleteThreadImage for the ordering rationale. No
-- matching row (already deleted / wrong thread) surfaces as
-- pgx.ErrNoRows, which the repository treats as a silent no-op, same as
-- the :exec version this replaces.
DELETE FROM thread_images
WHERE id = sqlc.arg(id)
    AND thread_id = sqlc.arg(thread_id)
RETURNING *;
