package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

const (
	maxLoginAttempts  = 5
	lockDuration      = 15 * time.Minute
	challengeTTL      = 5 * time.Minute
)

// LoginUseCase handles password-based authentication and token issuance.
type LoginUseCase struct {
	users    domain.UserRepository
	roles    domain.RoleRepository
	tokens   domain.RefreshTokenRepository
	twofa    domain.TwoFARepository
	jwtSvc   domain.JWTIssuer
	auditLog *audit.Logger
}

// NewLoginUseCase constructs a LoginUseCase.
func NewLoginUseCase(
	users domain.UserRepository,
	roles domain.RoleRepository,
	tokens domain.RefreshTokenRepository,
	twofa domain.TwoFARepository,
	jwtSvc domain.JWTIssuer,
	auditLog *audit.Logger,
) *LoginUseCase {
	return &LoginUseCase{
		users:    users,
		roles:    roles,
		tokens:   tokens,
		twofa:    twofa,
		jwtSvc:   jwtSvc,
		auditLog: auditLog,
	}
}

// Execute authenticates a user and returns a token pair or a 2FA challenge.
// Returns domain sentinel errors that the handler translates to HTTP status codes.
func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest, ipAddress string) (*LoginResponse, error) {
	// 1. Find user
	user, err := uc.users.FindByEmail(ctx, req.Email)
	if err != nil {
		// Mask not-found as invalid credentials to prevent user enumeration
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("login: find user: %w", err)
	}

	// 2. Check permanent account status
	switch user.Status {
	case "locked":
		return nil, domain.ErrUserLocked
	case "inactive":
		return nil, domain.ErrUserInactive
	}

	// 3. Check time-based brute-force lock (set by IncrementLoginAttempts)
	if user.LoginLockedUntil != nil && user.LoginLockedUntil.After(time.Now()) {
		return nil, domain.ErrAccountLocked
	}

	// 4. Verify password
	if err := pkgauth.CheckPassword(req.Password, user.HashedPassword); err != nil {
		if uc.twofa != nil {
			count, incErr := uc.twofa.IncrementLoginAttempts(ctx, user.ID)
			if incErr == nil && count >= maxLoginAttempts {
				_ = uc.twofa.LockAccount(ctx, user.ID, time.Now().Add(lockDuration))
			}
		}
		return nil, domain.ErrInvalidCredentials
	}

	// 5. Reset login attempt counter on successful password
	if uc.twofa != nil {
		_ = uc.twofa.ResetLoginAttempts(ctx, user.ID)
	}

	// 6. 2FA check
	if user.TwoFactorEnabled && uc.twofa != nil {
		// 6a. Check trusted device (skip 2FA if recognised)
		if req.DeviceFingerprint != "" {
			deviceHash := pkgauth.HashDeviceFingerprint(req.DeviceFingerprint)
			trusted, err := uc.twofa.FindTrustedDevice(ctx, user.ID, deviceHash)
			if err == nil && trusted != nil {
				// Trusted device — proceed straight to token issuance
				return uc.issueTokens(ctx, user, req.DeviceID, ipAddress)
			}
		}

		// 6b. Create 2FA challenge
		challengeID := uuid.New().String()
		ch := domain.TwoFactorChallenge{
			UserID:      user.ID,
			ChallengeID: challengeID,
			Method:      "totp",
			IPAddress:   ipAddress,
			ExpiresAt:   time.Now().Add(challengeTTL),
		}
		if err := uc.twofa.CreateChallenge(ctx, ch); err != nil {
			return nil, fmt.Errorf("login: create challenge: %w", err)
		}

		return &LoginResponse{
			ChallengeID:   challengeID,
			ChallengeType: "totp",
		}, domain.ErrTwoFARequired
	}

	// 7. No 2FA — issue tokens directly
	return uc.issueTokens(ctx, user, req.DeviceID, ipAddress)
}

// issueTokens creates an access + refresh token pair and emits the audit log.
func (uc *LoginUseCase) issueTokens(ctx context.Context, user *domain.UserForAuth, deviceID, ipAddress string) (*LoginResponse, error) {
	claims := pkgauth.TokenClaims{
		UserID:       user.ID,
		Email:        user.Email,
		Roles:        user.Roles,
		Permissions:  user.Permissions,
		BranchID:     user.BranchID,
		DepartmentID: user.DepartmentID,
	}

	accessToken, err := uc.jwtSvc.IssueAccessToken(claims)
	if err != nil {
		return nil, fmt.Errorf("login: issue access token: %w", err)
	}

	rawRefresh, expiresAt := uc.jwtSvc.IssueRefreshToken()
	hash := pkgauth.HashRefreshToken(rawRefresh)

	if err := uc.tokens.CreateRefreshToken(ctx, domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		DeviceID:  deviceID,
		IPAddress: ipAddress,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, fmt.Errorf("login: store refresh token: %w", err)
	}

	_ = uc.users.UpdateLastLogin(ctx, user.ID)

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     &user.ID,
			Module:     "global",
			Resource:   "users",
			ResourceID: &user.ID,
			Action:     "LOGIN",
			IPAddress:  ipAddress,
		})
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    uc.jwtSvc.AccessTokenTTLSeconds(),
	}, nil
}
