-- name: AddMomentImage :one
INSERT INTO moment_images(moment_id, image_path, image_alt)
VALUES (sqlc.arg(moment_id), sqlc.arg(image_path), sqlc.narg(image_alt))
RETURNING *;

-- name: ListMomentImagesByMomentID :many
SELECT *
FROM moment_images
WHERE moment_id = sqlc.arg(moment_id)
ORDER BY created_at;

-- name: DeleteMomentImage :exec
DELETE FROM moment_images
WHERE id = sqlc.arg(id)
    AND moment_id = sqlc.arg(moment_id);
