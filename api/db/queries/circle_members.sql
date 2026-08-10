-- Beyond the Circle Invite flow queries, this file also carries general
-- membership management: role changes, permission edits, leaving, and
-- listing a Circle's roster.

-- name: AddCircleCreatorAsAdmin :one
-- The one write CreateCircle's caller must run in the same transaction
-- right after inserting the Circle row (FEATURES.md, Circle
-- Permissions: "The user who creates a circle becomes its admin").
-- Deliberately separate from AddCircleMembers, which is the invite
-- flow's admit path and always seats people as plain members.
INSERT INTO circle_members(circle_id, user_id, role, can_invite, can_capture)
VALUES (sqlc.arg(circle_id), sqlc.arg(user_id), 'admin', TRUE, TRUE)
RETURNING *;

-- name: ListActiveCircleMembers :many
-- The Circle's roster.
SELECT *
FROM circle_members
WHERE circle_id = sqlc.arg(circle_id)
    AND left_at IS NULL
ORDER BY joined_at;

-- name: IsSoleActiveCircleAdmin :one
-- True when user_id is currently an active admin of circle_id and no
-- other active admin exists — the guard LeaveCircle and
-- UpdateMemberRole (usecase layer) check before letting an admin
-- self-leave or self-demote, so a Circle is never left with zero
-- admins. Same TOCTOU tolerance as the other unlocked read-then-write
-- sequences already accepted elsewhere in this codebase (see TODO.md).
SELECT
    EXISTS(
        SELECT 1 FROM circle_members AS target
        WHERE target.circle_id = sqlc.arg(circle_id)
            AND target.user_id = sqlc.arg(user_id)
            AND target.left_at IS NULL
            AND target.role = 'admin'
    )
    AND NOT EXISTS(
        SELECT 1 FROM circle_members AS other
        WHERE other.circle_id = sqlc.arg(circle_id)
            AND other.user_id != sqlc.arg(user_id)
            AND other.left_at IS NULL
            AND other.role = 'admin'
    ) AS is_sole_admin;

-- name: LeaveCircle :exec
-- Self-service departure. :exec — a mismatched WHERE (not a member,
-- already left) is a silent no-op; the caller already knows the actor's
-- own membership state without re-checking here.
UPDATE circle_members
SET left_at = NOW()
WHERE circle_id = sqlc.arg(circle_id)
    AND user_id = sqlc.arg(user_id)
    AND left_at IS NULL;

-- name: RemoveCircleMember :exec
-- Admin-forced removal of someone else. Requires the actor to be an
-- active admin, checked in the same statement so a non-admin's attempt
-- and a foreign circle_id/user_id are indistinguishable — same
-- no-row-is-a-no-op reasoning as LeaveCircle. The target row is aliased
-- explicitly since the EXISTS subquery below also scans circle_members —
-- an unaliased target reads as ambiguous to sqlc's analyzer otherwise.
UPDATE circle_members AS target
SET left_at = NOW()
WHERE target.circle_id = sqlc.arg(circle_id)
    AND target.user_id = sqlc.arg(user_id)
    AND target.left_at IS NULL
    AND EXISTS (
        SELECT 1 FROM circle_members AS remover
        WHERE remover.circle_id = sqlc.arg(circle_id)
            AND remover.user_id = sqlc.arg(removed_by_user_id)
            AND remover.left_at IS NULL
            AND remover.role = 'admin'
    );

-- name: UpdateCircleMemberRole :one
-- Admin-only, same actor-check shape as RemoveCircleMember. Promoting to
-- 'admin' also forces can_invite = TRUE in the same statement —
-- chk_circle_members_admin_can_invite requires it, and leaving that to a
-- second call would let the row violate the constraint in between.
UPDATE circle_members AS target
SET role = sqlc.arg(role),
    can_invite = CASE WHEN sqlc.arg(role)::circle_role = 'admin' THEN TRUE ELSE target.can_invite END
WHERE target.circle_id = sqlc.arg(circle_id)
    AND target.user_id = sqlc.arg(user_id)
    AND target.left_at IS NULL
    AND EXISTS (
        SELECT 1 FROM circle_members AS actor
        WHERE actor.circle_id = sqlc.arg(circle_id)
            AND actor.user_id = sqlc.arg(changed_by_user_id)
            AND actor.left_at IS NULL
            AND actor.role = 'admin'
    )
