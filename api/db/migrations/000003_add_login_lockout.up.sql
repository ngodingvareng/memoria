-- ---------------------------------------------------------
-- Per-account login lockout: after too many consecutive failed
-- password attempts, a credential account is locked for a cooldown
-- period regardless of which IP is attempting it. Complements
-- IP-based rate limiting, which alone can't stop a distributed
-- brute force against a single account.
-- ---------------------------------------------------------
ALTER TABLE user_accounts
    ADD COLUMN failed_login_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN locked_until TIMESTAMPTZ;
