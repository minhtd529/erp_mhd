package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
)

// BillingReader provides read-only access to billing tables for commission accrual.
type BillingReader struct{ pool *pgxpool.Pool }

func NewBillingReader(pool *pgxpool.Pool) *BillingReader { return &BillingReader{pool: pool} }

func (r *BillingReader) GetInvoiceForAccrual(ctx context.Context, invoiceID uuid.UUID) (*domain.InvoiceAccrualData, error) {
	var d domain.InvoiceAccrualData
	err := r.pool.QueryRow(ctx,
		`SELECT id, engagement_id, total_amount FROM invoices WHERE id=$1 AND is_deleted=false`,
		invoiceID,
	).Scan(&d.InvoiceID, &d.EngagementID, &d.TotalAmount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("billing_reader: invoice %s not found", invoiceID)
	}
	if err != nil {
		return nil, fmt.Errorf("billing_reader.GetInvoiceForAccrual: %w", err)
	}
	return &d, nil
}

func (r *BillingReader) GetPaymentForAccrual(ctx context.Context, paymentID uuid.UUID) (*domain.PaymentAccrualData, error) {
	var d domain.PaymentAccrualData
	err := r.pool.QueryRow(ctx, `
		SELECT p.id, p.invoice_id, i.engagement_id, p.amount
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		WHERE p.id=$1 AND i.is_deleted=false
	`, paymentID).Scan(&d.PaymentID, &d.InvoiceID, &d.EngagementID, &d.Amount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("billing_reader: payment %s not found", paymentID)
	}
	if err != nil {
		return nil, fmt.Errorf("billing_reader.GetPaymentForAccrual: %w", err)
	}
	return &d, nil
}
