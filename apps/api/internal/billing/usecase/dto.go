package usecase

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// PaginatedResult is a generic paginated response alias.
type PaginatedResult[T any] = pagination.OffsetResult[T]

// ── Invoice DTOs ──────────────────────────────────────────────────────────────

type InvoiceCreateRequest struct {
	InvoiceNumber string             `json:"invoice_number" binding:"required"`
	ClientID      uuid.UUID          `json:"client_id"      binding:"required"`
	EngagementID  *uuid.UUID         `json:"engagement_id"`
	InvoiceType   domain.InvoiceType `json:"invoice_type"   binding:"required"`
	IssueDate     *time.Time         `json:"issue_date"`
	DueDate       *time.Time         `json:"due_date"`
	TotalAmount   float64            `json:"total_amount"`
	TaxAmount     float64            `json:"tax_amount"`
	Notes         *string            `json:"notes"`
}

type InvoiceUpdateRequest struct {
	IssueDate   *time.Time `json:"issue_date"`
	DueDate     *time.Time `json:"due_date"`
	TotalAmount float64    `json:"total_amount"`
	TaxAmount   float64    `json:"tax_amount"`
	Notes       *string    `json:"notes"`
}

type InvoiceListRequest struct {
	Page         int                   `form:"page"`
	Size         int                   `form:"size"`
	ClientID     *uuid.UUID            `form:"client_id"`
	EngagementID *uuid.UUID            `form:"engagement_id"`
	Status       domain.InvoiceStatus  `form:"status"`
	Q            string                `form:"q"`
}

