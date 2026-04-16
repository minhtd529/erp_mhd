-- ============================================================
-- TOTP secret management
-- ============================================================

-- name: SetTOTPSecret :exec
UPDATE users
SET two_factor_secret = $2, updated_at = NOW()
WHERE id = $1;

-- name: GetTOTPSecret :one
SELECT two_factor_secret
FROM users
WHERE id = $1 AND is_deleted = false;

-- name: SetTwoFactorEnabled :exec
UPDATE users
SET two_factor_enabled = $2,
    two_factor_verified_at = CASE WHEN $2 THEN NOW() ELSE NULL END,
    updated_at = NOW()
WHERE id = $1;

-- name: ClearTwoFactorSecret :exec
UPDATE users
SET two_factor_secret = NULL,
    two_factor_enabled = false,
    two_factor_verified_at = NULL,
    updated_at = NOW()
WHERE id = $1;

-- ============================================================
-- Two-factor challenges
-- ============================================================

-- name: CreateChallenge :exec
INSERT INTO two_factor_challenges (user_id, challenge_id, method, ip_address, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- name: FindChallenge :one
SELECT id, user_id, challenge_id, method, ip_address,
       attempt_count, expires_at, verified_at, invalidated_at, created_at
FROM two_factor_challenges
WHERE challenge_id = $1;

-- name: IncrementChallengeAttempt :one
UPDATE two_factor_challenges
SET attempt_count = attempt_count + 1
WHERE challenge_id = $1
RETURNING attempt_count;

-- name: InvalidateChallenge :exec
UPDATE two_factor_challenges
SET invalidated_at = NOW()
WHERE challenge_id = $1 AND invalidated_at IS NULL;

-- name: MarkChallengeVerified :exec
UPDATE two_factor_challenges
SET verified_at = NOW()
WHERE challenge_id = $1 AND verified_at IS NULL;

-- ============================================================
-- Backup codes
-- ============================================================

-- name: InsertBackupCode :exec
INSERT INTO two_factor_backup_codes (user_id, code_hash)
VALUES ($1, $2);

-- name: GetUnusedBackupCodes :many
SELECT id, user_id, code_hash, used_at, created_at
FROM two_factor_backup_codes
WHERE user_id = $1 AND used_at IS NULL;

-- name: MarkBackupCodeUsed :exec
UPDATE two_factor_backup_codes
SET used_at = NOW()
WHERE id = $1 AND used_at IS NULL;

-- name: DeleteAllBackupCodes :exec
DELETE FROM two_factor_backup_codes
WHERE user_id = $1;

-- name: CountRemainingBackupCodes :one
SELECT COUNT(*) FROM two_factor_backup_codes
WHERE user_id = $1 AND used_at IS NULL;

-- ============================================================
-- Trusted devices
-- ============================================================

-- name: InsertTrustedDevice :exec
INSERT INTO trusted_devices (user_id, device_id, device_name, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, device_id) DO UPDATE
    SET device_name = EXCLUDED.device_name,
        ip_address  = EXCLUDED.ip_address,
        user_agent  = EXCLUDED.user_agent,
        expires_at  = EXCLUDED.expires_at;

-- name: FindTrustedDevice :one
SELECT id, user_id, device_id, device_name, ip_address, user_agent, expires_at, created_at
FROM trusted_devices
WHERE user_id = $1 AND device_id = $2 AND expires_at > NOW();

-- name: CountTrustedDevices :one
SELECT COUNT(*) FROM trusted_devices
WHERE user_id = $1 AND expires_at > NOW();

-- name: RevokeOldestTrustedDevice :exec
DELETE FROM trusted_devices
WHERE id = (
    SELECT id FROM trusted_devices
    WHERE user_id = $1
    ORDER BY created_at ASC
    LIMIT 1
);

-- name: RevokeAllTrustedDevices :exec
DELETE FROM trusted_devices
WHERE user_id = $1;

-- ============================================================
-- Login brute-force protection
-- ============================================================

-- name: IncrementLoginAttempts :one
UPDATE users
SET login_attempt_count = login_attempt_count + 1,
    updated_at = NOW()
WHERE id = $1
RETURNING login_attempt_count;

-- name: ResetLoginAttempts :exec
UPDATE users
SET login_attempt_count = 0,
    login_locked_until = NULL,
    updated_at = NOW()
WHERE id = $1;

-- name: LockAccount :exec
UPDATE users
SET login_locked_until = $2,
    login_attempt_count = 0,
    updated_at = NOW()
WHERE id = $1;
