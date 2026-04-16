package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// LogoutUseCase revokes the provided refresh token.
// Access tokens expire naturally — no server-side blacklist is needed at this scale.
type LogoutUseCase struct {
	tokens domain.RefreshTokenRepository
}

// NewLogoutUseCase constructs a LogoutUseCase.
func NewLogoutUseCase(tokens domain.RefreshTokenRepository) *LogoutUseCase {
	return &LogoutUseCase{tokens: tokens}
}

// Execute revokes the refresh token identified by rawToken.
func (uc *LogoutUseCase) Execute(ctx context.Context, rawToken string) error {
	hash := pkgauth.HashRefreshToken(rawToken)
	return uc.tokens.Revoke(ctx, hash, time.Now())
}

// LogoutAllUseCase revokes all refresh tokens for a given user (e.g. "log out everywhere").
type LogoutAllUseCase struct {
	tokens domain.RefreshTokenRepository
}

// NewLogoutAllUseCase constructs a LogoutAllUseCase.
func NewLogoutAllUseCase(tokens domain.RefreshTokenRepository) *LogoutAllUseCase {
	return &LogoutAllUseCase{tokens: tokens}
}

// Execute revokes all active refresh tokens for the given user.
func (uc *LogoutAllUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	return uc.tokens.RevokeAllForUser(ctx, userID)
}
