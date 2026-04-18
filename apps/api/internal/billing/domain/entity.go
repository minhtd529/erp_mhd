// Package domain defines the Billing bounded context aggregates.
package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// InvoiceStatus represents the lifecycle of an invoice.
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "DRAFT"
	InvoiceStatusSent      InvoiceStatus = "SENT"
	InvoiceStatusConfirmed InvoiceStatus = "CONFIRMED"
	InvoiceStatusIssued    InvoiceStatus = "ISSUED"
	InvoiceStatusPaid      InvoiceStatus = "PAID"
	InvoiceStatusCancelled InvoiceStatus = "CANCELLED"
)

// InvoiceType enumerates the billing model for an invoice.
type InvoiceType string

const (
	InvoiceTypeTimeAndMaterial InvoiceType = "TIME_AND_MATERIAL"
	InvoiceTypeFixedFee        InvoiceType = "FIXED_FEE"
	InvoiceTypeRetainer        InvoiceType = "RETAINER"
	InvoiceTypeCreditNote      InvoiceType = "CREDIT_NOTE"
)

// LineItemSourceType enumerates how a line item was generated.
type LineItemSourceType string

const (
	SourceEngagementFee  LineItemSourceType = "ENGAGEMENT_FEE"
	SourceTimesheetHours LineItemSourceType = "TIMESHEET_HOURS"
	SourceDirectCost     LineItemSourceType = "DIRECT_COST"
	SourceManual         LineItemSourceType = "MANUAL"
)

// PaymentMethod enumerates how a payment was made.
type PaymentMethod string

const (
	PaymentBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentCheque       PaymentMethod = "CHEQUE"
	PaymentCash         PaymentMethod = "CASH"
	PaymentCreditCard   PaymentMethod = "CREDIT_CARD"
)

// PaymentStatus represents the lifecycle of a payment.
type PaymentStatus string

const (
	PaymentRecorded  PaymentStatus = "RECORDED"
	PaymentCleared   PaymentStatus = "CLEARED"
	PaymentDisputed  PaymentStatus = "DISPUTED"
	PaymentReversed  PaymentStatus = "REVERSED"
)

// MemoType enumerates billing memo variants.
type MemoType string

const (
	MemoCreditNote  MemoType = "CREDIT_NOTE"
	MemoAdjustment  MemoType = "ADJUSTMENT"
)

// MemoStatus represents the lifecycle of a billing memo.
type MemoStatus string

const (
	MemoStatusDraft    MemoStatus = "DRAFT"
	MemoStatusIssued   MemoStatus = "ISSUED"
	MemoStatusReversed MemoStatus = "REVERSED"
)

// Invoice is the aggregate root for the Billing context.
type Invoice struct {
	ID            uuid.UUID       `json:"id"`
	InvoiceNumber string          `json:"invoice_number"`
	ClientID      uuid.UUID       `json:"client_id"`
	EngagementID  *uuid.UUID      `json:"engagement_id"`
	InvoiceType   InvoiceType     `json:"invoice_type"`
	Status        InvoiceStatus   `json:"status"`
	IssueDate     *time.Time      `json:"issue_date"`
	DueDate       *time.Time      `json:"due_date"`
	TotalAmount   float64         `json:"total_amount"`
	TaxAmount     float64         `json:"tax_amount"`
	SnapshotData  json.RawMessage `json:"snapshot_data"`
	Notes         *string         `json:"notes"`
	IsDeleted     bool            `json:"is_deleted"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	CreatedBy     uuid.UUID       `json:"created_by"`
	UpdatedBy     uuid.UUID       `json:"updated_by"`
}

// InvoiceLineItem represents a line item on an invoice.
type InvoiceLineItem struct {
	ID           uuid.UUID          `json:"id"`
	InvoiceID    uuid.UUID          `json:"invoice_id"`
	Description  string             `json:"description"`
	Quantity     float64            `json:"quantity"`
	UnitPrice    float64            `json:"unit_price"`
	TaxAmount    float64            `json:"tax_amount"`
	TotalAmount  float64            `json:"total_amount"`
	SourceType   LineItemSourceType `json:"source_type"`
	SnapshotData json.RawMessage    `json:"snapshot_data"`
	CreatedAt    time.Time          `json:"created_at"`
}

// Payment represents a payment recorded against an invoice.
type Payment struct {
	ID              uuid.UUID     `json:"id"`
	InvoiceID       uuid.UUID     `json:"invoice_id"`
	PaymentMethod   PaymentMethod `json:"payment_method"`
	Amount          float64       `json:"amount"`
	PaymentDate     time.Time     `json:"payment_date"`
	ReferenceNumber *string       `json:"reference_number"`
	Status          PaymentStatus `json:"status"`
	Notes           *string       `json:"notes"`
	RecordedBy      uuid.UUID     `json:"recorded_by"`
	RecordedAt      time.Time     `json:"recorded_at"`
	ClearedAt       *time.Time    `json:"cleared_at"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// BillingMemo represents a credit note or adjustment memo.
type BillingMemo struct {
	ID               uuid.UUID  `json:"id"`
	RelatedInvoiceID *uuid.UUID `json:"related_invoice_id"`
	MemoType         MemoType   `json:"memo_type"`
	MemoNumber       string     `json:"memo_number"`
	Amount           float64    `json:"amount"`
	Reason           string     `json:"reason"`
	Status           MemoStatus `json:"status"`
	IsDeleted        bool       `json:"is_deleted"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CreatedBy        uuid.UUID  `json:"created_by"`
	UpdatedBy        uuid.UUID  `json:"updated_by"`
}
