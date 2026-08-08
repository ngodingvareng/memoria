-- Only the queries the Circle Invite flow needs; general membership
-- management (role changes, permission edits, leaving, listing a
-- Circle's roster) is not written yet.

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

-- name: DoUsersShareAnyCircle :one
-- Shared-Circle membership on its own, for surfaces that genuinely mean
-- that and nothing else — such as the mention flow's optional "Share to
-- circle too?" step. Both memberships must be active: a Circle one of
-- them has left is not shared ground.
--
-- This is NOT how audience_policy = 'known' is evaluated. That tier is a
-- union of this and an explicit mark, and has its own query — use
-- IsUserKnownTo, or a user who marked someone without sharing a Circle
-- will be told they are unreachable.
SELECT EXISTS(
    SELECT 1
    FROM circle_members AS a
        JOIN circle_members AS b ON b.circle_id = a.circle_id
    WHERE a.user_id = sqlc.arg(user_a)
        AND b.user_id = sqlc.arg(user_b)
        AND a.left_at IS NULL
        AND b.left_at IS NULL
) AS shares_circle;
