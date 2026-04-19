-- ============================================================
-- Push Devices: mobile device registration for push relay
-- ============================================================

CREATE TABLE push_devices (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token   VARCHAR(500) NOT NULL UNIQUE,
    platform       VARCHAR(20) NOT NULL CHECK (platform IN ('ios', 'android', 'web_push')),
    device_name    VARCHAR(255) NOT NULL DEFAULT '',
    app_version    VARCHAR(50)  NOT NULL DEFAULT '',
    os_version     VARCHAR(50)  NOT NULL DEFAULT '',
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    last_active_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_push_devices_user   ON push_devices (user_id, is_active);
CREATE INDEX idx_push_devices_token  ON push_devices (device_token) WHERE is_active = TRUE;

-- ============================================================
-- Extend two_factor_challenges to support push 2FA response
-- ============================================================

ALTER TABLE two_factor_challenges
    ADD COLUMN IF NOT EXISTS push_response  VARCHAR(10)  DEFAULT NULL
        CHECK (push_response IN ('approved', 'rejected')),
    ADD COLUMN IF NOT EXISTS responded_at   TIMESTAMPTZ  DEFAULT NULL;
