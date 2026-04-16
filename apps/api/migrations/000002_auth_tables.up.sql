-- ============================================================
-- Auth: Trusted Devices (devices that have passed 2FA; skip 2FA for 30 days)
-- ============================================================
CREATE TABLE trusted_devices (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id   VARCHAR(64) NOT NULL,
    device_name VARCHAR(200),
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, device_id)
);

CREATE INDEX idx_trusted_devices_user ON trusted_devices(user_id, expires_at);

-- ============================================================
-- Auth: Two-Factor Backup Codes (one-time use; hashed)
-- ============================================================
CREATE TABLE two_factor_backup_codes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash  VARCHAR(255) NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_2fa_backup_codes_user
    ON two_factor_backup_codes(user_id)
    WHERE used_at IS NULL;

-- ============================================================
-- Auth: Two-Factor Challenges (ephemeral; TTL ~5 min)
-- TODO: Phase 1.2 — also mirror in Redis for fast lookup
-- ============================================================
CREATE TABLE two_factor_challenges (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_id VARCHAR(64) NOT NULL UNIQUE,
    method       VARCHAR(20) NOT NULL CHECK (method IN ('totp', 'push')),
    ip_address   VARCHAR(45),
    expires_at   TIMESTAMPTZ NOT NULL,
    verified_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_2fa_challenges_id
    ON two_factor_challenges(challenge_id)
    WHERE verified_at IS NULL;

-- ============================================================
-- Auth: Refresh Tokens (stored for explicit revocation)
-- ============================================================
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    device_id   VARCHAR(64),
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user
    ON refresh_tokens(user_id)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_refresh_tokens_hash
    ON refresh_tokens(token_hash)
    WHERE revoked_at IS NULL;
