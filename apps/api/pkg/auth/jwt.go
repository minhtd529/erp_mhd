package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// jwtClaims is the internal registered+custom claims struct used with the JWT library.
type jwtClaims struct {
	jwt.RegisteredClaims
	TokenClaims
}

// JWTService issues and validates JWT access tokens plus opaque refresh tokens.
type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewJWTService constructs a JWTService.
func NewJWTService(secret string, accessTTLMinutes, refreshTTLDays int) *JWTService {
	return &JWTService{
		secret:          []byte(secret),
		accessTokenTTL:  time.Duration(accessTTLMinutes) * time.Minute,
		refreshTokenTTL: time.Duration(refreshTTLDays) * 24 * time.Hour,
	}
}

// IssueAccessToken signs a new access token with the given claims.
func (s *JWTService) IssueAccessToken(claims TokenClaims) (string, error) {
	now := time.Now()
	c := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			ID:        uuid.New().String(),
		},
		TokenClaims: claims,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

// IssueRefreshToken generates a random opaque refresh token and its expiry.
// The caller must hash and persist it.
func (s *JWTService) IssueRefreshToken() (rawToken string, expiresAt time.Time) {
	return uuid.New().String(), time.Now().Add(s.refreshTokenTTL)
}

// ValidateAccessToken parses and validates an access token; returns its claims on success.
func (s *JWTService) ValidateAccessToken(tokenStr string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	c, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return &c.TokenClaims, nil
}

// AccessTokenTTLSeconds returns the access token TTL in seconds (for expires_in responses).
func (s *JWTService) AccessTokenTTLSeconds() int64 {
	return int64(s.accessTokenTTL.Seconds())
}
