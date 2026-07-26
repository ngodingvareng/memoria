-- =========================================================
-- Auth schema
-- Reviewed & consolidated version
-- =========================================================

CREATE TYPE auth_provider_id AS ENUM('google', 'github', 'credential');

-- ---------------------------------------------------------
-- users
-- ---------------------------------------------------------
CREATE TABLE users(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    image_path TEXT,

    -- IANA timezone name (e.g. 'Asia/Jakarta'). Used to bucket
    -- occurred_at/scheduled_at into "days" correctly for the heatmap and
    -- other statistics — without this there's no way to know whose day
    -- boundary to use when converting an absolute TIMESTAMPTZ to a date.
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Soft delete: lets account deletion have a recovery grace period
    -- instead of instantly cascading a hard DELETE into activities /
    -- activity_captures (which would permanently erase the user's history).
    -- The app should hard-delete the row (triggering the real CASCADE)
    -- only after the grace period via a purge job.
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

-- Case-insensitive uniqueness among live users only: a plain
-- UNIQUE(email) would let "User@x.com" and "user@x.com" register as two
-- different accounts, and would also block re-registering an email
-- that belongs to a soft-deleted (not yet purged) account.
CREATE UNIQUE INDEX uq_users_email_lower
ON users(lower(email))
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
-- user_sessions
-- ---------------------------------------------------------
CREATE TABLE user_sessions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Hash of the session token (e.g. SHA-256 hex digest), not the raw
    -- token. The raw token is only ever held by the client (cookie); the
    -- app hashes the incoming token with the same function before doing
    -- the lookup. This way a DB leak doesn't hand out usable sessions.
    token_hash VARCHAR(255) NOT NULL UNIQUE,

    expires_at TIMESTAMPTZ NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);

-- Supports a periodic cleanup job (DELETE FROM user_sessions WHERE
-- expires_at < NOW()); Postgres has no built-in TTL, so this needs
-- to be run by the app or a scheduler (e.g. pg_cron).
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);


-- ---------------------------------------------------------
-- user_verifications
-- ---------------------------------------------------------
CREATE TABLE user_verifications(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Supports "find the pending verification for this identifier" lookups.
CREATE INDEX idx_user_verifications_identifier ON user_verifications(identifier);

-- Supports the same periodic cleanup pattern as user_sessions.expires_at.
CREATE INDEX idx_user_verifications_expires_at ON user_verifications(expires_at);
