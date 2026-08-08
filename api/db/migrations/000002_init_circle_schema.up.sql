-- =========================================================
-- Circles: private social groups. Sharing is relational,
-- never broadcast (Principle 5) — a Circle is the only
-- many-person audience that exists in the product.
-- =========================================================

CREATE TYPE circle_role AS ENUM('admin', 'member');

CREATE TYPE circle_invite_kind AS ENUM('username', 'link');

-- The vocabulary is split by kind, and chk_circle_invites_status_by_kind
-- enforces the split: a username invite is answered ('pending' until
-- 'accepted'/'declined', or ended by 'revoked'/'expired'), while a link
-- is only ever live or replaced ('active' -> 'revoked'). A link is not
-- pending anything, and it cannot be declined or expire.
CREATE TYPE circle_invite_status AS ENUM('pending', 'active', 'accepted', 'declined', 'revoked', 'expired');

CREATE TYPE circle_join_request_status AS ENUM('pending', 'approved', 'rejected', 'cancelled');


-- ---------------------------------------------------------
-- circles
-- ---------------------------------------------------------
CREATE TABLE circles(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color_hex VARCHAR(7),
    image_path TEXT,

    -- The creator, who becomes the first admin. SET NULL rather than
    -- CASCADE: one member deleting their account must not dissolve a
    -- Circle everyone else is still using.
    created_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Dissolution is a soft delete first, so the cascade below can be
    -- previewed and undone during a grace period. Hard-deleting the row
    -- is what actually enacts the lifecycle rules:
    --   * circle-owned threads      -> CASCADE (container disappears)
    --   * moments in those threads  -> thread_id SET NULL, so each one
    --                                  returns to its capturer's archive
    --   * personal moments shared in-> moment_circles CASCADE, moment
    --                                  stays personal and untouched
    --   * comments/reactions scoped
    --     to this circle            -> CASCADE (audience no longer exists)
    dissolved_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX idx_circles_dissolved_at ON circles(dissolved_at) WHERE dissolved_at IS NOT NULL;


-- ---------------------------------------------------------
-- circle_members
-- ---------------------------------------------------------
-- Leaving a Circle only affects the future: the user stops seeing new
-- activity and can no longer post, comment, or react — but nothing they
-- contributed is removed. That is why leaving stamps left_at instead of
-- deleting the row: the historical fact of membership is what lets a
-- past comment still render with its Circle badge.
--
-- Active membership is therefore always (left_at IS NULL), never mere
-- row existence.
CREATE TABLE circle_members(
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role circle_role NOT NULL DEFAULT 'member',

    -- Non-admins have no invite privilege by default, but can capture
    -- into the Circle's collaborative Threads by default.
    can_invite BOOLEAN NOT NULL DEFAULT FALSE,
    can_capture BOOLEAN NOT NULL DEFAULT TRUE,

    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ DEFAULT NULL,

    PRIMARY KEY(circle_id, user_id),

    -- An admin always carries invite privilege; the flag exists so the
    -- two never have to be checked separately at the query layer.
    CONSTRAINT chk_circle_members_admin_can_invite
    CHECK (role <> 'admin' OR can_invite = TRUE)
);

CREATE INDEX idx_circle_members_user_id
ON circle_members(user_id)
WHERE left_at IS NULL;

CREATE INDEX idx_circle_members_circle_id
ON circle_members(circle_id)
WHERE left_at IS NULL;


-- ---------------------------------------------------------
-- circle_invites
-- ---------------------------------------------------------
-- Two shapes, deliberately asymmetric per the spec:
--
--   kind = 'username' -> the exception path of a direct add. The normal
--                        direct add writes nothing here: the named users
--                        go straight into circle_members. A row appears
--                        only when the recipient's circle_invite_policy
--                        holds the add back for their confirmation, and
--                        that pending row is the one thing in this table
--                        with a lifetime.
--   kind = 'link'      -> the Circle's shareable token. Exactly one is
--                        live per Circle, it never expires, and it has no
--                        usage limit — rotation is the whole revocation
--                        model. requires_approval decides whether people
--                        arriving through it join outright or open a
--                        circle_join_requests row first.
CREATE TABLE circle_invites(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    invited_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    kind circle_invite_kind NOT NULL,

    invitee_user_id UUID REFERENCES users(id) ON DELETE CASCADE,

    -- Hash of the link token, never the raw value. Deterministic
    -- (SHA-256, as with refresh_tokens.token_hash) because the join path
    -- has to find the row by it.
    token_hash VARCHAR(255),

    -- Link invites only: whether following the link joins the Circle
    -- outright or opens a join request first. Defaults to requiring
    -- approval, unlike a group chat — a link travels anywhere, so
    -- admitting strangers on its own is opted into, never out of.
    requires_approval BOOLEAN,

    status circle_invite_status NOT NULL DEFAULT 'pending',

    -- Username invites only; a link has no lifetime by design.
    expires_at TIMESTAMPTZ,

    -- When the invite left its open state, whichever way it left:
    -- answered, revoked, or expired.
    responded_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_circle_invites_shape
    CHECK (
        (kind = 'username' AND invitee_user_id IS NOT NULL
         AND token_hash IS NULL AND requires_approval IS NULL)
        OR
        (kind = 'link' AND invitee_user_id IS NULL
         AND token_hash IS NOT NULL AND requires_approval IS NOT NULL
         AND expires_at IS NULL)
    ),

    CONSTRAINT chk_circle_invites_status_by_kind
    CHECK (
        (kind = 'username'
         AND status IN ('pending', 'accepted', 'declined', 'revoked', 'expired'))
        OR
        (kind = 'link' AND status IN ('active', 'revoked'))
    )
);

CREATE UNIQUE INDEX uq_circle_invites_token_hash
ON circle_invites(token_hash)
WHERE token_hash IS NOT NULL;

-- At most one outstanding username invite per (circle, invitee).
CREATE UNIQUE INDEX uq_circle_invites_pending_username
ON circle_invites(circle_id, invitee_user_id)
WHERE kind = 'username' AND status = 'pending';

-- Exactly one live link per Circle. Regenerating therefore has to revoke
-- the current row in the same transaction as it inserts the new one,
-- which makes "the old link stops working" a property of the schema
-- rather than something every caller has to remember.
CREATE UNIQUE INDEX uq_circle_invites_active_link
ON circle_invites(circle_id)
WHERE kind = 'link' AND status = 'active';

CREATE INDEX idx_circle_invites_invitee_user_id
ON circle_invites(invitee_user_id)
WHERE status = 'pending';


-- ---------------------------------------------------------
-- circle_join_requests
-- ---------------------------------------------------------
-- The confirmation step for a link whose requires_approval is TRUE. A
-- link without it, and every direct add, never produce a row here.
--
-- Rotating a Circle's link leaves requests already open against the old
-- one untouched: the request was to join the Circle, not to use a
-- particular string, and an invite-privileged member still has to
-- approve each one by hand — so a leaked link cannot convert them.
CREATE TABLE circle_join_requests(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_id UUID REFERENCES circle_invites(id) ON DELETE SET NULL,
    status circle_join_request_status NOT NULL DEFAULT 'pending',
    decided_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_circle_join_requests_pending
ON circle_join_requests(circle_id, user_id)
WHERE status = 'pending';

CREATE INDEX idx_circle_join_requests_circle_id
ON circle_join_requests(circle_id)
WHERE status = 'pending';