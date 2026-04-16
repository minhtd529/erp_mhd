package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserForAuth is a read-projection of the User aggregate containing only the
// fields required by auth use cases.
type UserForAuth struct {
	ID               uuid.UUID
	Email            string
	HashedPassword   string
	FullName         string
	BranchID         *uuid.UUID
	DepartmentID     *uuid.UUID
	Status           string // "active" | "inactive" | "locked"
	TwoFactorEnabled bool
	TwoFactorMethod  string // "totp" | "push"
	TwoFactorSecret  string // AES-256-GCM encrypted base32 TOTP secret

	// Brute-force protection (from migration 000003)
	LoginAttemptCount int
	LoginLockedUntil  *time.Time

	Roles       []string // role codes e.g. ["AUDIT_MANAGER"]
	Permissions []string // "module:resource:action" e.g. ["crm:client:read"]
}

// RefreshToken is a stored representation of a refresh token for revocation support.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string // SHA-256 of the raw opaque token
	DeviceID  string
	IPAddress string
	UserAgent string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

// TwoFactorChallenge is an ephemeral record created when a user with 2FA enabled
// attempts to log in. The challenge must be verified within ExpiresAt.
type TwoFactorChallenge struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	ChallengeID  string // opaque UUID sent to client
	Method       string // "totp" | "push"
	IPAddress    string
	AttemptCount int        // incremented on each wrong OTP; invalidated at MaxAttempts
	ExpiresAt    time.Time
	VerifiedAt   *time.Time
	InvalidatedAt *time.Time // set when attempt_count reaches max
	CreatedAt    time.Time
}

// BackupCode represents a one-time-use recovery code.
type BackupCode struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	CodeHash string // bcrypt cost=10
	UsedAt   *time.Time
}

// TrustedDevice is a device that has completed 2FA within the last TrustDeviceDays days.
type TrustedDevice struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	DeviceHash string // SHA-256(fingerprint) stored as device_id in DB
	DeviceName string
	IPAddress  string
	UserAgent  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

// Role is the minimal role record needed by auth operations.
type Role struct {
	ID   uuid.UUID
	Code string
	Name string
}
