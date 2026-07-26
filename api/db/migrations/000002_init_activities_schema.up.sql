CREATE TYPE activity_capture_status AS ENUM('awaiting', 'captured', 'missed');

-- ---------------------------------------------------------
-- activities
-- ---------------------------------------------------------
CREATE TABLE activities(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_fixed_schedule BOOLEAN NOT NULL DEFAULT true,
    color_hex VARCHAR(7),
    confirmation_timeout_minutes INT DEFAULT 1440,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT chk_activities_color_hex
    CHECK (color_hex ~ '^#[0-9A-Fa-f]{6}$'),
    CONSTRAINT chk_activities_timeout_positive
    CHECK (confirmation_timeout_minutes > 0)
);

CREATE INDEX idx_activities_user_id ON activities(user_id);


-- ---------------------------------------------------------
-- activity_images
-- ---------------------------------------------------------
-- Images attached to the activity itself (e.g. a cover photo / gallery
-- describing the activity in general), as opposed to activity_capture_images
-- which are moments captured at a specific occurrence.
CREATE TABLE activity_images(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    image_path TEXT NOT NULL,
    image_alt VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_images_activity_id ON activity_images(activity_id);


-- ---------------------------------------------------------
-- activity_schedules
-- ---------------------------------------------------------
-- No unique constraint on activity_id here: an activity may intentionally
-- have more than one concurrently active schedule (e.g. a "pagi" cron and
-- a "sore" cron with different expressions/timezones for the same
-- activity). Each row is one currently-active schedule; when a specific
-- schedule is changed or retired, the app should copy its current values
-- into activity_schedule_histories (active_from = this row's created_at,
-- active_until = now) before updating/removing this row.
CREATE TABLE activity_schedules(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    cron_expression VARCHAR(255) NOT NULL,
    timezone VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- fixed: removed trailing comma that caused a syntax error

    -- Lets activity_captures enforce, via a composite FK, that a linked
    -- schedule actually belongs to the same activity as the item.
    CONSTRAINT uq_activity_schedules_activity_id_id UNIQUE(activity_id, id)
);

CREATE INDEX idx_activity_schedules_activity_id ON activity_schedules(activity_id);


-- ---------------------------------------------------------
-- activity_schedule_histories
-- ---------------------------------------------------------
CREATE TABLE activity_schedule_histories(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    schedule_id UUID REFERENCES activity_schedules(id) ON DELETE SET NULL,
    cron_expression VARCHAR(255) NOT NULL,
    timezone VARCHAR(50),
    active_from TIMESTAMPTZ NOT NULL,
    active_until TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- note: DEFAULT NOW() only makes sense if a row is inserted exactly
    -- at the moment the old schedule is superseded. Double-check this
    -- against the application write path.
);

CREATE INDEX idx_activity_schedule_histories_activity_id ON activity_schedule_histories(activity_id);
CREATE INDEX idx_activity_schedule_histories_schedule_id ON activity_schedule_histories(schedule_id);


-- ---------------------------------------------------------
-- activity_captures
-- ---------------------------------------------------------
CREATE TABLE activity_captures(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,

    -- Which active schedule generated this item, if any. An activity can
    -- have more than one concurrently active schedule (e.g. "pagi" and
    -- "sore" with different cron expressions), so this tells them apart.
    -- NULL for manually-created items (is_fixed_schedule = false).
    -- No inline REFERENCES here: enforced below via a composite FK
    -- together with activity_id, so a linked schedule is guaranteed to
    -- belong to the same activity as this item.
    schedule_id UUID,

    status activity_capture_status NOT NULL DEFAULT 'awaiting',

    -- When the system expected/generated this item (fixed-schedule only).
    -- Drives the confirmation-timeout logic and the "muncul -> konfirmasi"
    -- diff chart. NULL for manually-created items.
    scheduled_at TIMESTAMPTZ,

    -- When the user actually confirmed/checked the item off.
    confirmed_at TIMESTAMPTZ,

    -- When the activity actually took place, as reported by the user.
    -- For fixed-schedule items this is set equal to scheduled_at at
    -- creation time. For manual items the user provides it directly,
    -- and it may be backdated (differs from created_at). This is the
    -- field statistics/heatmap should bucket by, since it uniformly
    -- represents "kapan dilaksanakan" for both item types.
    occurred_at TIMESTAMPTZ NOT NULL,

    content TEXT,
    color_hex VARCHAR(7),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT chk_activity_captures_color_hex
    CHECK (color_hex ~ '^#[0-9A-Fa-f]{6}$'),

    -- Requires PostgreSQL 15+: the column list on SET NULL means only
    -- schedule_id gets nulled when the referenced schedule is deleted —
    -- without it, Postgres would try to null activity_id too (since it's
    -- part of this composite FK), which would fail because activity_id
    -- is NOT NULL.
    CONSTRAINT fk_activity_captures_schedule_activity
    FOREIGN KEY(activity_id, schedule_id)
    REFERENCES activity_schedules(activity_id, id)
    ON DELETE SET NULL (schedule_id)
);

CREATE INDEX idx_activity_captures_schedule_id ON activity_captures(schedule_id);

-- Supports heatmap / history queries that bucket items by the date they
-- actually took place ("kapan dilaksanakan"), for both fixed-schedule
-- and manual (possibly backdated) items alike.
CREATE INDEX idx_activity_captures_occurred_at
ON activity_captures(occurred_at)
WHERE deleted_at IS NULL;

-- Uniqueness of (activity_id, scheduled_at) among "live" rows only —
-- prevents the scheduler from generating two items for the same
-- activity at the same scheduled occurrence (even across concurrently
-- active schedules, e.g. "pagi" and "sore", if they ever computed the
-- same instant). Postgres treats NULL <> NULL, so rows with
-- scheduled_at IS NULL (manual items) never conflict with each other.
CREATE UNIQUE INDEX uq_activity_captures_activity_scheduled_at
ON activity_captures(activity_id, scheduled_at)
WHERE deleted_at IS NULL;


-- ---------------------------------------------------------
-- activity_capture_images
-- ---------------------------------------------------------
CREATE TABLE activity_capture_images(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_capture_id UUID NOT NULL REFERENCES activity_captures(id) ON DELETE CASCADE,
    image_path TEXT NOT NULL,
    image_alt VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_capture_images_activity_capture_id ON activity_capture_images(activity_capture_id);
