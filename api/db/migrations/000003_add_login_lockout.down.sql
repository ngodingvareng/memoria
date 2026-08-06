ALTER TABLE user_accounts
    DROP COLUMN failed_login_attempts,
    DROP COLUMN locked_until;
