package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/reporting/domain"
)

type ReportingRepo struct{ pool *pgxpool.Pool }

func NewReportingRepo(pool *pgxpool.Pool) *ReportingRepo { return &ReportingRepo{pool: pool} }

// ── Materialized view reads ───────────────────────────────────────────────────

func (r *ReportingRepo) GetRevenueByService(ctx context.Context) ([]domain.RevenueByService, error) {
	rows, err := r.pool.Query(ctx, `SELECT service_type, invoice_count, total_revenue, total_tax FROM mv_revenue_by_service ORDER BY total_revenue DESC`)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetRevenueByService: %w", err)
	}
	defer rows.Close()
	var list []domain.RevenueByService
	for rows.Next() {
		var s domain.RevenueByService
		if err := rows.Scan(&s.ServiceType, &s.InvoiceCount, &s.TotalRevenue, &s.TotalTax); err != nil {
			return nil, fmt.Errorf("reporting.GetRevenueByService scan: %w", err)
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *ReportingRepo) GetUtilizationRates(ctx context.Context, month time.Time) ([]domain.UtilizationRate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT staff_id, month, total_hours, utilization_percent FROM mv_utilization_rate WHERE month=$1 ORDER BY utilization_percent DESC`,
		month,
	)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetUtilizationRates: %w", err)
	}
	defer rows.Close()
	var list []domain.UtilizationRate
	for rows.Next() {
		var u domain.UtilizationRate
		if err := rows.Scan(&u.StaffID, &u.Month, &u.TotalHours, &u.UtilizationPercent); err != nil {
			return nil, fmt.Errorf("reporting.GetUtilizationRates scan: %w", err)
		}
		list = append(list, u)
	}
	return list, rows.Err()
}

func (r *ReportingRepo) GetARAgingAll(ctx context.Context) ([]domain.ARAgingRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT client_id, current_amount, days_1_30, days_31_60, days_61_90, days_over_90, total_outstanding FROM mv_ar_aging ORDER BY total_outstanding DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetARAgingAll: %w", err)
	}
	defer rows.Close()
	var list []domain.ARAgingRow
	for rows.Next() {
		var a domain.ARAgingRow
		if err := rows.Scan(&a.ClientID, &a.CurrentAmount, &a.Days1To30, &a.Days31To60, &a.Days61To90, &a.DaysOver90, &a.TotalOutstanding); err != nil {
			return nil, fmt.Errorf("reporting.GetARAgingAll scan: %w", err)
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *ReportingRepo) GetARAgingByClient(ctx context.Context, clientID uuid.UUID) (*domain.ARAgingRow, error) {
	var a domain.ARAgingRow
	err := r.pool.QueryRow(ctx,
		`SELECT client_id, current_amount, days_1_30, days_31_60, days_61_90, days_over_90, total_outstanding FROM mv_ar_aging WHERE client_id=$1`,
		clientID,
	).Scan(&a.ClientID, &a.CurrentAmount, &a.Days1To30, &a.Days31To60, &a.Days61To90, &a.DaysOver90, &a.TotalOutstanding)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetARAgingByClient: %w", err)
	}
	return &a, nil
}

func (r *ReportingRepo) GetEngagementProgress(ctx context.Context, limit int) ([]domain.EngagementProgress, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT engagement_id, client_id, status, budgeted_hours, hours_logged, completion_percent FROM mv_engagement_progress WHERE status='ACTIVE' ORDER BY completion_percent ASC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetEngagementProgress: %w", err)
	}
	defer rows.Close()
	var list []domain.EngagementProgress
	for rows.Next() {
		var e domain.EngagementProgress
		if err := rows.Scan(&e.EngagementID, &e.ClientID, &e.Status, &e.BudgetedHours, &e.HoursLogged, &e.CompletionPercent); err != nil {
			return nil, fmt.Errorf("reporting.GetEngagementProgress scan: %w", err)
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *ReportingRepo) GetCommissionMonthlySummary(ctx context.Context, months int) ([]domain.CommissionMonthlySummary, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT month, total_accrued, total_approved, total_paid, total_pending, total_on_hold, clawback_count FROM mv_commission_summary ORDER BY month DESC LIMIT $1`,
		months,
	)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetCommissionMonthlySummary: %w", err)
	}
	defer rows.Close()
	var list []domain.CommissionMonthlySummary
	for rows.Next() {
		var s domain.CommissionMonthlySummary
		if err := rows.Scan(&s.Month, &s.TotalAccrued, &s.TotalApproved, &s.TotalPaid, &s.TotalPending, &s.TotalOnHold, &s.ClawbackCount); err != nil {
			return nil, fmt.Errorf("reporting.GetCommissionMonthlySummary scan: %w", err)
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

// ── Aggregated metrics ────────────────────────────────────────────────────────

func (r *ReportingRepo) GetRevenueYTD(ctx context.Context, year int) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM invoices WHERE status IN ('PAID','PARTIALLY_PAID') AND is_void=FALSE AND EXTRACT(YEAR FROM issue_date)=$1`,
		year,
	).Scan(&total)
	return total, err
}

func (r *ReportingRepo) GetRevenueMonth(ctx context.Context, year, month int) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM invoices WHERE status IN ('PAID','PARTIALLY_PAID') AND is_void=FALSE AND EXTRACT(YEAR FROM issue_date)=$1 AND EXTRACT(MONTH FROM issue_date)=$2`,
		year, month,
	).Scan(&total)
	return total, err
}

