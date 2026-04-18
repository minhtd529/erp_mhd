package usecase

import "github.com/mdh/erp-audit/api/pkg/pagination"

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginResponse is returned after successful authentication
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	// Populated when 2FA is required instead of a full token
	ChallengeID   string `json:"challenge_id,omitempty"`
	ChallengeType string `json:"challenge_type,omitempty"`
}

// PaginatedResult is the shared offset pagination type.
type PaginatedResult[T any] = pagination.OffsetResult[T]

// NewPaginatedResult builds a PaginatedResult.
func NewPaginatedResult[T any](data []T, total int64, page, size int) PaginatedResult[T] {
	return pagination.NewOffsetResult(data, total, page, size)
}
