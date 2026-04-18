package worker_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/mdh/erp-audit/api/internal/commission/worker"
)

// stubAccrualUC is a minimal stub that records the last invoice/payment ID seen.
type stubClawbackUC struct {
	calledWithInvoiceID uuid.UUID
	err                 error
}

func (s *stubClawbackUC) ClawbackOnInvoiceCancelled(_ context.Context, id uuid.UUID) error {
	s.calledWithInvoiceID = id
	return s.err
}

// newInvoiceCancelledTask builds an Asynq task with the given invoice_id payload.
func newInvoiceCancelledTask(invoiceID string) *asynq.Task {
	payload, _ := json.Marshal(map[string]string{"invoice_id": invoiceID})
	return asynq.NewTask("invoice.cancelled", payload)
}

func TestInvoiceCancelledHandler_ValidPayload(t *testing.T) {
	t.Parallel()
	invoiceID := uuid.New()
	stub := &stubClawbackUC{}

	// Build a fake AccrualUseCase wrapper using the worker factory
	// We test the handler factory separately via the real handler function.
	handler := worker.NewInvoiceCancelledHandlerFn(stub)

	task := newInvoiceCancelledTask(invoiceID.String())
	if err := handler(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.calledWithInvoiceID != invoiceID {
		t.Errorf("handler called with %v, want %v", stub.calledWithInvoiceID, invoiceID)
	}
}

func TestInvoiceCancelledHandler_BadPayload(t *testing.T) {
	t.Parallel()
	stub := &stubClawbackUC{}
	handler := worker.NewInvoiceCancelledHandlerFn(stub)

	task := asynq.NewTask("invoice.cancelled", []byte("not json"))
	if err := handler(context.Background(), task); err == nil {
		t.Fatal("expected error for bad JSON payload")
	}
}

func TestInvoiceCancelledHandler_InvalidUUID(t *testing.T) {
	t.Parallel()
	stub := &stubClawbackUC{}
	handler := worker.NewInvoiceCancelledHandlerFn(stub)

	task := newInvoiceCancelledTask("not-a-uuid")
	if err := handler(context.Background(), task); err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}
