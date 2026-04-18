package domain

import (
	"context"
	"time"
)

// BillingPeriodSummary aggregates invoice and payment data for a date range.
type BillingPeriodSummary struct {
	PeriodStart      time.Time     `json:"period_start"`
	PeriodEnd        time.Time     `json:"period_end"`
	TotalInvoiced    float64       `json:"total_invoiced"`
	TotalPaid        float64       `json:"total_paid"`
	TotalOutstanding float64       `json:"total_outstanding"`
	InvoiceCount     int           `json:"invoice_count"`
	PaidCount        int           `json:"paid_count"`
	OverdueCount     int           `json:"overdue_count"`
	ByStatus         []StatusCount `json:"by_status"`
}

// StatusCount holds the invoice count and amount for one status.
type StatusCount struct {
	Status InvoiceStatus `json:"status"`
	Count  int           `json:"count"`
	Amount float64       `json:"amount"`
}

// PaymentSummary aggregates payment data for a date range.
type PaymentSummary struct {
	PeriodStart   time.Time     `json:"period_start"`
	PeriodEnd     time.Time     `json:"period_end"`
	TotalReceived float64       `json:"total_received"`
	TotalReversed float64       `json:"total_reversed"`
	PaymentCount  int           `json:"payment_count"`
	ByMethod      []MethodCount `json:"by_method"`
}

// MethodCount holds payment totals for one payment method.
type MethodCount struct {
	Method PaymentMethod `json:"method"`
	Count  int           `json:"count"`
	Amount float64       `json:"amount"`
}

// ReportRepository defines analytics queries for billing reports.
type ReportRepository interface {
	GetPeriodSummary(ctx context.Context, start, end time.Time) (*BillingPeriodSummary, error)
	GetPaymentSummary(ctx context.Context, start, end time.Time) (*PaymentSummary, error)
	ListInvoicesForExport(ctx context.Context, f ListInvoicesFilter) ([]*Invoice, error)
}
