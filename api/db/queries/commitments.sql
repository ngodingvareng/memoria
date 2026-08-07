-- name: CreateCommitment :one
INSERT INTO commitments(thread_id, cron_expression, timezone)
VALUES (sqlc.arg(thread_id), sqlc.arg(cron_expression), sqlc.narg(timezone))
RETURNING *;

-- name: GetCommitmentByID :one
SELECT *
FROM commitments
WHERE id = sqlc.arg(id) AND thread_id = sqlc.arg(thread_id);

-- name: ListCommitmentsByThreadID :many
-- Can return more than one row: a Thread may carry several concurrent
-- Commitments (e.g. a morning one and an evening one), each with its own
-- recurrence rule.
SELECT *
FROM commitments
WHERE thread_id = sqlc.arg(thread_id)
ORDER BY created_at;

-- name: UpdateCommitment :one
-- Update in place (same id) so existing moments.commitment_id links stay
-- valid. The caller must insert a row into commitment_histories with the
-- OLD cron_expression/timezone BEFORE calling this, in the same
-- transaction.
UPDATE commitments
SET cron_expression = sqlc.arg(cron_expression),
    timezone = sqlc.narg(timezone)
WHERE id = sqlc.arg(id) AND thread_id = sqlc.arg(thread_id)
    RETURNING *;

-- name: DeleteCommitment :exec
-- Retire a Commitment entirely. The caller should archive it into
-- commitment_histories first. Existing Moments that referenced it keep
-- thread_id intact and just get commitment_id set to NULL (composite FK,
-- ON DELETE SET NULL (commitment_id)).
DELETE FROM commitments
WHERE id = sqlc.arg(id) AND thread_id = sqlc.arg(thread_id);

-- name: ListCommitmentsForGeneration :many
-- Feed for the scheduler worker: every Commitment belonging to a
-- non-deleted Thread. The worker evaluates cron_expression/timezone
-- itself and calls moments.CreateCommittedMoment when a slot fires.
SELECT commitments.* FROM commitments
    JOIN threads ON threads.id = commitments.thread_id
WHERE threads.deleted_at IS NULL AND threads.has_commitment = TRUE;
