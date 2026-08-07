-- name: CreateCommitmentHistory :one
-- Archive a Commitment's values before it's changed or retired.
INSERT INTO commitment_histories(
    thread_id,
    commitment_id,
    cron_expression,
    timezone,
    active_from,
    active_until
)
VALUES (
    sqlc.arg(thread_id),
    sqlc.narg(commitment_id),
    sqlc.arg(cron_expression),
    sqlc.narg(timezone),
    sqlc.arg(active_from),
    sqlc.arg(active_until)
)
RETURNING *;

-- name: ListCommitmentHistoryByThreadID :many
-- "How has this Thread's commitments changed over time" view.
SELECT * FROM commitment_histories
WHERE thread_id = sqlc.arg(thread_id)
ORDER BY active_from DESC;
