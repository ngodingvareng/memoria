-- name: CreateResponseEvent :exec
-- Populates the batched daily digest queue (FEATURES.md, Notification:
-- "Responses ... batched into a single daily digest, never delivered
-- individually"). Nothing reads or digests this yet — that's the
-- Notification domain's worker, not built yet (same shape as the
-- Commitment scheduler/timeout workers in TODO.md). The caller never
-- writes a row when actor_user_id = recipient_user_id (reacting/
-- commenting on your own Moment is not a response event).
INSERT INTO response_events(recipient_user_id, actor_user_id, moment_id, kind, comment_id, reaction_id)
VALUES (
    sqlc.arg(recipient_user_id), sqlc.arg(actor_user_id), sqlc.arg(moment_id),
    sqlc.arg(kind), sqlc.narg(comment_id), sqlc.narg(reaction_id)
);