RETURNING *;

-- name: UpdateCircleMemberPermissions :one
-- Admin-only. Setting can_invite = FALSE on an active admin row would
-- violate chk_circle_members_admin_can_invite and fails at the database
-- level — callers should only offer this control for 'member' rows.
UPDATE circle_members AS target
SET can_invite = sqlc.arg(can_invite),
    can_capture = sqlc.arg(can_capture)
WHERE target.circle_id = sqlc.arg(circle_id)
    AND target.user_id = sqlc.arg(user_id)
    AND target.left_at IS NULL
    AND EXISTS (
        SELECT 1 FROM circle_members AS actor
        WHERE actor.circle_id = sqlc.arg(circle_id)
            AND actor.user_id = sqlc.arg(changed_by_user_id)
            AND actor.left_at IS NULL
            AND actor.role = 'admin'
    )
RETURNING *;

-- name: AddCircleMembers :many
-- The single write that admits people, whichever path they came in by:
-- a direct add, an accepted username invite, or an approved join
-- request. Takes a set because a direct add names several users at once
-- (FEATURES.md, Circle Invite).
--
-- The upsert is not decoration. Leaving a Circle stamps left_at instead
-- of deleting the row, so the primary key survives the departure and a
-- returning member would collide with it on a plain INSERT. Rejoining
-- resets role and permissions to the defaults rather than restoring what
-- the user held before: otherwise an admin who was removed could be
-- handed admin back by any member who happens to have invite rights.
--
-- Users who are already active members conflict but do not match the
-- WHERE, so they are silently skipped and absent from the result — which
-- makes the call idempotent and the returned rows exactly "who newly
-- joined".
INSERT INTO circle_members(circle_id, user_id)
SELECT sqlc.arg(circle_id)::uuid, member_id
FROM unnest(sqlc.arg(user_ids)::uuid[]) AS member_id
ON CONFLICT (circle_id, user_id) DO UPDATE
SET left_at = NULL,
    joined_at = NOW(),
    role = 'member',
    can_invite = FALSE,
    can_capture = TRUE
WHERE circle_members.left_at IS NOT NULL
RETURNING *;

-- name: GetActiveCircleMember :one
-- Active membership is (left_at IS NULL), never mere row existence — see
-- the migration's comment on circle_members.
SELECT *
FROM circle_members
WHERE circle_id = sqlc.arg(circle_id)
    AND user_id = sqlc.arg(user_id)
    AND left_at IS NULL;

-- name: CanUserInviteToCircle :one
-- Gates every entry point in the Circle Invite flow: adding users
-- directly, generating or rotating the link, and deciding join requests.
-- Checks can_invite alone because chk_circle_members_admin_can_invite
-- guarantees an admin always carries it — role never has to be tested
-- separately here.
SELECT EXISTS(
    SELECT 1
    FROM circle_members
    WHERE circle_id = sqlc.arg(circle_id)
        AND user_id = sqlc.arg(user_id)
        AND left_at IS NULL
        AND can_invite = TRUE
) AS can_invite;

-- name: ListCircleIDsSharedBetweenUsers :many
-- Which Circles user_a and user_b both actively belong to — for
-- surfaces that genuinely mean that and nothing else, such as the
-- mention flow's optional "Share to circle too?" step (FEATURES.md,
-- Mention). Both memberships must be active: a Circle one of them has
-- left is not shared ground. Deliberately returns the ids, not just a
-- boolean — the offer needs to name the Circles, and computing this
-- server-side means user_b's other, unrelated Circle memberships are
-- never exposed to user_a.
--
-- This is NOT how audience_policy = 'known' is evaluated. That tier is a
-- union of this and an explicit mark, and has its own query — use
-- IsUserKnownTo, or a user who marked someone without sharing a Circle
-- will be told they are unreachable.
SELECT a.circle_id
FROM circle_members AS a
    JOIN circle_members AS b ON b.circle_id = a.circle_id
WHERE a.user_id = sqlc.arg(user_a)
    AND b.user_id = sqlc.arg(user_b)
    AND a.left_at IS NULL
    AND b.left_at IS NULL;
