package usecase

import (
	"context"
	"fmt"
	"time"

	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	pkgcrypto "github.com/mdh/erp-audit/api/pkg/crypto"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// Verify2FALoginUseCase validates a TOTP code against an active challenge and
// issues tokens on success.
type Verify2FALoginUseCase struct {
	users             domain.UserRepository
	tokens            domain.RefreshTokenRepository
	twofa             domain.TwoFARepository
	jwtSvc            domain.JWTIssuer
	encKey            string
	maxAttempts       int
	trustDeviceDays   int
	maxTrustedDevices int
	auditLog          *audit.Logger
}

func NewVerify2FALoginUseCase(
	users domain.UserRepository,
	tokens domain.RefreshTokenRepository,
	twofa domain.TwoFARepository,
	jwtSvc domain.JWTIssuer,
	encKey string,
	maxAttempts, trustDeviceDays, maxTrustedDevices int,
	auditLog *audit.Logger,
) *Verify2FALoginUseCase {
	return &Verify2FALoginUseCase{
		users:             users,
		tokens:            tokens,
		twofa:             twofa,
		jwtSvc:            jwtSvc,
		encKey:            encKey,
		maxAttempts:       maxAttempts,
		trustDeviceDays:   trustDeviceDays,
		maxTrustedDevices: maxTrustedDevices,
		auditLog:          auditLog,
	}
}

// Execute validates the TOTP code and issues tokens.
func (uc *Verify2FALoginUseCase) Execute(ctx context.Context, req Verify2FALoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	// 1. Fetch challenge
	ch, err := uc.twofa.FindChallenge(ctx, req.ChallengeID)
	if err != nil {
		return nil, err // ErrChallengeNotFound
	}

	// 2. Validate challenge state
	if ch.InvalidatedAt != nil {
		return nil, domain.ErrChallengeInvalidated
	}
	if time.Now().After(ch.ExpiresAt) {
		return nil, domain.ErrChallengeExpired
	}
	if ch.VerifiedAt != nil {
		return nil, domain.ErrChallengeInvalidated // already used
	}

	// 3. Get encrypted TOTP secret
	encSecret, err := uc.twofa.GetTOTPSecret(ctx, ch.UserID)
	if err != nil {
		return nil, err
	}
	if encSecret == "" {
		return nil, domain.ErrTwoFANotEnabled
	}

	secret, err := pkgcrypto.Decrypt(uc.encKey, encSecret)
	if err != nil {
		return nil, fmt.Errorf("verify 2fa: decrypt: %w", err)
	}

	// 4. Validate TOTP code
	if !pkgauth.ValidateTOTP(secret, req.Code) {
		count, _ := uc.twofa.IncrementChallengeAttempt(ctx, req.ChallengeID)
		if count >= uc.maxAttempts {
			_ = uc.twofa.InvalidateChallenge(ctx, req.ChallengeID)
			return nil, domain.ErrTooManyAttempts
		}
		return nil, domain.ErrTwoFAInvalid
	}

	// 5. Mark challenge verified
	if err := uc.twofa.MarkChallengeVerified(ctx, req.ChallengeID); err != nil {
		return nil, fmt.Errorf("verify 2fa: mark verified: %w", err)
	}

	// 6. Optionally trust this device
	if req.TrustDevice && req.DeviceFingerprint != "" {
		deviceHash := pkgauth.HashDeviceFingerprint(req.DeviceFingerprint)
		count, _ := uc.twofa.CountTrustedDevices(ctx, ch.UserID)
		if count >= uc.maxTrustedDevices {
			_ = uc.twofa.RevokeOldestTrustedDevice(ctx, ch.UserID)
		}
		_ = uc.twofa.AddTrustedDevice(ctx, domain.TrustedDevice{
			UserID:     ch.UserID,
			DeviceHash: deviceHash,
			DeviceName: req.DeviceName,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			ExpiresAt:  time.Now().AddDate(0, 0, uc.trustDeviceDays),
		})
	}

	// 7. Load user and issue tokens
	user, err := uc.users.FindByID(ctx, ch.UserID)
	if err != nil {
		return nil, err
	}

	claims := pkgauth.TokenClaims{
		UserID:        user.ID,
		Email:         user.Email,
		Roles:         user.Roles,
		Permissions:   user.Permissions,
		BranchID:      user.BranchID,
		DepartmentID:  user.DepartmentID,
		TwoFAVerified: true,
	}

	accessToken, err := uc.jwtSvc.IssueAccessToken(claims)
	if err != nil {
		return nil, fmt.Errorf("verify 2fa: issue token: %w", err)
	}

	rawRefresh, expiresAt := uc.jwtSvc.IssueRefreshToken()
	hash := pkgauth.HashRefreshToken(rawRefresh)

	if err := uc.tokens.CreateRefreshToken(ctx, domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, fmt.Errorf("verify 2fa: store refresh token: %w", err)
	}

	_ = uc.users.UpdateLastLogin(ctx, user.ID)

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     &user.ID,
			Module:     "global",
			Resource:   "users",
			ResourceID: &user.ID,
			Action:     "LOGIN_2FA",
			IPAddress:  ipAddress,
		})
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    uc.jwtSvc.AccessTokenTTLSeconds(),
	}, nil
}
