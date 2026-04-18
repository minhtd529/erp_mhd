package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
)

// ─── TwoFARepository implementation on *Repo ─────────────────────────────────

// SetTOTPSecret stores the AES-256-GCM encrypted TOTP secret for a user.
func (r *Repo) SetTOTPSecret(ctx context.Context, userID uuid.UUID, encryptedSecret string) error {
	const q = `UPDATE users SET two_factor_secret = $2, updated_at = NOW() WHERE id = $1`
	if _, err := r.pool.Exec(ctx, q, userID, encryptedSecret); err != nil {
		return fmt.Errorf("SetTOTPSecret: %w", err)
	}
	return nil
}

// GetTOTPSecret retrieves the encrypted TOTP secret for a user.
func (r *Repo) GetTOTPSecret(ctx context.Context, userID uuid.UUID) (string, error) {
	const q = `SELECT two_factor_secret FROM users WHERE id = $1 AND is_deleted = false`
	var secret *string
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&secret); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrUserNotFound
		}
		return "", fmt.Errorf("GetTOTPSecret: %w", err)
	}
	if secret == nil {
		return "", nil
	}
	return *secret, nil
}

// SetTwoFactorEnabled enables or disables 2FA for a user.
func (r *Repo) SetTwoFactorEnabled(ctx context.Context, userID uuid.UUID, enabled bool) error {
	const q = `
		UPDATE users
		SET two_factor_enabled = $2,
		    two_factor_verified_at = CASE WHEN $2 THEN NOW() ELSE NULL END,
		    updated_at = NOW()
		WHERE id = $1`
	if _, err := r.pool.Exec(ctx, q, userID, enabled); err != nil {
		return fmt.Errorf("SetTwoFactorEnabled: %w", err)
	}
	return nil
}

// ClearTwoFactorSecret removes the TOTP secret and disables 2FA.
func (r *Repo) ClearTwoFactorSecret(ctx context.Context, userID uuid.UUID) error {
	const q = `
		UPDATE users
		SET two_factor_secret = NULL,
		    two_factor_enabled = false,
		    two_factor_verified_at = NULL,
		    updated_at = NOW()
		WHERE id = $1`
	if _, err := r.pool.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("ClearTwoFactorSecret: %w", err)
	}
	return nil
}

// ─── Challenges ──────────────────────────────────────────────────────────────

// CreateChallenge inserts a new 2FA challenge.
func (r *Repo) CreateChallenge(ctx context.Context, ch domain.TwoFactorChallenge) error {
	const q = `
		INSERT INTO two_factor_challenges (user_id, challenge_id, method, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)`
	if _, err := r.pool.Exec(ctx, q,
		ch.UserID, ch.ChallengeID, ch.Method, ch.IPAddress, ch.ExpiresAt,
	); err != nil {
		return fmt.Errorf("CreateChallenge: %w", err)
	}
	return nil
}

// FindChallenge retrieves a challenge by its challenge_id.
func (r *Repo) FindChallenge(ctx context.Context, challengeID string) (*domain.TwoFactorChallenge, error) {
	const q = `
		SELECT id, user_id, challenge_id, method, ip_address,
		       attempt_count, expires_at, verified_at, invalidated_at, created_at
		FROM two_factor_challenges
		WHERE challenge_id = $1`

	var ch domain.TwoFactorChallenge
	err := r.pool.QueryRow(ctx, q, challengeID).Scan(
		&ch.ID, &ch.UserID, &ch.ChallengeID, &ch.Method, &ch.IPAddress,
		&ch.AttemptCount, &ch.ExpiresAt, &ch.VerifiedAt, &ch.InvalidatedAt, &ch.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrChallengeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("FindChallenge: %w", err)
	}
	return &ch, nil
}

// IncrementChallengeAttempt adds 1 to the attempt_count and returns the new count.
func (r *Repo) IncrementChallengeAttempt(ctx context.Context, challengeID string) (int, error) {
	const q = `
		UPDATE two_factor_challenges
		SET attempt_count = attempt_count + 1
		WHERE challenge_id = $1
		RETURNING attempt_count`

	var count int
	if err := r.pool.QueryRow(ctx, q, challengeID).Scan(&count); err != nil {
		return 0, fmt.Errorf("IncrementChallengeAttempt: %w", err)
	}
	return count, nil
}

// InvalidateChallenge marks a challenge as invalidated (rate-limit exceeded).
func (r *Repo) InvalidateChallenge(ctx context.Context, challengeID string) error {
	const q = `
		UPDATE two_factor_challenges
		SET invalidated_at = NOW()
		WHERE challenge_id = $1 AND invalidated_at IS NULL`
	if _, err := r.pool.Exec(ctx, q, challengeID); err != nil {
		return fmt.Errorf("InvalidateChallenge: %w", err)
	}
	return nil
}

