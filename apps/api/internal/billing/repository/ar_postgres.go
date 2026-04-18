package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
)

// ARRepo provides accounts receivable analytics queries.
type ARRepo struct{ pool *pgxpool.Pool }

func NewARRepo(pool *pgxpool.Pool) *ARRepo { return &ARRepo{pool: pool} }

// GetAging returns the AR aging breakdown grouped by client.
// Buckets: current (not yet due), 1-30, 31-60, 61-90, 90+ days overdue.
func (r *ARRepo) GetAging(ctx context.Context) ([]*domain.ARAgingRow, error) {
	const q = `
		SELECT
			i.client_id,
			SUM(CASE WHEN i.due_date IS NULL OR i.due_date >= CURRENT_DATE
			         THEN i.total_amount - COALESCE(paid.total_paid, 0) ELSE 0 END) AS current_amount,
			SUM(CASE WHEN i.due_date < CURRENT_DATE AND i.due_date >= CURRENT_DATE - INTERVAL '30 days'
			         THEN i.total_amount - COALESCE(paid.total_paid, 0) ELSE 0 END) AS days_1_30,
			SUM(CASE WHEN i.due_date < CURRENT_DATE - INTERVAL '30 days' AND i.due_date >= CURRENT_DATE - INTERVAL '60 days'
			         THEN i.total_amount - COALESCE(paid.total_paid, 0) ELSE 0 END) AS days_31_60,
			SUM(CASE WHEN i.due_date < CURRENT_DATE - INTERVAL '60 days' AND i.due_date >= CURRENT_DATE - INTERVAL '90 days'
			         THEN i.total_amount - COALESCE(paid.total_paid, 0) ELSE 0 END) AS days_61_90,
			SUM(CASE WHEN i.due_date < CURRENT_DATE - INTERVAL '90 days'
			         THEN i.total_amount - COALESCE(paid.total_paid, 0) ELSE 0 END) AS days_90_plus,
			SUM(i.total_amount - COALESCE(paid.total_paid, 0)) AS total_outstanding
		FROM invoices i
		LEFT JOIN (
			SELECT invoice_id, SUM(amount) AS total_paid
			FROM payments
			WHERE status NOT IN ('REVERSED')
			GROUP BY invoice_id
		) paid ON paid.invoice_id = i.id
		WHERE i.is_deleted = false
		  AND i.status IN ('ISSUED', 'SENT', 'CONFIRMED')
		  AND (i.total_amount - COALESCE(paid.total_paid, 0)) > 0
		GROUP BY i.client_id
		ORDER BY total_outstanding DESC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("ar.GetAging: %w", err)
	}
	defer rows.Close()

	var result []*domain.ARAgingRow
	for rows.Next() {
		row := &domain.ARAgingRow{}
		if err := rows.Scan(
			&row.ClientID, &row.Current, &row.Days1to30, &row.Days31to60,
			&row.Days61to90, &row.Days90Plus, &row.TotalOutstanding,
		); err != nil {
			return nil, fmt.Errorf("ar.GetAging scan: %w", err)
		}
		result = append(result, row)
	}
	if result == nil {
		result = []*domain.ARAgingRow{}
	}
	return result, rows.Err()
}

// GetOutstanding returns total outstanding balances grouped by client.
func (r *ARRepo) GetOutstanding(ctx context.Context) ([]*domain.AROutstandingRow, error) {
	const q = `
		SELECT
			i.client_id,
			COUNT(i.id) AS invoice_count,
			SUM(i.total_amount) AS total_billed,
			COALESCE(SUM(paid.total_paid), 0) AS total_paid,
			SUM(i.total_amount) - COALESCE(SUM(paid.total_paid), 0) AS total_outstanding
		FROM invoices i
		LEFT JOIN (
			SELECT invoice_id, SUM(amount) AS total_paid
			FROM payments
			WHERE status NOT IN ('REVERSED')
			GROUP BY invoice_id
		) paid ON paid.invoice_id = i.id
		WHERE i.is_deleted = false
		  AND i.status IN ('ISSUED', 'SENT', 'CONFIRMED')
		GROUP BY i.client_id
		HAVING SUM(i.total_amount) - COALESCE(SUM(paid.total_paid), 0) > 0
		ORDER BY total_outstanding DESC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("ar.GetOutstanding: %w", err)
	}
	defer rows.Close()

	var result []*domain.AROutstandingRow
	for rows.Next() {
		row := &domain.AROutstandingRow{}
		if err := rows.Scan(
			&row.ClientID, &row.InvoiceCount, &row.TotalBilled, &row.TotalPaid, &row.TotalOutstanding,
		); err != nil {
			return nil, fmt.Errorf("ar.GetOutstanding scan: %w", err)
		}
		result = append(result, row)
	}
	if result == nil {
		result = []*domain.AROutstandingRow{}
	}
	return result, rows.Err()
}
