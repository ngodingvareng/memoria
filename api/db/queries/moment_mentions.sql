-- name: CreateMomentMention :one
-- display_name is a snapshot taken at mention time (see the column
-- comment): the caller resolves it from the mentioned user's current
-- name before calling this, it is not derived here.
INSERT INTO moment_mentions(moment_id, mentioned_user_id, display_name)
VALUES (sqlc.arg(moment_id), sqlc.arg(mentioned_user_id), sqlc.arg(display_name))
RETURNING *;

-- name: ListMomentMentionsByMomentID :many
SELECT *
FROM moment_mentions
WHERE moment_id = sqlc.arg(moment_id)
    AND removed_at IS NULL
ORDER BY created_at;

-- name: ListMentionedMomentsByUserID :many
-- "Mentioned Moments" (FEATURES.md, Looking Back): every Moment the
-- user has ever been mentioned in, newest occurrence first.
SELECT moments.*
FROM moment_mentions
    JOIN moments ON moments.id = moment_mentions.moment_id
WHERE moment_mentions.mentioned_user_id = sqlc.arg(mentioned_user_id)
    AND moment_mentions.removed_at IS NULL
    AND moments.deleted_at IS NULL
ORDER BY moments.occurred_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: RemoveMomentMention :exec
-- "Leaving a mention": the mentioned user removes themselves. The
-- Moment stays with its owner, unchanged apart from this — see
-- FEATURES.md, Leaving a mention.
UPDATE moment_mentions
SET removed_at = NOW()
WHERE moment_id = sqlc.arg(moment_id)
    AND mentioned_user_id = sqlc.arg(mentioned_user_id)
    AND removed_at IS NULL;

-- name: DeleteMomentMention :exec
-- The Moment's owner deleting a mention they added by mistake — distinct
-- from RemoveMomentMention, which is the mentioned user leaving on
-- their own initiative.
DELETE FROM moment_mentions
WHERE id = sqlc.arg(id)
    AND moment_id = sqlc.arg(moment_id);
