package outbox_test

import (
	"context"
	"testing"

	"github.com/mdh/erp-audit/api/pkg/outbox"
)

// A nil Publisher must never panic — it is used in unit tests without a real DB.
func TestPublisher_NilSafe(t *testing.T) {
	t.Parallel()
	var p *outbox.Publisher
	if err := p.Publish(context.Background(), "timesheet", [16]byte{}, outbox.EventTimesheetSubmitted, nil); err != nil {
		t.Fatalf("expected nil error from nil publisher, got %v", err)
	}
}
