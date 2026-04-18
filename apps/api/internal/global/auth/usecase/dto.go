package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// ─── Auth DTOs ───────────────────────────────────────────────────────────────

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Email             string `json:"email"              binding:"required,email"`
	Password          string `json:"password"           binding:"required,min=8"`
	DeviceID          string `json:"device_id"`          // stable device identifier for refresh token tracking
	DeviceFingerprint string `json:"device_fingerprint"` // raw fingerprint for trusted device check (hashed before storage)
}

// LoginResponse is returned on successful login.
// If the user has 2FA enabled the access/refresh tokens are omitted and
// challenge fields are populated instead.
type LoginResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"` // seconds

	// 2FA challenge (populated instead of tokens when 2FA is required)
	ChallengeID   string `json:"challenge_id,omitempty"`
	ChallengeType string `json:"challenge_type,omitempty"` // "totp"
}

// RefreshRequest is the body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResponse is returned with a new access token.
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// LogoutRequest is the body for POST /auth/logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ─── User Management DTOs ────────────────────────────────────────────────────

// UserCreateRequest is the body for POST /users.
type UserCreateRequest struct {
	Email        string     `json:"email"         binding:"required,email"`
	Password     string     `json:"password"      binding:"required,min=12"`
	FullName     string     `json:"full_name"     binding:"required"`
	RoleCode     string     `json:"role_code"     binding:"required"`
	BranchID     *uuid.UUID `json:"branch_id"`
	DepartmentID *uuid.UUID `json:"department_id"`
}

// UserDetailResponse is returned for user GET/POST responses.
type UserDetailResponse struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	FullName     string     `json:"full_name"`
	Status       string     `json:"status"`
	Roles        []string   `json:"roles"`
	BranchID     *uuid.UUID `json:"branch_id"`
	DepartmentID *uuid.UUID `json:"department_id"`
}

// AssignRoleRequest is the body for POST /users/{id}/roles.
type AssignRoleRequest struct {
	RoleCode string `json:"role_code" binding:"required"`
}

// PaginatedResult is the shared offset pagination type.
type PaginatedResult[T any] = pagination.OffsetResult[T]

// ─── 2FA DTOs ────────────────────────────────────────────────────────────────

// Enable2FAResponse is returned by POST /auth/2fa/setup.
type Enable2FAResponse struct {
	Secret          string `json:"secret"`           // base32 secret for manual entry
	QRCodePNG       []byte `json:"qr_code_png"`      // raw PNG bytes (base64-encoded by JSON)
	BackupCodes     []string `json:"backup_codes"`   // shown once — user must save these
	RemainingCodes  int    `json:"remaining_codes"`
}

// Verify2FASetupRequest is the body for POST /auth/2fa/confirm.
type Verify2FASetupRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// TwoFAStatusResponse is returned by GET /auth/2fa/status.
type TwoFAStatusResponse struct {
	Enabled        bool      `json:"enabled"`
	Method         string    `json:"method,omitempty"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
	BackupCodes    int       `json:"backup_codes_remaining"`
	TrustedDevices int       `json:"trusted_devices"`
}

// Verify2FALoginRequest is the body for POST /auth/2fa/verify.
type Verify2FALoginRequest struct {
	ChallengeID       string `json:"challenge_id"        binding:"required"`
	Code              string `json:"code"                binding:"required,len=6"`
	TrustDevice       bool   `json:"trust_device"`
	DeviceName        string `json:"device_name"`
	DeviceFingerprint string `json:"device_fingerprint"`
}

// VerifyBackupCodeRequest is the body for POST /auth/2fa/backup.
type VerifyBackupCodeRequest struct {
	ChallengeID string `json:"challenge_id" binding:"required"`
	Code        string `json:"code"         binding:"required"`
}

// Disable2FARequest is the body for DELETE /auth/2fa.
type Disable2FARequest struct {
	Password string `json:"password" binding:"required"`
}

// RegenBackupCodesRequest is the body for POST /auth/2fa/backup-codes/regenerate.
type RegenBackupCodesRequest struct {
	Password string `json:"password" binding:"required"`
}

// RegenBackupCodesResponse is returned when backup codes are regenerated.
type RegenBackupCodesResponse struct {
	BackupCodes    []string `json:"backup_codes"`
	RemainingCodes int      `json:"remaining_codes"`
}

// TrustedDeviceResponse represents a trusted device in list responses.
type TrustedDeviceResponse struct {
	ID         string    `json:"id"`
	DeviceName string    `json:"device_name"`
	IPAddress  string    `json:"ip_address"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}
