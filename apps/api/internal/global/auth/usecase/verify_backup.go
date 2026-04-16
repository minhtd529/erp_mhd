package usecase

import (
	"context"
	"fmt"
	"time"

	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// VerifyBackupCodeUseCase allows a user to authenticate using a one-time backup code
// when they cannot access their TOTP app.
type VerifyBackupCodeUseCase struct {
	users    domain.UserRepository
	tokens   domain.RefreshTokenRepository
	twofa    domain.TwoFARepository
	jwtSvc   domain.JWTIssuer
	auditLog *audit.Logger
}

func NewVerifyBackupCodeUseCase(
	users domain.UserRepository,
	tokens domain.RefreshTokenRepository,
	twofa domain.TwoFARepository,
	jwtSvc domain.JWTIssuer,
	auditLog *audit.Logger,
) *VerifyBackupCodeUseCase {
	return &VerifyBackupCodeUseCase{
		users:    users,
		tokens:   tokens,
		twofa:    twofa,
		jwtSvc:   jwtSvc,
		auditLog: auditLog,
	}
}

// Execute validates the backup code against the challenge's user and issues tokens.
func (uc *VerifyBackupCodeUseCase) Execute(ctx context.Context, req VerifyBackupCodeRequest, ipAddress string) (*LoginResponse, error) {
	// 1. Fetch and validate the challenge
	ch, err := uc.twofa.FindChallenge(ctx, req.ChallengeID)
	if err != nil {
		return nil, err
	}
	if ch.InvalidatedAt != nil {
		return nil, domain.ErrChallengeInvalidated
	}
	if time.Now().After(ch.ExpiresAt) {
		return nil, domain.ErrChallengeExpired
	}
	if ch.VerifiedAt != nil {
		return nil, domain.ErrChallengeInvalidated
	}

	// 2. Find matching unused backup code
	codes, err := uc.twofa.GetUnusedBackupCodes(ctx, ch.UserID)
	if err != nil {
		return nil, fmt.Errorf("verify backup: get codes: %w", err)
	}

	matched := false
	for _, bc := range codes {
		if pkgauth.CheckBackupCode(req.Code, bc.CodeHash) {
			if err := uc.twofa.MarkBackupCodeUsed(ctx, bc.ID); err != nil {
				return nil, fmt.Errorf("verify backup: mark used: %w", err)
			}
			matched = true
			break
		}
	}
	if !matched {
		return nil, domain.ErrBackupCodeInvalid
	}

	// 3. Mark challenge verified
	_ = uc.twofa.MarkChallengeVerified(ctx, req.ChallengeID)

	// 4. Load user and issue tokens
	user, err := uc.users.FindByID(ctx, ch.UserID)
	if err != nil {
		return nil, err
	}

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
		return nil, fmt.Errorf("verify backup: issue token: %w", err)
	}

	rawRefresh, expiresAt := uc.jwtSvc.IssueRefreshToken()
	hash := pkgauth.HashRefreshToken(rawRefresh)

	if err := uc.tokens.CreateRefreshToken(ctx, domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		IPAddress: ipAddress,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, fmt.Errorf("verify backup: store refresh token: %w", err)
	}

	_ = uc.users.UpdateLastLogin(ctx, user.ID)

	if uc.auditLog != nil {
		_ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     &user.ID,
			Module:     "global",
			Resource:   "users",
			ResourceID: &user.ID,
			Action:     "LOGIN_BACKUP_CODE",
			IPAddress:  ipAddress,
		})
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    uc.jwtSvc.AccessTokenTTLSeconds(),
	}, nil
}
