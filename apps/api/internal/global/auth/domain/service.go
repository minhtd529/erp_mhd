package domain

import (
	"time"

	"github.com/mdh/erp-audit/api/pkg/auth"
)

// JWTIssuer is the interface for token operations used by auth use cases.
// The concrete implementation lives in pkg/auth.JWTService.
type JWTIssuer interface {
	IssueAccessToken(claims auth.TokenClaims) (string, error)
	IssueRefreshToken() (rawToken string, expiresAt time.Time)
	ValidateAccessToken(tokenStr string) (*auth.TokenClaims, error)
	AccessTokenTTLSeconds() int64
}
