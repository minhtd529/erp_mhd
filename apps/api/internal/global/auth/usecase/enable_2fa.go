package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	pkgcrypto "github.com/mdh/erp-audit/api/pkg/crypto"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
)

const (
	backupCodeCount = 10
	qrCodeSize      = 256
)

// Enable2FAUseCase generates a new TOTP key, encrypts the secret, and stores
// backup codes.  The user must confirm with a valid code (VerifySetupUseCase)
// to activate 2FA.
type Enable2FAUseCase struct {
	users  domain.UserRepository
	twofa  domain.TwoFARepository
	encKey string
	issuer string
}

func NewEnable2FAUseCase(users domain.UserRepository, twofa domain.TwoFARepository, encKey, issuer string) *Enable2FAUseCase {
	return &Enable2FAUseCase{users: users, twofa: twofa, encKey: encKey, issuer: issuer}
}

// Execute generates a TOTP key for the user and pre-stores backup codes.
// Returns the plaintext secret and QR code PNG for display; the caller must
// call VerifySetupUseCase with a valid TOTP code to enable 2FA.
func (uc *Enable2FAUseCase) Execute(ctx context.Context, userID uuid.UUID) (*Enable2FAResponse, error) {
	user, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.TwoFactorEnabled {
		return nil, domain.ErrTwoFAAlreadyEnabled
	}

	// Generate TOTP key
	key, err := pkgauth.GenerateTOTPKey(uc.issuer, user.Email)
	if err != nil {
		return nil, fmt.Errorf("enable 2fa: %w", err)
	}

	// Encrypt secret before storage
	encrypted, err := pkgcrypto.Encrypt(uc.encKey, key.Secret)
	if err != nil {
		return nil, fmt.Errorf("enable 2fa: encrypt secret: %w", err)
	}

	// Store encrypted secret (not yet enabled — user must verify first)
	if err := uc.twofa.SetTOTPSecret(ctx, userID, encrypted); err != nil {
		return nil, fmt.Errorf("enable 2fa: store secret: %w", err)
	}

	// Generate and store backup codes
	rawCodes, err := pkgauth.GenerateBackupCodes(backupCodeCount)
	if err != nil {
		return nil, fmt.Errorf("enable 2fa: backup codes: %w", err)
	}

	hashes := make([]string, len(rawCodes))
	for i, c := range rawCodes {
		h, err := pkgauth.HashBackupCode(c)
		if err != nil {
			return nil, fmt.Errorf("enable 2fa: hash backup code: %w", err)
		}
		hashes[i] = h
	}

	// Delete any old backup codes then store new ones
	_ = uc.twofa.DeleteAllBackupCodes(ctx, userID)
	if err := uc.twofa.StoreBackupCodes(ctx, userID, hashes); err != nil {
		return nil, fmt.Errorf("enable 2fa: store backup codes: %w", err)
	}

	// Render QR code
	png, err := pkgauth.QRCodePNG(key.URL, qrCodeSize)
	if err != nil {
		return nil, fmt.Errorf("enable 2fa: qr code: %w", err)
	}

	return &Enable2FAResponse{
		Secret:         key.Secret,
		QRCodePNG:      png,
		BackupCodes:    rawCodes,
		RemainingCodes: backupCodeCount,
	}, nil
}
