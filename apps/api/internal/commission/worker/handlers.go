// Package worker provides Asynq task handlers for commission accrual events.
package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/mdh/erp-audit/api/internal/commission/usecase"
)

// InvoiceCancelledUC is the subset of AccrualUseCase needed by the cancelled handler.
// Defined as an interface to allow test stubs.
type InvoiceCancelledUC interface {
	ClawbackOnInvoiceCancelled(ctx context.Context, invoiceID uuid.UUID) error
}

// invoiceIssuedPayload matches the payload published by billing invoice.Issue.
type invoiceIssuedPayload struct {
	InvoiceID string `json:"invoice_id"`
	ClientID  string `json:"client_id"`
}

// paymentReceivedPayload matches the payload published by billing RecordPayment.
type paymentReceivedPayload struct {
	PaymentID string `json:"payment_id"`
	InvoiceID string `json:"invoice_id"`
}

// engagementSettledPayload matches the payload for EngagementSettled events.
type engagementSettledPayload struct {
	EngagementID string `json:"engagement_id"`
	Status       string `json:"status"`
	CallerID     string `json:"caller_id"`
}

// NewInvoiceIssuedHandler returns an Asynq handler for invoice.issued events.
func NewInvoiceIssuedHandler(uc *usecase.AccrualUseCase) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p invoiceIssuedPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invoice.issued: bad payload: %w", err)
		}
		id, err := uuid.Parse(p.InvoiceID)
		if err != nil {
			return fmt.Errorf("invoice.issued: invalid invoice_id %q: %w", p.InvoiceID, err)
		}
		if err := uc.AccrueOnInvoiceIssued(ctx, id); err != nil {
			return fmt.Errorf("invoice.issued: accrue: %w", err)
		}
		return nil
	}
}

// NewPaymentReceivedHandler returns an Asynq handler for payment.received events.
func NewPaymentReceivedHandler(uc *usecase.AccrualUseCase) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p paymentReceivedPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("payment.received: bad payload: %w", err)
		}
		id, err := uuid.Parse(p.PaymentID)
		if err != nil {
			return fmt.Errorf("payment.received: invalid payment_id %q: %w", p.PaymentID, err)
		}
		if err := uc.AccrueOnPaymentReceived(ctx, id); err != nil {
			return fmt.Errorf("payment.received: accrue: %w", err)
		}
		return nil
	}
}

// NewInvoiceCancelledHandler returns an Asynq handler for invoice.cancelled events.
func NewInvoiceCancelledHandler(uc *usecase.AccrualUseCase) func(context.Context, *asynq.Task) error {
	return NewInvoiceCancelledHandlerFn(uc)
}

// NewInvoiceCancelledHandlerFn is the testable variant that accepts the InvoiceCancelledUC interface.
func NewInvoiceCancelledHandlerFn(uc InvoiceCancelledUC) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p invoiceIssuedPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("invoice.cancelled: bad payload: %w", err)
		}
		id, err := uuid.Parse(p.InvoiceID)
		if err != nil {
			return fmt.Errorf("invoice.cancelled: invalid invoice_id %q: %w", p.InvoiceID, err)
		}
		if err := uc.ClawbackOnInvoiceCancelled(ctx, id); err != nil {
			return fmt.Errorf("invoice.cancelled: clawback: %w", err)
		}
		return nil
	}
}

// NewEngagementSettledHandler returns an Asynq handler for EngagementSettled events.
func NewEngagementSettledHandler(uc *usecase.AccrualUseCase) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p engagementSettledPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("EngagementSettled: bad payload: %w", err)
		}
		id, err := uuid.Parse(p.EngagementID)
		if err != nil {
			return fmt.Errorf("EngagementSettled: invalid engagement_id %q: %w", p.EngagementID, err)
		}
		if err := uc.ReleaseHoldback(ctx, id); err != nil {
			return fmt.Errorf("EngagementSettled: release holdback: %w", err)
		}
		return nil
	}
}