func (r *ReportingRepo) GetTotalOutstandingReceivables(ctx context.Context) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(SUM(total_outstanding), 0) FROM mv_ar_aging`).Scan(&total)
	return total, err
}

func (r *ReportingRepo) GetActiveEngagementsCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM engagements WHERE status='ACTIVE'`).Scan(&count)
	return count, err
}

func (r *ReportingRepo) GetEngagementsByStatus(ctx context.Context) (map[string]int64, error) {
	rows, err := r.pool.Query(ctx, `SELECT status, COUNT(*) FROM engagements GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int64{}
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		result[status] = count
	}
	return result, rows.Err()
}

func (r *ReportingRepo) GetAvgUtilizationMonth(ctx context.Context, year, month int) (float64, error) {
	var avg float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(utilization_percent), 0) FROM mv_utilization_rate WHERE EXTRACT(YEAR FROM month)=$1 AND EXTRACT(MONTH FROM month)=$2`,
		year, month,
	).Scan(&avg)
	return avg, err
}

func (r *ReportingRepo) GetCommissionKPIs(ctx context.Context, year, month int) (*domain.CommissionKPIs, error) {
	const q = `
		SELECT
			COALESCE(SUM(calculated_amount) FILTER (WHERE DATE_PART('year', accrued_at)=$1 AND DATE_PART('month', accrued_at)=$2), 0),
			COALESCE(SUM(payable_amount) FILTER (WHERE status='paid' AND DATE_PART('year', accrued_at)=$1 AND DATE_PART('month', accrued_at)=$2), 0),
			COALESCE(SUM(payable_amount) FILTER (WHERE status IN ('accrued','on_hold')), 0),
			COALESCE(SUM(holdback_amount) FILTER (WHERE status='on_hold'), 0)
		FROM commission_records WHERE is_clawback=FALSE`
	var kpis domain.CommissionKPIs
	err := r.pool.QueryRow(ctx, q, year, month).Scan(
		&kpis.TotalAccruedMonth, &kpis.TotalPaidMonth, &kpis.TotalPending, &kpis.TotalOnHold,
	)
	return &kpis, err
}

func (r *ReportingRepo) GetCommissionByStaff(ctx context.Context, staffID uuid.UUID, year int) (*domain.CommissionKPIs, error) {
	const q = `
		SELECT
			COALESCE(SUM(calculated_amount) FILTER (WHERE DATE_PART('year', accrued_at)=$2), 0),
			COALESCE(SUM(payable_amount) FILTER (WHERE status='paid' AND DATE_PART('year', accrued_at)=$2), 0),
			COALESCE(SUM(payable_amount) FILTER (WHERE status IN ('accrued','on_hold')), 0),
			COALESCE(SUM(holdback_amount) FILTER (WHERE status='on_hold'), 0)
		FROM commission_records WHERE salesperson_id=$1 AND is_clawback=FALSE`
	var kpis domain.CommissionKPIs
	err := r.pool.QueryRow(ctx, q, staffID, year).Scan(
		&kpis.TotalAccruedMonth, &kpis.TotalPaidMonth, &kpis.TotalPending, &kpis.TotalOnHold,
	)
	return &kpis, err
}

