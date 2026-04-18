package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateInvoiceParams contains fields for creating a new invoice.
type CreateInvoiceParams struct {
	InvoiceNumber string
	ClientID      uuid.UUID
	EngagementID  *uuid.UUID
	InvoiceType   InvoiceType
	IssueDate     *time.Time
	DueDate       *time.Time
	TotalAmount   float64
	TaxAmount     float64
	Notes         *string
	CreatedBy     uuid.UUID
}

// UpdateInvoiceParams contains mutable fields for a DRAFT invoice.
type UpdateInvoiceParams struct {
	ID          uuid.UUID
	IssueDate   *time.Time
	DueDate     *time.Time
	TotalAmount float64
	TaxAmount   float64
	Notes       *string
	UpdatedBy   uuid.UUID
}

// AddLineItemParams contains data for adding a line item.
type AddLineItemParams struct {
	InvoiceID    uuid.UUID
	Description  string
	Quantity     float64
	UnitPrice    float64
	TaxAmount    float64
	TotalAmount  float64
	SourceType   LineItemSourceType
	SnapshotData json.RawMessage
}

// ListInvoicesFilter holds pagination + filtering options.
type ListInvoicesFilter struct {
	Page         int
	Size         int
	ClientID     *uuid.UUID
	EngagementID *uuid.UUID
	Status       InvoiceStatus
	Statuses     []InvoiceStatus // overrides Status when non-empty
	Q            string
}

// RecordPaymentParams contains data for recording a payment.
type RecordPaymentParams struct {
	InvoiceID       uuid.UUID
	PaymentMethod   PaymentMethod
	Amount          float64
	PaymentDate     time.Time
	ReferenceNumber *string
	Notes           *string
	RecordedBy      uuid.UUID
}

// UpdatePaymentParams contains mutable fields for a RECORDED payment.
type UpdatePaymentParams struct {
	ID              uuid.UUID
	PaymentMethod   PaymentMethod
	Amount          float64
	PaymentDate     time.Time
	ReferenceNumber *string
	Notes           *string
}

// CreateMemoParams contains data for creating a billing memo.
type CreateMemoParams struct {
	RelatedInvoiceID *uuid.UUID
	MemoType         MemoType
	MemoNumber       string
	Amount           float64
	Reason           string
	CreatedBy        uuid.UUID
}
