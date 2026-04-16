package usecase

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

// PaginatedResult is the generic paginated list response
type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalPages int   `json:"total_pages"`
}

// NewPaginatedResult builds a PaginatedResult
func NewPaginatedResult[T any](data []T, total int64, page, size int) PaginatedResult[T] {
	totalPages := int(total) / size
	if int(total)%size != 0 {
		totalPages++
	}
	return PaginatedResult[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	}
}