func (r *ReportingRepo) GetTeamSize(ctx context.Context, managerID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT em.user_id) FROM engagement_members em JOIN engagements e ON e.id=em.engagement_id WHERE e.manager_id=$1 AND e.status='ACTIVE'`,
		managerID,
	).Scan(&count)
	return count, err
}

func (r *ReportingRepo) GetTeamUtilization(ctx context.Context, managerID uuid.UUID, year, month int) (float64, error) {
	var avg float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(AVG(u.utilization_percent), 0)
		 FROM mv_utilization_rate u
		 JOIN engagement_members em ON em.user_id = u.staff_id
		 JOIN engagements e ON e.id = em.engagement_id
		 WHERE e.manager_id=$1 AND EXTRACT(YEAR FROM u.month)=$2 AND EXTRACT(MONTH FROM u.month)=$3`,
		managerID, year, month,
	).Scan(&avg)
	return avg, err
}

func (r *ReportingRepo) GetTeamEngagementProgress(ctx context.Context, managerID uuid.UUID) ([]domain.EngagementProgress, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ep.engagement_id, ep.client_id, ep.status, ep.budgeted_hours, ep.hours_logged, ep.completion_percent
		 FROM mv_engagement_progress ep
		 JOIN engagements e ON e.id=ep.engagement_id
		 WHERE e.manager_id=$1
		 ORDER BY ep.completion_percent ASC LIMIT 20`,
		managerID,
	)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetTeamEngagementProgress: %w", err)
	}
	defer rows.Close()
	var list []domain.EngagementProgress
	for rows.Next() {
		var e domain.EngagementProgress
		if err := rows.Scan(&e.EngagementID, &e.ClientID, &e.Status, &e.BudgetedHours, &e.HoursLogged, &e.CompletionPercent); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (r *ReportingRepo) GetTeamOutstandingReceivables(ctx context.Context, managerID uuid.UUID) (int64, error) {
	var total int64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(a.total_outstanding), 0) FROM mv_ar_aging a JOIN engagements e ON e.client_id=a.client_id WHERE e.manager_id=$1 AND e.status='ACTIVE'`,
		managerID,
	).Scan(&total)
	return total, err
}

func (r *ReportingRepo) GetStaffActiveEngagements(ctx context.Context, staffID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(DISTINCT em.engagement_id) FROM engagement_members em JOIN engagements e ON e.id=em.engagement_id WHERE em.user_id=$1 AND e.status='ACTIVE'`,
		staffID,
	).Scan(&count)
	return count, err
}

func (r *ReportingRepo) GetStaffHoursMonth(ctx context.Context, staffID uuid.UUID, year, month int) (float64, error) {
	var hours float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(total_hours, 0) FROM mv_utilization_rate WHERE staff_id=$1 AND EXTRACT(YEAR FROM month)=$2 AND EXTRACT(MONTH FROM month)=$3`,
		staffID, year, month,
	).Scan(&hours)
	return hours, err
}

