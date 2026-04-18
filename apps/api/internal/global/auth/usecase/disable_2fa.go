package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// Disable2FAUseCase removes TOTP configuration for a user after verifying password.
type Disable2FAUseCase struct {
	users    domain.UserRepository
	twofa    domain.TwoFARepository
	tokens   domain.RefreshTokenRepository
	auditLog *audit.Logger
}

func NewDisable2FAUseCase(
	users domain.UserRepository,
	twofa domain.TwoFARepository,
	tokens domain.RefreshTokenRepository,
	auditLog *audit.Logger,
) *Disable2FAUseCase {
	return &Disable2FAUseCase{users: users, twofa: twofa, tokens: tokens, auditLog: auditLog}
}

// Execute verifies the user's password, then clears all 2FA data and trusted devices.
func (uc *Disable2FAUseCase) Execute(ctx context.Context, userID uuid.UUID, password, ipAddress string) error {
	user, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.TwoFactorEnabled {
		return domain.ErrTwoFANotEnabled
	}

	// Verify password before allowing disable
	if err := pkgauth.CheckPassword(password, user.HashedPassword); err != nil {
		return domain.ErrInvalidCredentials
	}

	// Clear TOTP secret and disable flag
	if err := uc.twofa.ClearTwoFactorSecret(ctx, userID); err != nil {
		return fmt.Errorf("disable 2fa: clear secret: %w", err)
	}

	// Delete backup codes
	if err := uc.twofa.DeleteAllBackupCodes(ctx, userID); err != nil {
		return fmt.Errorf("disable 2fa: delete backup codes: %w", err)
	}

	// Revoke all trusted devices
	if err := uc.twofa.RevokeAllTrustedDevices(ctx, userID); err != nil {
		return fmt.Errorf("disable 2fa: revoke trusted devices: %w", err)
	}

	// Audit
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     &userID,
			Module:     "global",
			Resource:   "users",
			ResourceID: &userID,
			Action:     "DISABLE_2FA",
			IPAddress:  ipAddress,
		})
	}

	return nil
}
