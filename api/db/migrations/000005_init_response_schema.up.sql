-- =========================================================
-- Response: how a shared Moment gets answered.
--
-- Both tables carry a circle_id that names the *audience* a
-- response was made to, because visibility stays scoped to
-- context: a mentioned user who isn't in a Circle the Moment
-- was also shared to cannot see that Circle's comments or
-- reactions, even though they can see the Moment itself.
--
--   circle_id = <id>  -> made inside that Circle; visible to
--                        its members only, and removed with
--                        the Circle when it dissolves.
--   circle_id = NULL  -> the mention context: visible to the
--                        Moment's creator and its mentioned
--                        users. Survives any Circle's fate.
-- =========================================================

-- The fixed set: ❤️ 😂 🥹. Stored by name rather than by codepoint so
-- the rendered glyph can be changed without a data migration.
CREATE TYPE reaction_kind AS ENUM('heart', 'joy', 'tender');

CREATE TYPE response_event_kind AS ENUM('comment', 'reaction');


-- ---------------------------------------------------------
-- comments
-- ---------------------------------------------------------
-- Flat, permanently. There is no parent_id, reply_to_id, or depth
-- column here and there never will be one: "Thread" is the domain
-- entity, and flat comments are what keep that word unambiguous
-- everywhere else in the product.
CREATE TABLE comments(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,

    -- SET NULL, not CASCADE. When a user deletes their account, their
    -- comments on *other people's* Moments are anonymized, not deleted,
    -- so the discussion stays coherent. A NULL author renders as a
    -- former member — deliberately without a name snapshot, since
    -- anonymized means anonymized.
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,

    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT chk_comments_body_not_empty CHECK (length(btrim(body)) > 0)
);

CREATE INDEX idx_comments_moment_id
ON comments(moment_id, created_at)
WHERE deleted_at IS NULL;

CREATE INDEX idx_comments_user_id ON comments(user_id) WHERE user_id IS NOT NULL;


-- ---------------------------------------------------------
-- reactions
-- ---------------------------------------------------------
-- No counter column exists anywhere in this schema, on purpose. Who
-- reacted is a row; how many is never materialized and never returned.
CREATE TABLE reactions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,
    kind reaction_kind NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One reaction per person per Moment per audience. NULLS NOT DISTINCT
-- makes the mention context (circle_id IS NULL) behave like any other
-- single audience instead of allowing unlimited duplicates; the partial
-- predicate exempts anonymized rows, which must be allowed to coexist
-- after several users delete their accounts.
CREATE UNIQUE INDEX uq_reactions_user_moment_circle
ON reactions(moment_id, user_id, circle_id) NULLS NOT DISTINCT
WHERE user_id IS NOT NULL;

CREATE INDEX idx_reactions_moment_id ON reactions(moment_id);


-- ---------------------------------------------------------
-- response_events
-- ---------------------------------------------------------
-- The queue behind the daily response digest. Reactions and comments
-- are never delivered individually — instant per-reaction pings are
-- precisely the engagement mechanic Memoria exists as an alternative
-- to. The digest job collects everything with digested_at IS NULL for a
-- recipient, emits one notification, and stamps the rows.
--
-- Batching preserves the information and discards the slot machine.
CREATE TABLE response_events(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- The Moment's owner, i.e. who gets told.
    recipient_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,

    kind response_event_kind NOT NULL,
    comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    reaction_id UUID REFERENCES reactions(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    digested_at TIMESTAMPTZ,

    CONSTRAINT chk_response_events_shape
    CHECK (
        (kind = 'comment' AND comment_id IS NOT NULL AND reaction_id IS NULL)
        OR (kind = 'reaction' AND reaction_id IS NOT NULL AND comment_id IS NULL)
    )
);

CREATE INDEX idx_response_events_pending
ON response_events(recipient_user_id, created_at)
WHERE digested_at IS NULL;