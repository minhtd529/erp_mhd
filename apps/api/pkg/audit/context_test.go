package audit_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

func TestGetID_NoSlot_ReturnsNil(t *testing.T) {
	ctx := context.Background()
	if id := audit.GetID(ctx); id != uuid.Nil {
		t.Errorf("want uuid.Nil without slot, got %s", id)
	}
}

func TestWithAuditSlot_GetID_ReturnsNilBeforeLog(t *testing.T) {
	ctx := audit.WithAuditSlot(context.Background())
	if id := audit.GetID(ctx); id != uuid.Nil {
		t.Errorf("want uuid.Nil before any Log() call, got %s", id)
	}
}

func TestWithAuditSlot_NilLogger_GetID_ReturnsNil(t *testing.T) {
	ctx := audit.WithAuditSlot(context.Background())

	// nil Logger is a no-op — slot stays nil
	var l *audit.Logger
	_, _ = l.Log(ctx, audit.Entry{Module: "test", Action: "X"})

	if id := audit.GetID(ctx); id != uuid.Nil {
		t.Errorf("nil logger should leave slot empty, got %s", id)
	}
}
