package domain

import (
	"context"

	"github.com/google/uuid"
)

// ARAgingRow represents the receivables aging breakdown for one client.
type ARAgingRow struct {
	ClientID         uuid.UUID `json:"client_id"`
	Current          float64   `json:"current"`          // not yet due
	Days1to30        float64   `json:"days_1_30"`        // 1-30 days overdue
	Days31to60       float64   `json:"days_31_60"`       // 31-60 days overdue
	Days61to90       float64   `json:"days_61_90"`       // 61-90 days overdue
	Days90Plus       float64   `json:"days_90_plus"`     // >90 days overdue
	TotalOutstanding float64   `json:"total_outstanding"`
}

// AROutstandingRow represents the outstanding balance for one client.
type AROutstandingRow struct {
	ClientID         uuid.UUID `json:"client_id"`
	InvoiceCount     int       `json:"invoice_count"`
	TotalBilled      float64   `json:"total_billed"`
	TotalPaid        float64   `json:"total_paid"`
	TotalOutstanding float64   `json:"total_outstanding"`
}

// ARRepository defines read-only queries for accounts receivable analytics.
type ARRepository interface {
	GetAging(ctx context.Context) ([]*ARAgingRow, error)
	GetOutstanding(ctx context.Context) ([]*AROutstandingRow, error)
}