type InvoiceResponse struct {
	ID            uuid.UUID             `json:"id"`
	InvoiceNumber string                `json:"invoice_number"`
	ClientID      uuid.UUID             `json:"client_id"`
	EngagementID  *uuid.UUID            `json:"engagement_id"`
	InvoiceType   domain.InvoiceType    `json:"invoice_type"`
	Status        domain.InvoiceStatus  `json:"status"`
	IssueDate     *time.Time            `json:"issue_date"`
	DueDate       *time.Time            `json:"due_date"`
	TotalAmount   float64               `json:"total_amount"`
	TaxAmount     float64               `json:"tax_amount"`
	SnapshotData  json.RawMessage       `json:"snapshot_data"`
	Notes         *string               `json:"notes"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	CreatedBy     uuid.UUID             `json:"created_by"`
}

func toInvoiceResponse(inv *domain.Invoice) InvoiceResponse {
	return InvoiceResponse{
		ID:            inv.ID,
		InvoiceNumber: inv.InvoiceNumber,
		ClientID:      inv.ClientID,
		EngagementID:  inv.EngagementID,
		InvoiceType:   inv.InvoiceType,
		Status:        inv.Status,
		IssueDate:     inv.IssueDate,
		DueDate:       inv.DueDate,
		TotalAmount:   inv.TotalAmount,
		TaxAmount:     inv.TaxAmount,
		SnapshotData:  inv.SnapshotData,
		Notes:         inv.Notes,
		CreatedAt:     inv.CreatedAt,
		UpdatedAt:     inv.UpdatedAt,
		CreatedBy:     inv.CreatedBy,
	}
}

// ── Line Item DTOs ────────────────────────────────────────────────────────────

type LineItemAddRequest struct {
	Description  string                    `json:"description"   binding:"required"`
	Quantity     float64                   `json:"quantity"`
	UnitPrice    float64                   `json:"unit_price"`
	TaxAmount    float64                   `json:"tax_amount"`
	SourceType   domain.LineItemSourceType `json:"source_type"`
}

type LineItemResponse struct {
	ID          uuid.UUID                 `json:"id"`
	InvoiceID   uuid.UUID                 `json:"invoice_id"`
	Description string                    `json:"description"`
	Quantity    float64                   `json:"quantity"`
	UnitPrice   float64                   `json:"unit_price"`
	TaxAmount   float64                   `json:"tax_amount"`
	TotalAmount float64                   `json:"total_amount"`
	SourceType  domain.LineItemSourceType `json:"source_type"`
	CreatedAt   time.Time                 `json:"created_at"`
}

func toLineItemResponse(item *domain.InvoiceLineItem) LineItemResponse {
	return LineItemResponse{
		ID:          item.ID,
		InvoiceID:   item.InvoiceID,
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		TaxAmount:   item.TaxAmount,
		TotalAmount: item.TotalAmount,
		SourceType:  item.SourceType,
		CreatedAt:   item.CreatedAt,
	}
}

// ── Payment DTOs ──────────────────────────────────────────────────────────────

type PaymentRecordRequest struct {
	PaymentMethod   domain.PaymentMethod `json:"payment_method"   binding:"required"`
	Amount          float64              `json:"amount"           binding:"required"`
	PaymentDate     time.Time            `json:"payment_date"     binding:"required"`
	ReferenceNumber *string              `json:"reference_number"`
	Notes           *string              `json:"notes"`
}

type PaymentUpdateRequest struct {
	PaymentMethod   domain.PaymentMethod `json:"payment_method"`
	Amount          float64              `json:"amount"`
	PaymentDate     time.Time            `json:"payment_date"`
	ReferenceNumber *string              `json:"reference_number"`
	Notes           *string              `json:"notes"`
}

type PaymentResponse struct {
	ID              uuid.UUID            `json:"id"`
	InvoiceID       uuid.UUID            `json:"invoice_id"`
	PaymentMethod   domain.PaymentMethod `json:"payment_method"`
	Amount          float64              `json:"amount"`
	PaymentDate     time.Time            `json:"payment_date"`
	ReferenceNumber *string              `json:"reference_number"`
	Status          domain.PaymentStatus `json:"status"`
	Notes           *string              `json:"notes"`
	RecordedBy      uuid.UUID            `json:"recorded_by"`
	RecordedAt      time.Time            `json:"recorded_at"`
	CreatedAt       time.Time            `json:"created_at"`
}

func toPaymentResponse(p *domain.Payment) PaymentResponse {
	return PaymentResponse{
		ID:              p.ID,
		InvoiceID:       p.InvoiceID,
		PaymentMethod:   p.PaymentMethod,
		Amount:          p.Amount,
		PaymentDate:     p.PaymentDate,
		ReferenceNumber: p.ReferenceNumber,
		Status:          p.Status,
		Notes:           p.Notes,
		RecordedBy:      p.RecordedBy,
		RecordedAt:      p.RecordedAt,
		CreatedAt:       p.CreatedAt,
	}
}

// ── Memo DTOs ─────────────────────────────────────────────────────────────────

type MemoCreateRequest struct {
	RelatedInvoiceID *uuid.UUID       `json:"related_invoice_id"`
	MemoType         domain.MemoType  `json:"memo_type"  binding:"required"`
	MemoNumber       string           `json:"memo_number" binding:"required"`
	Amount           float64          `json:"amount"     binding:"required"`
	Reason           string           `json:"reason"     binding:"required"`
}

type MemoResponse struct {
	ID               uuid.UUID        `json:"id"`
	RelatedInvoiceID *uuid.UUID       `json:"related_invoice_id"`
	MemoType         domain.MemoType  `json:"memo_type"`
	MemoNumber       string           `json:"memo_number"`
	Amount           float64          `json:"amount"`
	Reason           string           `json:"reason"`
	Status           domain.MemoStatus `json:"status"`
	CreatedAt        time.Time        `json:"created_at"`
	CreatedBy        uuid.UUID        `json:"created_by"`
}

func toMemoResponse(m *domain.BillingMemo) MemoResponse {
	return MemoResponse{
		ID:               m.ID,
		RelatedInvoiceID: m.RelatedInvoiceID,
		MemoType:         m.MemoType,
		MemoNumber:       m.MemoNumber,
		Amount:           m.Amount,
		Reason:           m.Reason,
		Status:           m.Status,
		CreatedAt:        m.CreatedAt,
		CreatedBy:        m.CreatedBy,
	}
}

// ── AR DTOs ───────────────────────────────────────────────────────────────────

type ARAgingResponse struct {
	ClientID         uuid.UUID `json:"client_id"`
	Current          float64   `json:"current"`
	Days1to30        float64   `json:"days_1_30"`
	Days31to60       float64   `json:"days_31_60"`
	Days61to90       float64   `json:"days_61_90"`
	Days90Plus       float64   `json:"days_90_plus"`
	TotalOutstanding float64   `json:"total_outstanding"`
}

type AROutstandingResponse struct {
	ClientID         uuid.UUID `json:"client_id"`
	InvoiceCount     int       `json:"invoice_count"`
	TotalBilled      float64   `json:"total_billed"`
	TotalPaid        float64   `json:"total_paid"`
	TotalOutstanding float64   `json:"total_outstanding"`
}

func toARAgingResponse(r *domain.ARAgingRow) ARAgingResponse {
	return ARAgingResponse{
		ClientID:         r.ClientID,
		Current:          r.Current,
		Days1to30:        r.Days1to30,
		Days31to60:       r.Days31to60,
		Days61to90:       r.Days61to90,
		Days90Plus:       r.Days90Plus,
		TotalOutstanding: r.TotalOutstanding,
	}
}

func toAROutstandingResponse(r *domain.AROutstandingRow) AROutstandingResponse {
	return AROutstandingResponse{
		ClientID:         r.ClientID,
		InvoiceCount:     r.InvoiceCount,
		TotalBilled:      r.TotalBilled,
		TotalPaid:        r.TotalPaid,
		TotalOutstanding: r.TotalOutstanding,
	}
}