// MarkChallengeVerified marks a challenge as successfully verified.
func (r *Repo) MarkChallengeVerified(ctx context.Context, challengeID string) error {
	const q = `
		UPDATE two_factor_challenges
		SET verified_at = NOW()
		WHERE challenge_id = $1 AND verified_at IS NULL`
	if _, err := r.pool.Exec(ctx, q, challengeID); err != nil {
		return fmt.Errorf("MarkChallengeVerified: %w", err)
	}
	return nil
}

// RespondToPushChallenge records the mobile app's approve/reject decision.
func (r *Repo) RespondToPushChallenge(ctx context.Context, challengeID string, approved bool) error {
	response := "rejected"
	if approved {
		response = "approved"
	}
	const q = `
		UPDATE two_factor_challenges
		SET push_response = $2, responded_at = NOW()
		WHERE challenge_id = $1 AND method = 'push' AND push_response IS NULL AND expires_at > NOW()`
	tag, err := r.pool.Exec(ctx, q, challengeID, response)
	if err != nil {
		return fmt.Errorf("RespondToPushChallenge: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrChallengeNotFound
	}
	return nil
}

// FindPushChallenge retrieves a push challenge with its response fields.
func (r *Repo) FindPushChallenge(ctx context.Context, challengeID string) (*domain.TwoFactorChallenge, error) {
	const q = `
		SELECT id, user_id, challenge_id, method, ip_address,
		       attempt_count, expires_at, verified_at, invalidated_at,
		       push_response, responded_at, created_at
		FROM two_factor_challenges
		WHERE challenge_id = $1 AND method = 'push'`
	var ch domain.TwoFactorChallenge
	if err := r.pool.QueryRow(ctx, q, challengeID).Scan(
		&ch.ID, &ch.UserID, &ch.ChallengeID, &ch.Method, &ch.IPAddress,
		&ch.AttemptCount, &ch.ExpiresAt, &ch.VerifiedAt, &ch.InvalidatedAt,
		&ch.PushResponse, &ch.RespondedAt, &ch.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrChallengeNotFound
		}
		return nil, fmt.Errorf("FindPushChallenge: %w", err)
	}
	return &ch, nil
}

// ─── Backup codes ─────────────────────────────────────────────────────────────

// StoreBackupCodes inserts a batch of bcrypt-hashed backup codes for a user.
func (r *Repo) StoreBackupCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error {
	const q = `INSERT INTO two_factor_backup_codes (user_id, code_hash) VALUES ($1, $2)`
	for _, h := range codeHashes {
		if _, err := r.pool.Exec(ctx, q, userID, h); err != nil {
			return fmt.Errorf("StoreBackupCodes: %w", err)
		}
	}
	return nil
}

// GetUnusedBackupCodes returns all unused backup codes for a user.
func (r *Repo) GetUnusedBackupCodes(ctx context.Context, userID uuid.UUID) ([]domain.BackupCode, error) {
	const q = `
		SELECT id, user_id, code_hash, used_at
		FROM two_factor_backup_codes
		WHERE user_id = $1 AND used_at IS NULL`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUnusedBackupCodes: %w", err)
	}
	defer rows.Close()

	var codes []domain.BackupCode
	for rows.Next() {
		var c domain.BackupCode
		if err := rows.Scan(&c.ID, &c.UserID, &c.CodeHash, &c.UsedAt); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

// MarkBackupCodeUsed sets used_at=NOW() on a specific backup code.
func (r *Repo) MarkBackupCodeUsed(ctx context.Context, codeID uuid.UUID) error {
	const q = `UPDATE two_factor_backup_codes SET used_at = NOW() WHERE id = $1 AND used_at IS NULL`
	if _, err := r.pool.Exec(ctx, q, codeID); err != nil {
		return fmt.Errorf("MarkBackupCodeUsed: %w", err)
	}
	return nil
}

// DeleteAllBackupCodes removes all backup codes for a user (used on disable or regenerate).
func (r *Repo) DeleteAllBackupCodes(ctx context.Context, userID uuid.UUID) error {
	const q = `DELETE FROM two_factor_backup_codes WHERE user_id = $1`
	if _, err := r.pool.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("DeleteAllBackupCodes: %w", err)
	}
	return nil
}

