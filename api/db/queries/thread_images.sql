-- name: AddThreadImage :one
INSERT INTO thread_images(thread_id, image_path, image_alt)
VALUES (sqlc.arg(thread_id), sqlc.arg(image_path), sqlc.narg(image_alt))
RETURNING *;

-- name: ListThreadImagesByThreadID :many
SELECT *
FROM thread_images
WHERE thread_id = sqlc.arg(thread_id)
ORDER BY sort_order, created_at;

-- name: DeleteThreadImage :exec
DELETE FROM thread_images
WHERE id = sqlc.arg(id)
    AND thread_id = sqlc.arg(thread_id);
