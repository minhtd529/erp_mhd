package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/outbox"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// PaymentUseCase handles payment recording and lifecycle.
type PaymentUseCase struct {
	payRepo     domain.PaymentRepository
	invoiceRepo domain.InvoiceRepository
	auditLog    *audit.Logger
	publisher   *outbox.Publisher
}

// NewPaymentUseCase constructs a PaymentUseCase.
func NewPaymentUseCase(
	payRepo domain.PaymentRepository,
	invoiceRepo domain.InvoiceRepository,
	auditLog *audit.Logger,
	publisher *outbox.Publisher,
) *PaymentUseCase {
	return &PaymentUseCase{
		payRepo:     payRepo,
		invoiceRepo: invoiceRepo,
		auditLog:    auditLog,
		publisher:   publisher,
	}
}

func (uc *PaymentUseCase) Record(ctx context.Context, invoiceID uuid.UUID, req PaymentRecordRequest, callerID uuid.UUID, ip string) (*PaymentResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	// Payment only allowed on ISSUED or PAID invoices
	if inv.Status != domain.InvoiceStatusIssued && inv.Status != domain.InvoiceStatusPaid {
		return nil, domain.ErrInvalidStateTransition
	}

	// Guard: payment cannot exceed remaining balance
	paid, err := uc.payRepo.SumPaidByInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if paid+req.Amount > inv.TotalAmount {
		return nil, domain.ErrPaymentExceedsBalance
	}

	pay, err := uc.payRepo.Record(ctx, domain.RecordPaymentParams{
		InvoiceID:       invoiceID,
		PaymentMethod:   req.PaymentMethod,
		Amount:          req.Amount,
		PaymentDate:     req.PaymentDate,
		ReferenceNumber: req.ReferenceNumber,
		Notes:           req.Notes,
		RecordedBy:      callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "payments",
		ResourceID: &pay.ID, Action: "CREATE", IPAddress: ip,
	})

	// If full amount now paid, transition invoice to PAID
	if paid+req.Amount >= inv.TotalAmount {
		_, _ = uc.invoiceRepo.UpdateStatus(ctx, invoiceID, domain.InvoiceStatusPaid, callerID)
	}

	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, "payment", pay.ID, outbox.EventType("payment.received"), map[string]string{
			"payment_id": pay.ID.String(), "invoice_id": invoiceID.String(),
		})
	}

	resp := toPaymentResponse(pay)
	return &resp, nil
}

func (uc *PaymentUseCase) Update(ctx context.Context, id uuid.UUID, req PaymentUpdateRequest, callerID uuid.UUID, ip string) (*PaymentResponse, error) {
	pay, err := uc.payRepo.Update(ctx, domain.UpdatePaymentParams{
		ID:              id,
		PaymentMethod:   req.PaymentMethod,
		Amount:          req.Amount,
		PaymentDate:     req.PaymentDate,
		ReferenceNumber: req.ReferenceNumber,
		Notes:           req.Notes,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "payments",
		ResourceID: &id, Action: "UPDATE", IPAddress: ip,
	})

	resp := toPaymentResponse(pay)
	return &resp, nil
}

// Reverse marks a payment as REVERSED.
func (uc *PaymentUseCase) Reverse(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) error {
	pay, err := uc.payRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if pay.Status != domain.PaymentRecorded && pay.Status != domain.PaymentCleared {
		return domain.ErrPaymentNotRecorded
	}
	if _, err := uc.payRepo.UpdateStatus(ctx, id, domain.PaymentReversed); err != nil {
		return err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "payments",
		ResourceID: &id, Action: "DELETE", IPAddress: ip,
	})
	return nil
}

// ClearPayment confirms a RECORDED payment has cleared the bank — transitions to CLEARED.
func (uc *PaymentUseCase) ClearPayment(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*PaymentResponse, error) {
	pay, err := uc.payRepo.Clear(ctx, id)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "payments",
		ResourceID: &id, Action: "UPDATE", IPAddress: ip,
		NewValue: map[string]string{"status": "CLEARED"},
	})

	resp := toPaymentResponse(pay)
	return &resp, nil
}

// DisputePayment marks a CLEARED payment as DISPUTED.
func (uc *PaymentUseCase) DisputePayment(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*PaymentResponse, error) {
	pay, err := uc.payRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pay.Status != domain.PaymentCleared {
		return nil, domain.ErrPaymentNotCleared
	}

	updated, err := uc.payRepo.UpdateStatus(ctx, id, domain.PaymentDisputed)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "payments",
		ResourceID: &id, Action: "UPDATE", IPAddress: ip,
		NewValue: map[string]string{"status": "DISPUTED"},
	})

	resp := toPaymentResponse(updated)
	return &resp, nil
}

func (uc *PaymentUseCase) ListAll(ctx context.Context, page, size int) (PaginatedResult[PaymentResponse], error) {
	payments, total, err := uc.payRepo.ListAll(ctx, page, size)
	if err != nil {
		return PaginatedResult[PaymentResponse]{}, err
	}
	data := make([]PaymentResponse, len(payments))
	for i, p := range payments {
		data[i] = toPaymentResponse(p)
	}
	return pagination.NewOffsetResult(data, total, page, size), nil
}

func (uc *PaymentUseCase) ListByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]PaymentResponse, error) {
	payments, err := uc.payRepo.ListByInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	data := make([]PaymentResponse, len(payments))
	for i, p := range payments {
		data[i] = toPaymentResponse(p)
	}
	return data, nil
}
