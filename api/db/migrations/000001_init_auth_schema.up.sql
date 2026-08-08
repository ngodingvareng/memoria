-- =========================================================
-- Auth, account-level privacy controls, and data ownership.
-- =========================================================

CREATE TYPE auth_provider_id AS ENUM('google', 'credential');

-- Who may reach a user through a given social affordance. Used by
-- users.mention_policy and users.circle_invite_policy. 'circle_members'
-- means "only people who already share at least one Circle with me".
CREATE TYPE audience_policy AS ENUM('anyone', 'circle_members', 'nobody');

CREATE TYPE device_platform AS ENUM('ios', 'android', 'web');

CREATE TYPE data_export_status AS ENUM('pending', 'running', 'ready', 'failed', 'expired');


-- ---------------------------------------------------------
-- users
-- ---------------------------------------------------------
CREATE TABLE users(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,

    -- Handle used for @mentions and for "invite by entering their
    -- username" (Circle Invite). Stored as typed, matched
    -- case-insensitively via uq_users_username_lower.
    username VARCHAR(30) NOT NULL,

    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    image_path TEXT,

    -- Free-text self-description. Visible only to people who can already
    -- reach this user (circle co-members, and users they've mentioned or
    -- been mentioned by) — Memoria has no public profiles, so there is no
    -- surface where this is readable by a stranger.
    bio TEXT,

    -- IANA timezone name (e.g. 'Asia/Jakarta'). This is the user's
    -- *current* zone, used to decide when "today" starts for scheduling
    -- and digests. It is deliberately NOT what the archive is browsed
    -- in: each Moment carries the offset that was in effect where it
    -- happened (moments.occurred_utc_offset_minutes), so a trip to
    -- Tokyo stays on Tokyo dates after the user flies home.
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Social Interaction Controls (Privacy & Control).
    mention_policy audience_policy NOT NULL DEFAULT 'anyone',
    circle_invite_policy audience_policy NOT NULL DEFAULT 'anyone',

    -- Gates the "invite by entering their username" path specifically:
    -- a user who allows circle invites in general may still not want to
    -- be findable by typing their handle.
    discoverable_by_username BOOLEAN NOT NULL DEFAULT TRUE,

    -- Data Controls: strip EXIF/GPS from uploaded photos. Applied at
    -- upload time; moment_images.metadata_stripped records the outcome
    -- per file so a later policy change doesn't rewrite history.
    strip_photo_metadata BOOLEAN NOT NULL DEFAULT TRUE,

    -- Soft delete: gives account deletion a recovery grace period
    -- instead of instantly cascading a hard DELETE into the user's
    -- entire archive. A purge job hard-deletes the row (triggering the
    -- real CASCADE) once the grace period expires.
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT chk_users_username_format
    CHECK (username ~ '^[a-z0-9_.]{3,30}$'),

    -- TEXT is unbounded; without this a single user can store an
    -- arbitrarily large value. 500 chars is a bio, not a document.
    CONSTRAINT chk_users_bio_length
    CHECK (bio IS NULL OR length(bio) <= 500)
);

-- Case-insensitive uniqueness among live users only: a plain
-- UNIQUE(email) would let "User@x.com" and "user@x.com" register as two
-- different accounts, and would also block re-registering an email
-- that belongs to a soft-deleted (not yet purged) account.
CREATE UNIQUE INDEX uq_users_email_lower
ON users(lower(email))
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_users_username_lower
ON users(lower(username))
WHERE deleted_at IS NULL;

-- Supports a periodic purge job that hard-deletes users past their
-- grace period, e.g. WHERE deleted_at < NOW() - INTERVAL '30 days'.
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;


-- ---------------------------------------------------------
-- user_accounts
-- ---------------------------------------------------------
CREATE TABLE user_accounts(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id VARCHAR(255) NOT NULL,
    provider_id auth_provider_id NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    access_token_expires_at TIMESTAMPTZ,
    refresh_token_expires_at TIMESTAMPTZ,
    scope TEXT,
    id_token TEXT,
    password_hash TEXT,

    -- Per-account login lockout: after too many consecutive failed
    -- password attempts, a credential account is locked for a cooldown
    -- period regardless of which IP is attempting it. Complements
    -- IP-based rate limiting, which alone can't stop a distributed
    -- brute force against a single account.
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- password_hash must be set for 'credential' rows and empty for
    -- OAuth rows (google/github should never carry a local password).
    CONSTRAINT chk_user_accounts_credential_password
    CHECK (
        (provider_id = 'credential' AND password_hash IS NOT NULL)
        OR (provider_id <> 'credential' AND password_hash IS NULL)
    ),

    UNIQUE(provider_id, account_id)
);

