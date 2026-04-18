package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	pkgcrypto "github.com/mdh/erp-audit/api/pkg/crypto"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// VerifySetupUseCase confirms the user can produce a valid TOTP code for the
// secret generated in Enable2FAUseCase, then marks 2FA as active.
type VerifySetupUseCase struct {
	twofa    domain.TwoFARepository
	encKey   string
	auditLog *audit.Logger
}

func NewVerifySetupUseCase(twofa domain.TwoFARepository, encKey string, auditLog *audit.Logger) *VerifySetupUseCase {
	return &VerifySetupUseCase{twofa: twofa, encKey: encKey, auditLog: auditLog}
}

// Execute validates the TOTP code and activates 2FA for the user.
func (uc *VerifySetupUseCase) Execute(ctx context.Context, userID uuid.UUID, code, ipAddress string) error {
	// Retrieve encrypted secret
	encSecret, err := uc.twofa.GetTOTPSecret(ctx, userID)
	if err != nil {
		return err
	}
	if encSecret == "" {
		return domain.ErrTwoFANotEnabled // setup was never started
	}

	// Decrypt
	secret, err := pkgcrypto.Decrypt(uc.encKey, encSecret)
	if err != nil {
		return fmt.Errorf("verify setup: decrypt: %w", err)
	}

	// Validate TOTP code
	if !pkgauth.ValidateTOTP(secret, code) {
		return domain.ErrTwoFAInvalid
	}

	// Enable 2FA
	if err := uc.twofa.SetTwoFactorEnabled(ctx, userID, true); err != nil {
		return fmt.Errorf("verify setup: enable: %w", err)
	}

	// Audit
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     &userID,
			Module:     "global",
			Resource:   "users",
			ResourceID: &userID,
			Action:     "ENABLE_2FA",
			IPAddress:  ipAddress,
		})
	}

	return nil
}
