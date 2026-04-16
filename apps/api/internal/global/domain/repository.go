package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines the data access contract for User
type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
}

// AuditLogRepository defines the data access contract for AuditLog (write-only)
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
}
