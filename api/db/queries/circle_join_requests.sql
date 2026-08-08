-- name: CreateCircleJoinRequest :one
-- Written when someone follows a link whose requires_approval is TRUE.
-- A link without it admits people directly and never reaches here.
INSERT INTO circle_join_requests(circle_id, user_id, invite_id)
VALUES (sqlc.arg(circle_id), sqlc.arg(user_id), sqlc.arg(invite_id))
RETURNING *;

-- name: GetPendingCircleJoinRequestByUser :one
-- The "you already asked" check, so following the link twice reports the
-- outstanding request instead of tripping uq_circle_join_requests_pending.
SELECT *
FROM circle_join_requests
WHERE circle_id = sqlc.arg(circle_id)
    AND user_id = sqlc.arg(user_id)
    AND status = 'pending';

-- name: ListPendingCircleJoinRequestsByCircleID :many
-- The approval queue, for members with invite privileges. Oldest first:
-- someone waiting on an answer should not be pushed down the list by
-- newer arrivals.
SELECT *
FROM circle_join_requests
WHERE circle_id = sqlc.arg(circle_id)
    AND status = 'pending'
ORDER BY created_at
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountPendingCircleJoinRequestsByCircleID :one
SELECT COUNT(*) AS pending_count
FROM circle_join_requests
WHERE circle_id = sqlc.arg(circle_id)
    AND status = 'pending';

-- name: ApproveCircleJoinRequest :one
-- Returns no row if someone else already decided this request, which is
-- how two approvers acting at once resolve to one decision. Membership
-- is a separate AddCircleMembers call in the same transaction, and the
-- empty result must stop it.
UPDATE circle_join_requests
SET status = 'approved',
    decided_by_user_id = sqlc.arg(decided_by_user_id),
    decided_at = NOW()
WHERE id = sqlc.arg(id)
    AND circle_id = sqlc.arg(circle_id)
    AND status = 'pending'
RETURNING *;

-- name: RejectCircleJoinRequest :one
UPDATE circle_join_requests
SET status = 'rejected',
    decided_by_user_id = sqlc.arg(decided_by_user_id),
    decided_at = NOW()
WHERE id = sqlc.arg(id)
    AND circle_id = sqlc.arg(circle_id)
    AND status = 'pending'
RETURNING *;

-- name: CancelCircleJoinRequest :one
-- The requester withdrawing before anyone answered. decided_by_user_id
-- stays NULL: nobody in the Circle made a decision, so recording one
-- would misreport who acted.
UPDATE circle_join_requests
SET status = 'cancelled',
    decided_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND status = 'pending'
RETURNING *;
