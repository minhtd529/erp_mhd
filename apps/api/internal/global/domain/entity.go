package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is the aggregate root for authentication & identity
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Email          string     `json:"email" db:"email"`
	HashedPassword string     `json:"-" db:"hashed_password"`
	FullName       string     `json:"full_name" db:"full_name"`
	EmployeeID     *uuid.UUID `json:"employee_id" db:"employee_id"`
	BranchID       *uuid.UUID `json:"branch_id" db:"branch_id"`
	DepartmentID   *uuid.UUID `json:"department_id" db:"department_id"`
	Status         UserStatus `json:"status" db:"status"`
	LastLoginAt    *time.Time `json:"last_login_at" db:"last_login_at"`

	TwoFactorEnabled   bool        `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorMethod    TwoFAMethod `json:"two_factor_method" db:"two_factor_method"`
	TwoFactorSecret    string      `json:"-" db:"two_factor_secret"`
	TwoFactorVerifiedAt *time.Time `json:"two_factor_verified_at" db:"two_factor_verified_at"`
	BackupCodesHash    string      `json:"-" db:"backup_codes_hash"`

	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy *uuid.UUID `json:"updated_by" db:"updated_by"`
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

type TwoFAMethod string

const (
	TwoFAMethodTOTP TwoFAMethod = "totp"
	TwoFAMethodPush TwoFAMethod = "push"
)

// Role represents an RBAC role
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Level       int       `json:"level" db:"level"`
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AuditLog is an immutable record of a mutation
type AuditLog struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     *uuid.UUID `json:"user_id" db:"user_id"`
	Module     string     `json:"module" db:"module"`
	Resource   string     `json:"resource" db:"resource"`
	ResourceID *uuid.UUID `json:"resource_id" db:"resource_id"`
	Action     string     `json:"action" db:"action"`
	OldValue   []byte     `json:"old_value" db:"old_value"`
	NewValue   []byte     `json:"new_value" db:"new_value"`
	IPAddress  string     `json:"ip_address" db:"ip_address"`
	UserAgent  string     `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}
