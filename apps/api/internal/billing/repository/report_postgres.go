package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
)

// ReportRepo provides billing analytics queries for reports.
type ReportRepo struct{ pool *pgxpool.Pool }

func NewReportRepo(pool *pgxpool.Pool) *ReportRepo { return &ReportRepo{pool: pool} }

// GetPeriodSummary returns aggregate invoice data for a date range.
func (r *ReportRepo) GetPeriodSummary(ctx context.Context, start, end time.Time) (*domain.BillingPeriodSummary, error) {
	summary := &domain.BillingPeriodSummary{PeriodStart: start, PeriodEnd: end}

	// Totals
	const totalsQ = `
		SELECT
			COUNT(*) AS invoice_count,
			COALESCE(SUM(total_amount), 0) AS total_invoiced,
			COUNT(*) FILTER (WHERE status = 'PAID') AS paid_count,
			COUNT(*) FILTER (WHERE status NOT IN ('PAID','CANCELLED') AND due_date < CURRENT_DATE) AS overdue_count
		FROM invoices
		WHERE is_deleted = false
		  AND created_at >= $1 AND created_at < $2`

	err := r.pool.QueryRow(ctx, totalsQ, start, end).Scan(
		&summary.InvoiceCount, &summary.TotalInvoiced,
		&summary.PaidCount, &summary.OverdueCount,
	)
	if err != nil {
		return nil, fmt.Errorf("report.GetPeriodSummary totals: %w", err)
	}

	// Total paid via payments
	const paidQ = `
		SELECT COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		WHERE p.status NOT IN ('REVERSED')
		  AND i.is_deleted = false
		  AND i.created_at >= $1 AND i.created_at < $2`
	if err := r.pool.QueryRow(ctx, paidQ, start, end).Scan(&summary.TotalPaid); err != nil {
		return nil, fmt.Errorf("report.GetPeriodSummary paid: %w", err)
	}
	summary.TotalOutstanding = summary.TotalInvoiced - summary.TotalPaid

	// Breakdown by status
	const statusQ = `
		SELECT status, COUNT(*) AS cnt, COALESCE(SUM(total_amount), 0) AS amt
		FROM invoices
		WHERE is_deleted = false
		  AND created_at >= $1 AND created_at < $2
		GROUP BY status
		ORDER BY status`

	rows, err := r.pool.Query(ctx, statusQ, start, end)
	if err != nil {
		return nil, fmt.Errorf("report.GetPeriodSummary byStatus: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sc domain.StatusCount
		var status string
		if err := rows.Scan(&status, &sc.Count, &sc.Amount); err != nil {
			return nil, fmt.Errorf("report.GetPeriodSummary byStatus scan: %w", err)
		}
		sc.Status = domain.InvoiceStatus(status)
		summary.ByStatus = append(summary.ByStatus, sc)
	}
	if summary.ByStatus == nil {
		summary.ByStatus = []domain.StatusCount{}
	}

	return summary, rows.Err()
}

// GetPaymentSummary returns aggregate payment data for a date range.
func (r *ReportRepo) GetPaymentSummary(ctx context.Context, start, end time.Time) (*domain.PaymentSummary, error) {
	summary := &domain.PaymentSummary{PeriodStart: start, PeriodEnd: end}

	const totalsQ = `
		SELECT
			COUNT(*) FILTER (WHERE status NOT IN ('REVERSED')) AS payment_count,
			COALESCE(SUM(amount) FILTER (WHERE status NOT IN ('REVERSED')), 0) AS total_received,
			COALESCE(SUM(amount) FILTER (WHERE status = 'REVERSED'), 0) AS total_reversed
		FROM payments
		WHERE created_at >= $1 AND created_at < $2`

	err := r.pool.QueryRow(ctx, totalsQ, start, end).Scan(
		&summary.PaymentCount, &summary.TotalReceived, &summary.TotalReversed,
	)
	if err != nil {
		return nil, fmt.Errorf("report.GetPaymentSummary totals: %w", err)
	}

	const methodQ = `
		SELECT payment_method, COUNT(*) AS cnt, COALESCE(SUM(amount), 0) AS amt
		FROM payments
		WHERE status NOT IN ('REVERSED')
		  AND created_at >= $1 AND created_at < $2
		GROUP BY payment_method
		ORDER BY amt DESC`

	rows, err := r.pool.Query(ctx, methodQ, start, end)
	if err != nil {
		return nil, fmt.Errorf("report.GetPaymentSummary byMethod: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var mc domain.MethodCount
		var method string
		if err := rows.Scan(&method, &mc.Count, &mc.Amount); err != nil {
			return nil, fmt.Errorf("report.GetPaymentSummary byMethod scan: %w", err)
		}
		mc.Method = domain.PaymentMethod(method)
		summary.ByMethod = append(summary.ByMethod, mc)
	}
	if summary.ByMethod == nil {
		summary.ByMethod = []domain.MethodCount{}
	}

	return summary, rows.Err()
}

// ListInvoicesForExport returns all invoices matching the filter (no pagination) for CSV export.
func (r *ReportRepo) ListInvoicesForExport(ctx context.Context, f domain.ListInvoicesFilter) ([]*domain.Invoice, error) {
	args := []any{}
	where := "WHERE is_deleted = false"
	idx := 1

	if f.ClientID != nil {
		where += fmt.Sprintf(" AND client_id = $%d", idx)
		args = append(args, f.ClientID)
		idx++
	}
	if f.EngagementID != nil {
		where += fmt.Sprintf(" AND engagement_id = $%d", idx)
		args = append(args, f.EngagementID)
		idx++
	}
	if len(f.Statuses) > 0 {
		placeholders := ""
		for i, s := range f.Statuses {
			if i > 0 {
				placeholders += ","
			}
			placeholders += fmt.Sprintf("$%d", idx)
			args = append(args, string(s))
			idx++
		}
		where += " AND status IN (" + placeholders + ")"
	} else if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, string(f.Status))
		idx++
	}
	_ = idx

	q := `SELECT ` + invoiceCols + ` FROM invoices ` + where + ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("report.ListForExport: %w", err)
	}
	defer rows.Close()

	var list []*domain.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, fmt.Errorf("report.ListForExport scan: %w", err)
		}
		list = append(list, inv)
	}
	if list == nil {
		list = []*domain.Invoice{}
	}
	return list, rows.Err()
}
