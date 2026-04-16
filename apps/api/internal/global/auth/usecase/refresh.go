package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// RefreshTokenUseCase exchanges a valid refresh token for a new access token.
type RefreshTokenUseCase struct {
	users  domain.UserRepository
	tokens domain.RefreshTokenRepository
	jwtSvc domain.JWTIssuer
}

// NewRefreshTokenUseCase constructs a RefreshTokenUseCase.
func NewRefreshTokenUseCase(
	users domain.UserRepository,
	tokens domain.RefreshTokenRepository,
	jwtSvc domain.JWTIssuer,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{users: users, tokens: tokens, jwtSvc: jwtSvc}
}

// Execute validates the refresh token and issues a new access token.
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, rawToken string) (*RefreshResponse, error) {
	hash := pkgauth.HashRefreshToken(rawToken)

	// 1. Look up the stored token
	stored, err := uc.tokens.FindByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	// 2. Check revocation and expiry
	if stored.RevokedAt != nil {
		return nil, domain.ErrTokenInvalid
	}
	if time.Now().After(stored.ExpiresAt) {
		return nil, domain.ErrTokenExpired
	}

	// 3. Load user
	user, err := uc.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, fmt.Errorf("refresh: find user: %w", err)
	}
	if user.Status != "active" {
		// Revoke and deny
		_ = uc.tokens.Revoke(ctx, hash, time.Now())
		return nil, domain.ErrUserInactive
	}

	// 4. Issue new access token
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
		return nil, fmt.Errorf("refresh: issue token: %w", err)
	}

	return &RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   uc.jwtSvc.AccessTokenTTLSeconds(),
	}, nil
}
