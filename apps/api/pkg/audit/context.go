package audit

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const auditSlotKey contextKey = "audit_slot"

// auditIDSlot is a mutable pointer stored in context so any call to Log()
// can write the generated audit_id back through a shared reference.
type auditIDSlot struct{ id uuid.UUID }

// WithAuditSlot returns a child context that holds a mutable slot for the
// audit entry ID. Call GetID after the handler finishes to read the value.
func WithAuditSlot(ctx context.Context) context.Context {
	return context.WithValue(ctx, auditSlotKey, &auditIDSlot{})
}

// GetID returns the audit entry ID written by the most recent Log() call on
// this context. Returns uuid.Nil if no Log() was called or no slot was set.
func GetID(ctx context.Context) uuid.UUID {
	if slot, ok := ctx.Value(auditSlotKey).(*auditIDSlot); ok {
		return slot.id
	}
	return uuid.Nil
}
