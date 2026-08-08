-- name: CreateUserKnown :exec
-- Idempotent: marking someone already marked is a no-op, not an error.
-- Nothing is notified and nothing is returned to the other user — see
-- the user_knows migration comment.
INSERT INTO user_knows(knower_user_id, known_user_id)
VALUES (sqlc.arg(knower_user_id), sqlc.arg(known_user_id))
ON CONFLICT (knower_user_id, known_user_id) DO NOTHING;

-- name: DeleteUserKnown :exec
-- Unmarking is equally silent. Access granted by the mark stops at the
-- next policy check; nothing already captured is affected (Principle 4).
DELETE FROM user_knows
WHERE knower_user_id = sqlc.arg(knower_user_id)
    AND known_user_id = sqlc.arg(known_user_id);

-- name: ListKnownUsersByKnowerID :many
-- The "People I know" settings surface. There is no query for the
-- reverse direction on purpose: who marked *me* is not information
-- Memoria exposes, and adding it would turn this into the social graph
-- FEATURES.md rules out.
SELECT *
FROM user_knows
WHERE knower_user_id = sqlc.arg(knower_user_id)
ORDER BY created_at DESC;

-- name: IsUserKnownTo :one
-- Resolves audience_policy = 'known' for users.mention_policy and
-- users.circle_invite_policy. The union is the definition, not a
-- convenience: "known" means the owner marked them, or the two already
-- share an active Circle (see the audience_policy comment in the
-- migration for why the second half cannot be dropped).
--
-- Argument order matters and is not symmetric. owner_user_id is the
-- person whose policy is being enforced; other_user_id is the one trying
-- to reach them. Only the owner's outgoing mark counts, so passing these
-- the wrong way round would let anyone grant themselves access.
--
-- Blocking is checked separately and overrides this either way — see
-- user_blocks.IsBlockedEitherDirection.
SELECT (
    EXISTS(
        SELECT 1
        FROM user_knows
        WHERE knower_user_id = sqlc.arg(owner_user_id)
            AND known_user_id = sqlc.arg(other_user_id)
    )
    OR EXISTS(
        SELECT 1
        FROM circle_members AS owner_membership
            JOIN circle_members AS other_membership
                ON other_membership.circle_id = owner_membership.circle_id
        WHERE owner_membership.user_id = sqlc.arg(owner_user_id)
            AND other_membership.user_id = sqlc.arg(other_user_id)
            AND owner_membership.left_at IS NULL
            AND other_membership.left_at IS NULL
    )
) AS is_known;
