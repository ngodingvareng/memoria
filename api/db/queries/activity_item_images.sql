-- name: AddActivityItemImage :one
INSERT INTO activity_item_images(activity_item_id, image_path, image_alt)
VALUES (sqlc.arg(activity_item_id), sqlc.arg(image_path), sqlc.narg(image_alt))
RETURNING *;

-- name: ListActivityItemImagesByItemID :many
SELECT *
FROM activity_item_images
WHERE activity_item_id = sqlc.arg(activity_item_id)
ORDER BY created_at;

-- name: DeleteActivityItemImage :exec
DELETE FROM activity_item_images
WHERE id = sqlc.arg(id)
    AND activity_item_id = sqlc.arg(activity_item_id);