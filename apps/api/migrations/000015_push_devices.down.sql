ALTER TABLE two_factor_challenges
    DROP COLUMN IF EXISTS push_response,
    DROP COLUMN IF EXISTS responded_at;

DROP TABLE IF EXISTS push_devices;
