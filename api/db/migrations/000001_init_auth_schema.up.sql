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
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Supports "find the pending verification for this identifier" lookups.
CREATE INDEX idx_user_verifications_identifier ON user_verifications(identifier);

-- Supports the same periodic cleanup pattern as user_sessions.expires_at.
CREATE INDEX idx_user_verifications_expires_at ON user_verifications(expires_at);