func (r *ReportingRepo) IsSalesperson(ctx context.Context, staffID uuid.UUID) (bool, error) {
	var isSp bool
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(is_salesperson, FALSE) FROM employees WHERE user_id=$1`,
		staffID,
	).Scan(&isSp)
	return isSp, err
}

func (r *ReportingRepo) GetRevenueByStaff(ctx context.Context, f domain.ReportFilter) ([]domain.RevenueByStaffRow, error) {
	q := `
		SELECT e.primary_salesperson_id, COALESCE(SUM(i.total_amount), 0), COUNT(DISTINCT i.id), COUNT(DISTINCT e.id)
		FROM invoices i
		JOIN engagements e ON e.id = i.engagement_id
		WHERE i.status IN ('PAID','PARTIALLY_PAID') AND i.is_void=FALSE AND e.primary_salesperson_id IS NOT NULL`
	args := []any{}
	idx := 1
	if f.Year > 0 {
		q += fmt.Sprintf(" AND EXTRACT(YEAR FROM i.issue_date)=$%d", idx)
		args = append(args, f.Year)
		idx++
	}
	if f.Month > 0 {
		q += fmt.Sprintf(" AND EXTRACT(MONTH FROM i.issue_date)=$%d", idx)
		args = append(args, f.Month)
		idx++
	}
	q += " GROUP BY e.primary_salesperson_id ORDER BY 2 DESC LIMIT 50"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetRevenueByStaff: %w", err)
	}
	defer rows.Close()
	var list []domain.RevenueByStaffRow
	for rows.Next() {
		var row domain.RevenueByStaffRow
		if err := rows.Scan(&row.StaffID, &row.TotalRevenue, &row.InvoiceCount, &row.EngagementCount); err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

// ── MV refresh ────────────────────────────────────────────────────────────────

func (r *ReportingRepo) RefreshAllViews(ctx context.Context) error {
	views := []string{
		"mv_revenue_by_service",
		"mv_utilization_rate",
		"mv_ar_aging",
		"mv_engagement_progress",
		"mv_commission_summary",
	}
	for _, v := range views {
		if _, err := r.pool.Exec(ctx, fmt.Sprintf("REFRESH MATERIALIZED VIEW CONCURRENTLY %s", v)); err != nil {
			// Log failure but continue; some views may succeed
			_, _ = r.pool.Exec(ctx,
				`INSERT INTO mv_refresh_log (view_name, success, error_msg) VALUES ($1, FALSE, $2) ON CONFLICT (view_name) DO UPDATE SET success=FALSE, error_msg=$2, refreshed_at=NOW()`,
				v, err.Error(),
			)
			return fmt.Errorf("reporting.RefreshAllViews %s: %w", v, err)
		}
		_, _ = r.pool.Exec(ctx,
			`INSERT INTO mv_refresh_log (view_name, success) VALUES ($1, TRUE) ON CONFLICT (view_name) DO UPDATE SET success=TRUE, error_msg=NULL, refreshed_at=NOW()`,
			v,
		)
	}
	return nil
}

// ── Commission reports ────────────────────────────────────────────────────────

func (r *ReportingRepo) GetCommissionPayoutReport(ctx context.Context, months int) ([]domain.CommissionPayoutRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			DATE_TRUNC('month', accrued_at)        AS month,
			COALESCE(SUM(CASE WHEN status IN ('approved','paid') THEN payable_amount ELSE 0 END), 0) AS total_approved,
			COALESCE(SUM(CASE WHEN status = 'paid' THEN payable_amount ELSE 0 END), 0)              AS total_paid,
			COUNT(*)                                                                                  AS record_count
		FROM commission_records
		WHERE is_clawback = FALSE
		  AND accrued_at >= NOW() - ($1 * INTERVAL '1 month')
		GROUP BY DATE_TRUNC('month', accrued_at)
		ORDER BY month DESC`,
		months,
	)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetCommissionPayoutReport: %w", err)
	}
	defer rows.Close()
	var list []domain.CommissionPayoutRow
	for rows.Next() {
		var r domain.CommissionPayoutRow
		if err := rows.Scan(&r.Month, &r.TotalApproved, &r.TotalPaid, &r.RecordCount); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (r *ReportingRepo) GetCommissionByServiceReport(ctx context.Context, year int) ([]domain.CommissionByServiceRow, error) {
	q := `
		SELECT
			COALESCE(e.service_type, 'UNKNOWN')       AS service_type,
			COALESCE(SUM(cr.calculated_amount), 0)    AS total_accrued,
			COALESCE(SUM(cr.payable_amount), 0)       AS total_payable,
			COALESCE(SUM(CASE WHEN cr.status='paid' THEN cr.payable_amount ELSE 0 END), 0) AS total_paid,
			COUNT(cr.id)                               AS record_count,
			COALESCE(AVG(cr.rate), 0)                  AS avg_rate
		FROM commission_records cr
		JOIN engagements e ON e.id = cr.engagement_id
		WHERE cr.is_clawback = FALSE`
	args := []any{}
	if year > 0 {
		q += ` AND EXTRACT(YEAR FROM cr.accrued_at) = $1`
		args = append(args, year)
	}
	q += ` GROUP BY e.service_type ORDER BY total_accrued DESC`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetCommissionByServiceReport: %w", err)
	}
	defer rows.Close()
	var list []domain.CommissionByServiceRow
	for rows.Next() {
		var row domain.CommissionByServiceRow
		if err := rows.Scan(&row.ServiceType, &row.TotalAccrued, &row.TotalPayable, &row.TotalPaid, &row.RecordCount, &row.AvgRate); err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, rows.Err()
}

func (r *ReportingRepo) GetCommissionPendingReport(ctx context.Context) (*domain.CommissionPendingSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, salesperson_id, engagement_id, status, payable_amount, accrued_at, approved_at
		FROM commission_records
		WHERE status IN ('accrued', 'approved') AND is_clawback = FALSE
		ORDER BY accrued_at ASC
		LIMIT 500`,
	)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetCommissionPendingReport: %w", err)
	}
	defer rows.Close()

	var records []domain.CommissionPendingRow
	var pendingApproval, pendingPayout int64
	for rows.Next() {
		var rec domain.CommissionPendingRow
		if err := rows.Scan(&rec.RecordID, &rec.SalespersonID, &rec.EngagementID, &rec.Status, &rec.PayableAmount, &rec.AccruedAt, &rec.ApprovedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
		if rec.Status == "accrued" {
			pendingApproval += rec.PayableAmount
		} else {
			pendingPayout += rec.PayableAmount
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &domain.CommissionPendingSummary{
		TotalPendingApproval: pendingApproval,
		TotalPendingPayout:   pendingPayout,
		Records:              records,
	}, nil
}

func (r *ReportingRepo) GetCommissionClawbackReport(ctx context.Context, months int) (*domain.CommissionClawbackSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, salesperson_id, engagement_id, payable_amount, clawback_reason, accrued_at
		FROM commission_records
		WHERE (is_clawback = TRUE OR status = 'clawback')
		  AND accrued_at >= NOW() - ($1 * INTERVAL '1 month')
		ORDER BY accrued_at DESC`,
		months,
	)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetCommissionClawbackReport: %w", err)
	}
	defer rows.Close()

	var records []domain.CommissionClawbackRow
	var total int64
	for rows.Next() {
		var rec domain.CommissionClawbackRow
		if err := rows.Scan(&rec.RecordID, &rec.SalespersonID, &rec.EngagementID, &rec.ClawbackAmount, &rec.Reason, &rec.ClawbackAt); err != nil {
			return nil, err
		}
		// clawback amounts are stored as negative values; take abs for display
		if rec.ClawbackAmount < 0 {
			rec.ClawbackAmount = -rec.ClawbackAmount
		}
		records = append(records, rec)
		total += rec.ClawbackAmount
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &domain.CommissionClawbackSummary{
		TotalClawback: total,
		RecordCount:   int64(len(records)),
		Records:       records,
	}, nil
}

func (r *ReportingRepo) GetLastRefreshLog(ctx context.Context) ([]domain.MVRefreshLog, error) {
	rows, err := r.pool.Query(ctx, `SELECT view_name, refreshed_at, COALESCE(duration_ms,0), success, error_msg FROM mv_refresh_log ORDER BY view_name`)
	if err != nil {
		return nil, fmt.Errorf("reporting.GetLastRefreshLog: %w", err)
	}
	defer rows.Close()
	var list []domain.MVRefreshLog
	for rows.Next() {
		var l domain.MVRefreshLog
		if err := rows.Scan(&l.ViewName, &l.RefreshedAt, &l.DurationMs, &l.Success, &l.ErrorMsg); err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, rows.Err()
}
