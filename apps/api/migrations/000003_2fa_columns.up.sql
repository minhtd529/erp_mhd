-- ============================================================
-- Phase 1.2: 2FA — add brute-force columns to users,
-- and rate-limit columns to two_factor_challenges
-- ============================================================

ALTER TABLE users
    ADD COLUMN login_attempt_count INT NOT NULL DEFAULT 0,
    ADD COLUMN login_locked_until  TIMESTAMPTZ;

ALTER TABLE two_factor_challenges
    ADD COLUMN attempt_count   INT NOT NULL DEFAULT 0,
    ADD COLUMN invalidated_at  TIMESTAMPTZ;
