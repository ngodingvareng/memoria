-- name: CreateDataExport :one
-- "Export all my data" (FEATURES.md, Data Controls). Starts in
-- 'pending'; the export worker moves it through running -> ready/failed.
INSERT INTO data_exports(user_id)
VALUES (sqlc.arg(user_id))
RETURNING *;

-- name: GetDataExportByID :one
-- user_id in the WHERE clause doubles as an ownership check.
SELECT *
FROM data_exports
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);

-- name: ListDataExportsByUserID :many
-- Export history, newest first.
SELECT *
FROM data_exports
WHERE user_id = sqlc.arg(user_id)
ORDER BY requested_at DESC;

-- name: GetLatestActiveDataExportByUserID :one
-- Lets the usecase avoid queuing a duplicate export while one is
-- already pending/running for this user.
SELECT *
FROM data_exports
WHERE user_id = sqlc.arg(user_id)
    AND status IN ('pending', 'running')
ORDER BY requested_at DESC
LIMIT 1;

-- name: MarkDataExportRunning :exec
UPDATE data_exports
SET status = 'running'
WHERE id = sqlc.arg(id)
    AND status = 'pending';

-- name: MarkDataExportReady :exec
UPDATE data_exports
SET status = 'ready',
    file_path = sqlc.arg(file_path),
    byte_size = sqlc.arg(byte_size),
    completed_at = NOW(),
    expires_at = sqlc.arg(expires_at)
WHERE id = sqlc.arg(id);

-- name: MarkDataExportFailed :exec
UPDATE data_exports
SET status = 'failed',
    error_message = sqlc.arg(error_message),
    completed_at = NOW()
WHERE id = sqlc.arg(id);

-- name: ListExpiredReadyDataExports :many
-- Periodic cleanup job: exports whose download link has expired and
-- whose stored file can now be removed from object storage.
SELECT *
FROM data_exports
WHERE status = 'ready'
    AND expires_at IS NOT NULL
    AND expires_at < NOW();

-- name: DeleteDataExport :exec
-- Called after the underlying file has been removed from storage.
DELETE FROM data_exports
WHERE id = sqlc.arg(id);
