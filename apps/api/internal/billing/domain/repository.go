package domain

import (
	"context"

	"github.com/google/uuid"
)

// InvoiceRepository defines persistence operations for Invoice aggregates.
type InvoiceRepository interface {
	Create(ctx context.Context, p CreateInvoiceParams) (*Invoice, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Invoice, error)
	Update(ctx context.Context, p UpdateInvoiceParams) (*Invoice, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status InvoiceStatus, updatedBy uuid.UUID) (*Invoice, error)
	UpdateSnapshot(ctx context.Context, id uuid.UUID, snapshot []byte, updatedBy uuid.UUID) (*Invoice, error)
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	List(ctx context.Context, f ListInvoicesFilter) ([]*Invoice, int64, error)
}

// LineItemRepository defines persistence for invoice line items.
type LineItemRepository interface {
	Add(ctx context.Context, p AddLineItemParams) (*InvoiceLineItem, error)
	ListByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*InvoiceLineItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*InvoiceLineItem, error)
}

// PaymentRepository defines persistence for Payment aggregates.
type PaymentRepository interface {
	Record(ctx context.Context, p RecordPaymentParams) (*Payment, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	Update(ctx context.Context, p UpdatePaymentParams) (*Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) (*Payment, error)
	Clear(ctx context.Context, id uuid.UUID) (*Payment, error)
	ListByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*Payment, error)
	ListAll(ctx context.Context, page, size int) ([]*Payment, int64, error)
	SumPaidByInvoice(ctx context.Context, invoiceID uuid.UUID) (float64, error)
}

// MemoRepository defines persistence for BillingMemo aggregates.
type MemoRepository interface {
	Create(ctx context.Context, p CreateMemoParams) (*BillingMemo, error)
	FindByID(ctx context.Context, id uuid.UUID) (*BillingMemo, error)
	List(ctx context.Context, page, size int) ([]*BillingMemo, int64, error)
}
