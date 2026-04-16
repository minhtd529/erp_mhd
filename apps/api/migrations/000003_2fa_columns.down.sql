ALTER TABLE two_factor_challenges
    DROP COLUMN IF EXISTS invalidated_at,
    DROP COLUMN IF EXISTS attempt_count;

ALTER TABLE users
    DROP COLUMN IF EXISTS login_locked_until,
    DROP COLUMN IF EXISTS login_attempt_count;
