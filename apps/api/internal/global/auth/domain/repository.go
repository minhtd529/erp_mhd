package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserRepository is the read/write interface for user data used by auth.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*UserForAuth, error)
	FindByID(ctx context.Context, id uuid.UUID) (*UserForAuth, error)
	CreateUser(ctx context.Context, p CreateUserParams) (uuid.UUID, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateUser(ctx context.Context, p UpdateUserParams) error
	SoftDeleteUser(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	ListUsers(ctx context.Context, f ListUsersFilter) ([]*UserForAuth, int64, error)
}

// CreateUserParams holds the data needed to insert a new user row.
type CreateUserParams struct {
	Email          string
	HashedPassword string
	FullName       string
	BranchID       *uuid.UUID
	DepartmentID   *uuid.UUID
	CreatedBy      *uuid.UUID
}

// UpdateUserParams holds the fields that can be updated on an existing user.
type UpdateUserParams struct {
	ID           uuid.UUID
	FullName     string
	BranchID     *uuid.UUID
	DepartmentID *uuid.UUID
	Status       string
	UpdatedBy    *uuid.UUID
}

// ListUsersFilter controls pagination and filtering for the user list.
type ListUsersFilter struct {
	Page   int
	Size   int
	Status string
	Q      string // free-text on full_name / email
}

// RoleRepository handles role lookups and user-role assignments.
type RoleRepository interface {
	FindByCode(ctx context.Context, code string) (*Role, error)
	AssignToUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// RefreshTokenRepository persists and queries refresh tokens.
type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, t RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string, at time.Time) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

// TwoFARepository manages TOTP secrets, challenges, backup codes, and trusted devices.
type TwoFARepository interface {
	// TOTP secret management
	SetTOTPSecret(ctx context.Context, userID uuid.UUID, encryptedSecret string) error
	GetTOTPSecret(ctx context.Context, userID uuid.UUID) (string, error)
	SetTwoFactorEnabled(ctx context.Context, userID uuid.UUID, enabled bool) error
	ClearTwoFactorSecret(ctx context.Context, userID uuid.UUID) error

	// Challenges
	CreateChallenge(ctx context.Context, ch TwoFactorChallenge) error
	FindChallenge(ctx context.Context, challengeID string) (*TwoFactorChallenge, error)
	IncrementChallengeAttempt(ctx context.Context, challengeID string) (int, error)
	InvalidateChallenge(ctx context.Context, challengeID string) error
	MarkChallengeVerified(ctx context.Context, challengeID string) error

	// Backup codes
	StoreBackupCodes(ctx context.Context, userID uuid.UUID, codeHashes []string) error
	GetUnusedBackupCodes(ctx context.Context, userID uuid.UUID) ([]BackupCode, error)
	MarkBackupCodeUsed(ctx context.Context, codeID uuid.UUID) error
	DeleteAllBackupCodes(ctx context.Context, userID uuid.UUID) error
	CountRemainingBackupCodes(ctx context.Context, userID uuid.UUID) (int, error)

	// Trusted devices
	AddTrustedDevice(ctx context.Context, device TrustedDevice) error
	FindTrustedDevice(ctx context.Context, userID uuid.UUID, deviceHash string) (*TrustedDevice, error)
	CountTrustedDevices(ctx context.Context, userID uuid.UUID) (int, error)
	RevokeOldestTrustedDevice(ctx context.Context, userID uuid.UUID) error
	RevokeAllTrustedDevices(ctx context.Context, userID uuid.UUID) error

	// Login brute-force protection
	IncrementLoginAttempts(ctx context.Context, userID uuid.UUID) (int, error)
	ResetLoginAttempts(ctx context.Context, userID uuid.UUID) error
	LockAccount(ctx context.Context, userID uuid.UUID, until time.Time) error
}
