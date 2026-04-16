package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Entry holds the data for one audit log record
type Entry struct {
	UserID     *uuid.UUID
	Module     string
	Resource   string
	ResourceID *uuid.UUID
	Action     string
	OldValue   any
	NewValue   any
	IPAddress  string
	UserAgent  string
}

// Logger writes immutable audit entries to the database
type Logger struct {
	pool *pgxpool.Pool
}

// New creates a new audit Logger
func New(pool *pgxpool.Pool) *Logger {
	return &Logger{pool: pool}
}

// Log persists an audit entry (fire-and-forget by design — errors are logged but not propagated).
// A nil Logger is a no-op, allowing use cases to be tested without a real DB.
func (l *Logger) Log(ctx context.Context, e Entry) error {
	if l == nil {
		return nil
	}
	oldVal, _ := json.Marshal(e.OldValue)
	newVal, _ := json.Marshal(e.NewValue)

	_, err := l.pool.Exec(ctx, `
		INSERT INTO audit_logs
		    (user_id, module, resource, resource_id, action, old_value, new_value, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, e.UserID, e.Module, e.Resource, e.ResourceID, e.Action,
		oldVal, newVal, e.IPAddress, e.UserAgent)

	return err
}
