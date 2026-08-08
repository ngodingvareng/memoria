-- =========================================================
-- Commitment: the opt-in mode where a Thread takes on a
-- schedule and Memoria keeps an honest record of whether the
-- user kept it. A deliberate, bounded exception to
-- Principle 2 — which is why consent is a NOT NULL column
-- rather than a settings toggle.
--
-- This file is where the Commitment Firewall is enforced.
-- Schedule outcomes live here, in commitment_occurrences,
-- and never in `moments`. A `missed` occurrence has no
-- moment_id and no moments row, so browsing March 2024 shows
-- the mornings the user ran and structurally cannot show a
-- column of the ones they didn't.
-- =========================================================

CREATE TYPE commitment_strictness AS ENUM('gentle', 'standard', 'strict');

CREATE TYPE commitment_outcome AS ENUM('due', 'kept', 'missed');


-- ---------------------------------------------------------
-- commitments
-- ---------------------------------------------------------
-- A Thread may carry more than one, and each is fully independent:
-- its own recurrence rule, its own confirmation window, its own
-- strictness. "Morning Run" and "Evening Walk" under one Thread are two
-- Commitments, not one schedule with two times.
CREATE TABLE commitments(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,

    -- Distinguishes sibling Commitments under one Thread in the UI and
    -- in notification copy ("'Morning Run' is coming up soon").
    name VARCHAR(255) NOT NULL,

    cron_expression VARCHAR(255) NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',

    strictness commitment_strictness NOT NULL DEFAULT 'standard',

    -- Per Commitment, not per Thread. Strictness affects only this
    -- window and how adherence is counted — never how a Moment is
    -- stored.
    confirmation_window_minutes INT NOT NULL DEFAULT 1440,

    -- ----- The opt-in contract -----
    -- NOT NULL by design: a Commitment cannot come into existence
    -- without a recorded, explicit, informed consent event. consent_version
    -- pins which wording the user actually agreed to, so changing the
    -- copy later does not silently re-attribute consent.
    consented_at TIMESTAMPTZ NOT NULL,
    consent_version SMALLINT NOT NULL DEFAULT 1,

    -- Notification choices scoped to this Commitment, part of what the
    -- user opted into. notify_missed is off by default and is a second,
    -- separate choice — the record of the miss is kept regardless; this
    -- only decides whether the user is told that evening.
    notify_upcoming BOOLEAN NOT NULL DEFAULT TRUE,
    notify_due BOOLEAN NOT NULL DEFAULT TRUE,
    notify_missed BOOLEAN NOT NULL DEFAULT FALSE,

    -- The exit ramp. Pausing and turning off are ordinary choices, and
    -- both stop generation without touching a single past occurrence.
    -- Together they replace threads.has_commitment: the generator feed
    -- is simply the set of rows where both are NULL and the Thread is
    -- live, so there is no denormalized flag left to drift.
    paused_at TIMESTAMPTZ,
    archived_at TIMESTAMPTZ,

    -- Scheduler bookmark: the last occurrence instant already
    -- generated, so a restarted worker resumes instead of replaying.
    last_generated_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Optimistic locking for concurrent edits (SCHEMA_REVIEW #3): the
    -- update guards on the version it read and bumps it, so a
    -- double-submitted form produces one archived history row, not two
    -- overlapping ones.
    version INT NOT NULL DEFAULT 1,

    CONSTRAINT chk_commitments_window_positive
    CHECK (confirmation_window_minutes > 0),

    -- Lets commitment_occurrences enforce, via a composite FK, that an
    -- occurrence's commitment really belongs to its thread.
    CONSTRAINT uq_commitments_thread_id_id UNIQUE(thread_id, id)
);

-- The scheduler feed, replacing the old
-- "JOIN threads ... WHERE has_commitment = TRUE".
CREATE INDEX idx_commitments_active
ON commitments(thread_id)
WHERE paused_at IS NULL AND archived_at IS NULL;


-- ---------------------------------------------------------
-- commitment_histories
-- ---------------------------------------------------------
-- An append-only snapshot of each superseded version of a Commitment,
-- written in the same transaction as the UPDATE that replaces it.
--
-- Edit in place via UPDATE, never delete-and-reinsert: occurrences
-- reference commitments by id, and a delete would sever every one of
-- them (SCHEMA_REVIEW #6).
--
-- This table answers "what were the rules in March?" for adherence
-- recomputation. It is not what freezes an individual occurrence's
-- verdict — commitment_occurrences carries its own frozen copy of the
-- rule it was judged under.
CREATE TABLE commitment_histories(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    commitment_id UUID REFERENCES commitments(id) ON DELETE SET NULL,

    cron_expression VARCHAR(255) NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    strictness commitment_strictness NOT NULL,
    confirmation_window_minutes INT NOT NULL,

    active_from TIMESTAMPTZ NOT NULL,
    active_until TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_commitment_histories_interval
    CHECK (active_until >= active_from)
);

CREATE INDEX idx_commitment_histories_commitment_id
ON commitment_histories(commitment_id, active_from DESC);

CREATE INDEX idx_commitment_histories_thread_id
ON commitment_histories(thread_id, active_from DESC);


-- ---------------------------------------------------------
-- commitment_occurrences
-- ---------------------------------------------------------
-- One row per scheduled occurrence. This is the Firewall.
--
--   outcome = 'due'    -> generated, not yet answered. Lives in the inbox.
--   outcome = 'kept'   -> answered within the window; moment_id points
--                         at the Moment that was recorded.
--   outcome = 'missed' -> the window closed with no answer.
--
-- A missed occurrence has no Moment, so Album, Echoes, Search, and a
-- Thread's timeline — which all query `moments` — cannot surface it,
-- today or from any endpoint written later.
--
-- Aggregate adherence, by contrast, reads this table directly and is
-- welcome in Rhythms and Recap. Aggregates belong in the archive;
-- individual failures do not.
CREATE TABLE commitment_occurrences(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    commitment_id UUID NOT NULL,

    -- Denormalized so the composite FK below can prove the commitment
    -- belongs to this thread, and so adherence can be grouped by Thread
    -- without a join.
    thread_id UUID NOT NULL,

    -- Who owes it. Kept here so account deletion cascades cleanly and
    -- so adherence never has to walk threads -> circles to find a user.
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    outcome commitment_outcome NOT NULL DEFAULT 'due',

    -- When it came due, with the same instant/local/offset triple the
    -- archive uses, so adherence-by-day-of-week and
    -- adherence-by-time-of-day are computed in the zone the Commitment
    -- was actually scheduled in.
    due_at TIMESTAMPTZ NOT NULL,
    due_local TIMESTAMP NOT NULL,
    due_utc_offset_minutes SMALLINT NOT NULL DEFAULT 0,
    due_on DATE GENERATED ALWAYS AS (due_local::date) STORED,

    -- Frozen at generation from the Commitment's window *at that time*,
    -- rather than recomputed from the current setting. This is what
    -- makes "editing a Commitment affects future Moments only" true:
    -- shortening the window from 1440 to 60 minutes can no longer flip
    -- a pile of old still-due occurrences to missed on the next worker
    -- run (SCHEMA_REVIEW #7).
    expires_at TIMESTAMPTZ NOT NULL,

    -- Frozen for the same reason. Changing strictness never rewrites
    -- history; keeping the rule each occurrence was judged under is
    -- also what makes the middle path of Open Decision 2a possible —
    -- recomputing adherence *rates* under current strictness while
    -- individual outcomes stay exactly as they were.
    strictness_at_due commitment_strictness NOT NULL,

    -- When the user actually answered. May be after expires_at: a
    -- missed occurrence can still be recorded later, at any time, and
    -- being slow never locks anyone out.
    captured_at TIMESTAMPTZ,

    -- Under 'standard', a late recording counts as kept but is noted as
    -- late; under 'strict' only on-time recordings count. Generated so
    -- the note can never disagree with the timestamps.
    was_late BOOLEAN GENERATED ALWAYS AS (
        captured_at IS NOT NULL AND captured_at > expires_at
    ) STORED,

    -- The Moment recorded in answer to this occurrence, if any.
    -- SET NULL, not CASCADE: deleting the photo does not un-run the
    -- run. The user deleting their Moment removes it everywhere, but
    -- the settled fact that they showed up that day stays in the
    -- aggregate.
    moment_id UUID UNIQUE REFERENCES moments(id) ON DELETE SET NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_commitment_occurrences_commitment_thread
    FOREIGN KEY(thread_id, commitment_id)
    REFERENCES commitments(thread_id, id)
    ON DELETE CASCADE,

    -- A still-due occurrence has not been answered.
    CONSTRAINT chk_commitment_occurrences_due_unanswered
    CHECK (outcome <> 'due' OR captured_at IS NULL),

    -- 'kept' means answered. ('missed' may or may not carry a
    -- captured_at — that is exactly the late-recording case.)
    CONSTRAINT chk_commitment_occurrences_kept_answered
    CHECK (outcome <> 'kept' OR captured_at IS NOT NULL),

    CONSTRAINT chk_commitment_occurrences_expiry
    CHECK (expires_at > due_at)
);

-- Idempotent generation: a second scheduler instance computing the same
-- occurrence loses the race instead of duplicating it. Paired with
-- ON CONFLICT DO NOTHING in the insert.
CREATE UNIQUE INDEX uq_commitment_occurrences_commitment_due_at
ON commitment_occurrences(commitment_id, due_at);

-- The inbox: everything still awaiting an answer.
CREATE INDEX idx_commitment_occurrences_inbox
ON commitment_occurrences(user_id, due_at DESC)
WHERE outcome = 'due';

-- The timeout worker's feed: due and past its own frozen deadline.
CREATE INDEX idx_commitment_occurrences_expiring
ON commitment_occurrences(expires_at)
WHERE outcome = 'due';

-- Adherence: kept vs. due over a period, per Commitment, bucketed by
-- the local date it came due.
CREATE INDEX idx_commitment_occurrences_adherence
ON commitment_occurrences(commitment_id, due_on);