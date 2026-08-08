-- name: UpsertReaction :one
-- Reacting again with a different kind just changes it — there is no
-- separate update endpoint (FEATURES.md: a fixed emoji set, one
-- reaction per person per Moment per audience). Matches
-- uq_reactions_moment_user_circle (NULLS NOT DISTINCT WHERE user_id IS
-- NOT NULL).
INSERT INTO reactions(moment_id, user_id, circle_id, kind)
VALUES (sqlc.arg(moment_id), sqlc.arg(user_id), sqlc.narg(circle_id), sqlc.arg(kind))
ON CONFLICT (moment_id, user_id, circle_id) WHERE user_id IS NOT NULL
DO UPDATE SET kind = excluded.kind
RETURNING *;

-- name: DeleteReaction :exec
DELETE FROM reactions
WHERE moment_id = sqlc.arg(moment_id)
    AND user_id = sqlc.arg(user_id)
    AND circle_id IS NOT DISTINCT FROM sqlc.narg(circle_id);

-- name: ListReactionsForMoment :many
-- Who reacted, never how many — FEATURES.md: "Reaction counts are never
-- displayed as a number." Same visibility/mute shape as
-- ListCommentsForMoment.
SELECT reactions.*
FROM reactions
WHERE reactions.moment_id = sqlc.arg(moment_id)
    AND (
        (sqlc.arg(mention_allowed)::bool AND reactions.circle_id IS NULL)
        OR reactions.circle_id = ANY(sqlc.arg(visible_circle_ids)::uuid[])
    )
    AND (
        reactions.user_id IS NULL
        OR NOT EXISTS (
            SELECT 1 FROM user_mutes
            WHERE muter_user_id = sqlc.arg(viewer_id)
                AND muted_user_id = reactions.user_id
        )
    )
ORDER BY reactions.created_at;