CREATE INDEX idx_user_accounts_user_id ON user_accounts(user_id);


-- ---------------------------------------------------------
-- refresh_tokens
-- ---------------------------------------------------------
CREATE TABLE refresh_tokens(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Groups every token produced by rotating a given login into one
    -- chain. When a refresh token that has already been rotated out
    -- gets presented again (reuse/theft), the whole chain is revoked
    -- at once, not just that single row.
    family_id UUID NOT NULL DEFAULT gen_random_uuid(),

    -- Hash of the refresh token (e.g. SHA-256 hex digest), not the raw
    -- value. The raw value only ever exists on the client (cookie);
    -- the app hashes the incoming token with the same function before
    -- doing the lookup.
    token_hash VARCHAR(255) NOT NULL UNIQUE,

    expires_at TIMESTAMPTZ NOT NULL,

    -- NULL while still active; set once this token has been rotated
    -- (replaced by a new one) or explicitly logged out. A refresh
    -- request presenting a token whose revoked_at is already set is a
    -- reuse/theft signal.
    revoked_at TIMESTAMPTZ,

    -- Points to the row that replaced this one after rotation.
    replaced_by_id UUID REFERENCES refresh_tokens(id) ON DELETE SET NULL,

    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);

-- Supports a periodic cleanup job (DELETE FROM refresh_tokens WHERE
-- expires_at < NOW()); Postgres has no built-in TTL, so this needs
-- to be run by the app or a scheduler (e.g. pg_cron).
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);


-- ---------------------------------------------------------
-- user_verifications
-- ---------------------------------------------------------
CREATE TABLE user_verifications(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,

    -- Closes the replay hole logged as SCHEMA_REVIEW #10: validation
    -- checks consumed_at IS NULL and stamps it in the same statement,
    -- so two simultaneous requests can't both pass.
    consumed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Supports "find the pending verification for this identifier" lookups.
CREATE INDEX idx_user_verifications_identifier
ON user_verifications(identifier)
WHERE consumed_at IS NULL;

-- Supports the same periodic cleanup pattern as user_sessions.expires_at.
CREATE INDEX idx_user_verifications_expires_at ON user_verifications(expires_at);


-- ---------------------------------------------------------
-- user_devices
-- ---------------------------------------------------------
-- Push targets. Notifications are queued in `notifications` and
-- delivered to every live device row for the recipient.
CREATE TABLE user_devices(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform device_platform NOT NULL,
    push_token TEXT NOT NULL,

    -- The device's own zone, which may differ from users.timezone while
    -- travelling. Quiet hours are evaluated against this when present.
    timezone VARCHAR(50),

    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_user_devices_push_token
ON user_devices(push_token)
WHERE revoked_at IS NULL;

CREATE INDEX idx_user_devices_user_id ON user_devices(user_id) WHERE revoked_at IS NULL;


-- ---------------------------------------------------------
-- user_blocks
-- ---------------------------------------------------------
-- Blocking is symmetric in effect: neither user can see, mention,
-- comment on, or react to the other, in any context, in either
-- direction. Only one row is stored (the blocker's), so every access
-- check must test the pair both ways:
--
--   NOT EXISTS (SELECT 1 FROM user_blocks
--               WHERE (blocker_user_id = $viewer AND blocked_user_id = $owner)
--                  OR (blocker_user_id = $owner  AND blocked_user_id = $viewer))
--
-- Blocking never deletes content on either side (Principle 4).
CREATE TABLE user_blocks(
    blocker_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY(blocker_user_id, blocked_user_id),
    CONSTRAINT chk_user_blocks_not_self CHECK (blocker_user_id <> blocked_user_id)
);

CREATE INDEX idx_user_blocks_blocked_user_id ON user_blocks(blocked_user_id);


-- ---------------------------------------------------------
-- user_mutes
-- ---------------------------------------------------------
-- Muting is one-directional and purely a view filter: the muted user
-- is not restricted, is not notified, and their comments/reactions stay
-- visible to everyone else. Only the muter's own rendering changes.
CREATE TABLE user_mutes(
    muter_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    muted_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY(muter_user_id, muted_user_id),
    CONSTRAINT chk_user_mutes_not_self CHECK (muter_user_id <> muted_user_id)
);


-- ---------------------------------------------------------
-- data_exports
-- ---------------------------------------------------------
-- "Export all my data" — a complete, readable archive in an open
-- format, including photos at original quality. Generated
-- asynchronously; file_path points at object storage.
CREATE TABLE data_exports(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status data_export_status NOT NULL DEFAULT 'pending',
    file_path TEXT,
    byte_size BIGINT,
    error_message TEXT,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_data_exports_user_id ON data_exports(user_id, requested_at DESC);