-- name: AddActivityImage :one
INSERT INTO activity_images(activity_id, image_path, image_alt)
VALUES (sqlc.arg(activity_id), sqlc.arg(image_path), sqlc.narg(image_alt))
RETURNING *;

-- name: ListActivityImagesByActivityID :many
SELECT *
FROM activity_images
WHERE activity_id = sqlc.arg(activity_id)
ORDER BY created_at;

-- name: DeleteActivityImage :exec
DELETE FROM activity_images
WHERE id = sqlc.arg(id)
    AND activity_id = sqlc.arg(activity_id);