// CountRemainingBackupCodes returns the number of unused backup codes for a user.
func (r *Repo) CountRemainingBackupCodes(ctx context.Context, userID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM two_factor_backup_codes WHERE user_id = $1 AND used_at IS NULL`
	var n int
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&n); err != nil {
		return 0, fmt.Errorf("CountRemainingBackupCodes: %w", err)
	}
	return n, nil
}

// ─── Trusted devices ──────────────────────────────────────────────────────────

// AddTrustedDevice upserts a trusted device record (keyed on user_id + device_id hash).
func (r *Repo) AddTrustedDevice(ctx context.Context, d domain.TrustedDevice) error {
	const q = `
		INSERT INTO trusted_devices (user_id, device_id, device_name, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, device_id) DO UPDATE
		    SET device_name = EXCLUDED.device_name,
		        ip_address  = EXCLUDED.ip_address,
		        user_agent  = EXCLUDED.user_agent,
		        expires_at  = EXCLUDED.expires_at`
	if _, err := r.pool.Exec(ctx, q,
		d.UserID, d.DeviceHash, d.DeviceName, d.IPAddress, d.UserAgent, d.ExpiresAt,
	); err != nil {
		return fmt.Errorf("AddTrustedDevice: %w", err)
	}
	return nil
}

// FindTrustedDevice looks up a non-expired trusted device by user_id + device hash.
func (r *Repo) FindTrustedDevice(ctx context.Context, userID uuid.UUID, deviceHash string) (*domain.TrustedDevice, error) {
	const q = `
		SELECT id, user_id, device_id, device_name, ip_address, user_agent, expires_at, created_at
		FROM trusted_devices
		WHERE user_id = $1 AND device_id = $2 AND expires_at > NOW()`

	var d domain.TrustedDevice
	err := r.pool.QueryRow(ctx, q, userID, deviceHash).Scan(
		&d.ID, &d.UserID, &d.DeviceHash, &d.DeviceName,
		&d.IPAddress, &d.UserAgent, &d.ExpiresAt, &d.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // not found is not an error — caller checks nil
	}
	if err != nil {
		return nil, fmt.Errorf("FindTrustedDevice: %w", err)
	}
	return &d, nil
}

// CountTrustedDevices returns the number of non-expired trusted devices for a user.
func (r *Repo) CountTrustedDevices(ctx context.Context, userID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM trusted_devices WHERE user_id = $1 AND expires_at > NOW()`
	var n int
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&n); err != nil {
		return 0, fmt.Errorf("CountTrustedDevices: %w", err)
	}
	return n, nil
}

// RevokeOldestTrustedDevice deletes the oldest trusted device for a user (to cap at max).
func (r *Repo) RevokeOldestTrustedDevice(ctx context.Context, userID uuid.UUID) error {
	const q = `
		DELETE FROM trusted_devices
		WHERE id = (
		    SELECT id FROM trusted_devices
		    WHERE user_id = $1
		    ORDER BY created_at ASC
		    LIMIT 1
		)`
	if _, err := r.pool.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("RevokeOldestTrustedDevice: %w", err)
	}
	return nil
}

// RevokeAllTrustedDevices removes all trusted devices for a user (e.g. on 2FA disable).
func (r *Repo) RevokeAllTrustedDevices(ctx context.Context, userID uuid.UUID) error {
	const q = `DELETE FROM trusted_devices WHERE user_id = $1`
	if _, err := r.pool.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("RevokeAllTrustedDevices: %w", err)
	}
	return nil
}

// ─── Login brute-force protection ────────────────────────────────────────────

// IncrementLoginAttempts increments the login_attempt_count and returns the new value.
func (r *Repo) IncrementLoginAttempts(ctx context.Context, userID uuid.UUID) (int, error) {
	const q = `
		UPDATE users
		SET login_attempt_count = login_attempt_count + 1, updated_at = NOW()
		WHERE id = $1
		RETURNING login_attempt_count`
	var count int
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("IncrementLoginAttempts: %w", err)
	}
	return count, nil
}

// ResetLoginAttempts clears login_attempt_count and login_locked_until.
func (r *Repo) ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	const q = `
		UPDATE users
		SET login_attempt_count = 0, login_locked_until = NULL, updated_at = NOW()
		WHERE id = $1`
	if _, err := r.pool.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("ResetLoginAttempts: %w", err)
	}
	return nil
}

// LockAccount sets login_locked_until to the given time and resets attempt count.
func (r *Repo) LockAccount(ctx context.Context, userID uuid.UUID, until time.Time) error {
	const q = `
		UPDATE users
		SET login_locked_until = $2, login_attempt_count = 0, updated_at = NOW()
		WHERE id = $1`
	if _, err := r.pool.Exec(ctx, q, userID, until); err != nil {
		return fmt.Errorf("LockAccount: %w", err)
	}
	return nil
}
