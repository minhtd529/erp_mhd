// Package domain defines the notification entity and repository interface.
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotificationNotFound = errors.New("NOTIFICATION_NOT_FOUND")

// Notification types emitted by the HRM daily reminder job.
const (
	TypeCertExpiry          = "CERT_EXPIRY"
	TypeCPEDeadline         = "CPE_DEADLINE"
	TypeProvisioningExpired = "PROVISIONING_EXPIRED"
	TypeContractExpiry      = "CONTRACT_EXPIRY"
)

// Notification is one inbox row for a user.
type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string
	Title     string
	Body      string
	Data      []byte // raw JSONB
	SourceRef string
	IsRead    bool
	ReadAt    *time.Time
	CreatedAt time.Time
}

// InsertParams carries all fields needed for a new notification row.
type InsertParams struct {
	UserID    uuid.UUID
	Type      string
	Title     string
	Body      string
	Data      []byte
	SourceRef string
}

// Repository is the data-access contract for the notifications table.
type Repository interface {
	// Insert writes a new notification row. ON CONFLICT (user_id, source_ref) DO NOTHING
	// ensures idempotence when the daily job re-runs.
	Insert(ctx context.Context, p InsertParams) error

	// ListByUserID returns paginated notifications for a user, newest first.
	ListByUserID(ctx context.Context, userID uuid.UUID, page, size int) ([]*Notification, int64, error)

	// MarkRead flips is_read to true for a notification that belongs to userID.
	// Returns ErrNotificationNotFound if the row does not exist or belongs to a different user.
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
}
