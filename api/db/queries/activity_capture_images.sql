-- name: AddActivityCaptureImage :one
INSERT INTO activity_capture_images(activity_capture_id, image_path, image_alt)
VALUES (sqlc.arg(activity_capture_id), sqlc.arg(image_path), sqlc.narg(image_alt))
RETURNING *;

-- name: ListActivityCaptureImagesByItemID :many
SELECT *
FROM activity_capture_images
WHERE activity_capture_id = sqlc.arg(activity_capture_id)
ORDER BY created_at;

-- name: DeleteActivityCaptureImage :exec
DELETE FROM activity_capture_images
WHERE id = sqlc.arg(id)
    AND activity_capture_id = sqlc.arg(activity_capture_id);