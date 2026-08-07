CREATE TYPE moment_status AS ENUM('due', 'kept', 'missed');

-- ---------------------------------------------------------
-- threads
-- ---------------------------------------------------------
CREATE TABLE threads(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    has_commitment BOOLEAN NOT NULL DEFAULT true,
    color_hex VARCHAR(7),
    confirmation_timeout_minutes INT DEFAULT 1440,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT chk_threads_color_hex
    CHECK (color_hex ~ '^#[0-9A-Fa-f]{6}$'),
    CONSTRAINT chk_threads_timeout_positive
    CHECK (confirmation_timeout_minutes > 0)
);

CREATE INDEX idx_threads_user_id ON threads(user_id);


-- ---------------------------------------------------------
-- thread_images
-- ---------------------------------------------------------
-- Images attached to the thread itself (e.g. a cover photo / gallery
-- describing the thread in general), as opposed to moment_images which
-- are photos attached to a specific captured Moment.
CREATE TABLE thread_images(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    image_path TEXT NOT NULL,
    image_alt VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_thread_images_thread_id ON thread_images(thread_id);


-- ---------------------------------------------------------
-- commitments
-- ---------------------------------------------------------
-- No unique constraint on thread_id here: a thread may intentionally
-- carry more than one concurrently active Commitment (e.g. a "morning"
-- cron and an "evening" cron with different expressions/timezones for
-- the same thread). Each row is one currently-active Commitment; when a
-- specific Commitment is changed or retired, the app should copy its
-- current values into commitment_histories (active_from = this row's
-- created_at, active_until = now) before updating/removing this row.
CREATE TABLE commitments(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    cron_expression VARCHAR(255) NOT NULL,
    timezone VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Lets moments enforce, via a composite FK, that a linked
    -- commitment actually belongs to the same thread as the moment.
    CONSTRAINT uq_commitments_thread_id_id UNIQUE(thread_id, id)
);

CREATE INDEX idx_commitments_thread_id ON commitments(thread_id);


-- ---------------------------------------------------------
-- commitment_histories
-- ---------------------------------------------------------
CREATE TABLE commitment_histories(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    commitment_id UUID REFERENCES commitments(id) ON DELETE SET NULL,
    cron_expression VARCHAR(255) NOT NULL,
    timezone VARCHAR(50),
    active_from TIMESTAMPTZ NOT NULL,
    active_until TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- note: DEFAULT NOW() only makes sense if a row is inserted exactly
    -- at the moment the old commitment is superseded. Double-check this
    -- against the application write path.
);

CREATE INDEX idx_commitment_histories_thread_id ON commitment_histories(thread_id);
CREATE INDEX idx_commitment_histories_commitment_id ON commitment_histories(commitment_id);


-- ---------------------------------------------------------
-- moments
-- ---------------------------------------------------------
CREATE TABLE moments(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,

    -- Which active Commitment generated this Moment, if any. A thread can
    -- have more than one concurrently active Commitment (e.g. "morning"
    -- and "evening" with different cron expressions), so this tells them
    -- apart. NULL for manually-captured Moments (has_commitment = false).
    -- No inline REFERENCES here: enforced below via a composite FK
    -- together with thread_id, so a linked commitment is guaranteed to
    -- belong to the same thread as this Moment.
    commitment_id UUID,

    status moment_status NOT NULL DEFAULT 'due',

    -- When the Commitment expected/generated this Moment. Drives the
    -- confirmation-window logic and Commitment adherence stats. NULL for
    -- manually-captured Moments.
    due_at TIMESTAMPTZ,

    -- When the user actually recorded/confirmed the Moment.
    captured_at TIMESTAMPTZ,

    -- When the moment actually took place, as reported by the user.
    -- For Commitment-generated Moments this is set equal to due_at at
    -- creation time. For manual Moments the user provides it directly,
    -- and it may be backdated (differs from created_at). This is the
    -- field Rhythms/heatmap should bucket by, since it uniformly
    -- represents "when it happened" for both Moment origins.
    occurred_at TIMESTAMPTZ NOT NULL,

    content TEXT,
    color_hex VARCHAR(7),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT chk_moments_color_hex
    CHECK (color_hex ~ '^#[0-9A-Fa-f]{6}$'),

    -- Requires PostgreSQL 15+: the column list on SET NULL means only
    -- commitment_id gets nulled when the referenced commitment is
    -- deleted — without it, Postgres would try to null thread_id too
    -- (since it's part of this composite FK), which would fail because
    -- thread_id is NOT NULL.
    CONSTRAINT fk_moments_commitment_thread
    FOREIGN KEY(thread_id, commitment_id)
    REFERENCES commitments(thread_id, id)
    ON DELETE SET NULL (commitment_id)
);

CREATE INDEX idx_moments_commitment_id ON moments(commitment_id);

-- Supports heatmap / history queries that bucket Moments by the date
-- they actually took place ("when it happened"), for both
-- Commitment-generated and manual (possibly backdated) Moments alike.
CREATE INDEX idx_moments_occurred_at
ON moments(occurred_at)
WHERE deleted_at IS NULL;

-- Uniqueness of (thread_id, due_at) among "live" rows only — prevents
-- the scheduler from generating two Moments for the same thread at the
-- same due occurrence (even across concurrently active Commitments,
-- e.g. "morning" and "evening", if they ever computed the same
-- instant). Postgres treats NULL <> NULL, so rows with due_at IS NULL
-- (manual Moments) never conflict with each other.
CREATE UNIQUE INDEX uq_moments_thread_id_due_at
ON moments(thread_id, due_at)
WHERE deleted_at IS NULL;


-- ---------------------------------------------------------
-- moment_images
-- ---------------------------------------------------------
CREATE TABLE moment_images(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    image_path TEXT NOT NULL,
    image_alt VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_moment_images_moment_id ON moment_images(moment_id);
