-- name: AddMomentImage :one
INSERT INTO moment_images(
    moment_id, image_path, image_alt,
    content_type, byte_size, width, height, metadata_stripped
) VALUES (
    sqlc.arg(moment_id), sqlc.arg(image_path), sqlc.narg(image_alt),
    sqlc.narg(content_type), sqlc.narg(byte_size), sqlc.narg(width), sqlc.narg(height),
    sqlc.arg(metadata_stripped)
)
RETURNING *;

-- name: ListMomentImagesByMomentID :many
SELECT *
FROM moment_images
WHERE moment_id = sqlc.arg(moment_id)
ORDER BY sort_order, created_at;

-- name: DeleteMomentImage :one
-- Returns the deleted row (specifically image_path) so the caller can
-- remove the matching storage object after the DB record is gone — see
-- MomentImageUsecase.DeleteMomentImage for the ordering rationale. No
-- matching row (already deleted / wrong moment) surfaces as
-- pgx.ErrNoRows, which the repository treats as a silent no-op, same as
-- the :exec version this replaces.
DELETE FROM moment_images
WHERE id = sqlc.arg(id)
    AND moment_id = sqlc.arg(moment_id)
RETURNING *;
