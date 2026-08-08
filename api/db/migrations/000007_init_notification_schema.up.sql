-- =========================================================
-- Notification: an alert system that serves purely as a
-- reminder. Nothing here is real-time except the two cases
-- where a human being is actually waiting for an answer
-- (circle invite received, mentioned in a moment).
-- =========================================================

CREATE TYPE notification_kind AS ENUM(
    -- Commitment reminders — only for Threads carrying a Commitment,
    -- and part of what the user opted into.
    'commitment_upcoming',
    'moment_due',
    'moment_missed',

    -- Gifts, delivered when ready.
    'echo_ready',
    'recap_generated',

    -- Someone needs you, delivered immediately.
    'circle_invite_received',
    'mentioned_in_moment',

    -- Responses, batched into a single daily digest.
    'response_digest'
);


-- ---------------------------------------------------------
-- notification_preferences
-- ---------------------------------------------------------
-- One row per user; every type above can be enabled or disabled
-- independently. Column-per-type rather than key/value, so a missing
-- row can never silently mean "off" and so sqlc generates a plain
-- struct.
--
-- Defaults encode the spec exactly: gifts on, "someone needs you" on,
-- the response digest on at a fixed daily hour, Commitment reminders on
-- for Commitments only — with "Moment missed" off until asked for.
-- (The per-Commitment override lives on commitments.notify_missed;
-- both must be true for a miss notification to be delivered.)
CREATE TABLE notification_preferences(
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    commitment_upcoming_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    moment_due_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    moment_missed_enabled BOOLEAN NOT NULL DEFAULT FALSE,

    echo_ready_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    recap_generated_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    circle_invite_received_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    mentioned_in_moment_enabled BOOLEAN NOT NULL DEFAULT TRUE,

    response_digest_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    response_digest_hour SMALLINT NOT NULL DEFAULT 20,

    -- Quiet hours, in the user's local time. May wrap past midnight
    -- (e.g. 22:00 -> 07:00), so the check is a range test on the clock,
    -- not a simple comparison. Anything falling inside is held, not
    -- dropped, and re-scheduled via notifications.deliver_after.
    quiet_hours_start TIME,
    quiet_hours_end TIME,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_notification_preferences_digest_hour
    CHECK (response_digest_hour BETWEEN 0 AND 23),

    CONSTRAINT chk_notification_preferences_quiet_hours
    CHECK (num_nonnulls(quiet_hours_start, quiet_hours_end) <> 1)
);


-- ---------------------------------------------------------
-- notifications
-- ---------------------------------------------------------
-- One row per deliverable alert. Subject columns are nullable and
-- typed rather than a single polymorphic id, so every reference keeps
-- its foreign key and a deleted subject cannot leave a dangling
-- notification pointing at nothing.
CREATE TABLE notifications(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind notification_kind NOT NULL,

    -- The person who caused it, for 'mentioned_in_moment' and
    -- 'circle_invite_received'. NULL for gifts and reminders, which
    -- have no actor by design.
    actor_user_id UUID REFERENCES users(id) ON DELETE CASCADE,

    moment_id UUID REFERENCES moments(id) ON DELETE CASCADE,
    thread_id UUID REFERENCES threads(id) ON DELETE CASCADE,
    commitment_id UUID REFERENCES commitments(id) ON DELETE CASCADE,
    commitment_occurrence_id UUID REFERENCES commitment_occurrences(id) ON DELETE CASCADE,
    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,
    circle_invite_id UUID REFERENCES circle_invites(id) ON DELETE CASCADE,
    recap_id UUID REFERENCES recaps(id) ON DELETE CASCADE,
    resurfacing_id UUID REFERENCES resurfacings(id) ON DELETE CASCADE,

    -- Rendering data: the digest's counts, the Thread's name at send
    -- time, and so on. Never authoritative — always a snapshot for copy.
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Earliest permitted delivery instant. Quiet hours and the digest's
    -- fixed daily hour are both expressed by pushing this forward, so
    -- the sender is a single "deliver everything ready now" scan.
    deliver_after TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    delivered_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The sender's feed.
CREATE INDEX idx_notifications_pending
ON notifications(deliver_after)
WHERE delivered_at IS NULL;

-- The in-app notification list.
CREATE INDEX idx_notifications_user_created_at
ON notifications(user_id, created_at DESC);

-- At most one digest per user per local day. deliver_after::date alone is
-- rejected by Postgres (timestamptz -> date depends on the session
-- timezone, so it isn't IMMUTABLE); anchoring to a fixed zone first makes
-- the expression IMMUTABLE and safe to index.
CREATE UNIQUE INDEX uq_notifications_daily_digest
ON notifications(user_id, ((deliver_after AT TIME ZONE 'UTC')::date))
WHERE kind = 'response_digest';