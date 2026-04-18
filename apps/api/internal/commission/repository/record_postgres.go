package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
)

type RecordRepo struct{ pool *pgxpool.Pool }

func NewRecordRepo(pool *pgxpool.Pool) *RecordRepo { return &RecordRepo{pool: pool} }

const recCols = `id, engagement_commission_id, engagement_id, salesperson_id,
	invoice_id, payment_id,
	base_amount, rate, calculated_amount, holdback_amount, payable_amount,
	status, accrued_at, approved_by, approved_at, paid_at,
	paid_by_payroll_id, payout_reference,
	clawback_record_id, is_clawback, clawback_reason,
	notes, created_at, updated_at`

func scanRecord(row pgx.Row) (*domain.CommissionRecord, error) {
	var r domain.CommissionRecord
	err := row.Scan(
		&r.ID, &r.EngagementCommissionID, &r.EngagementID, &r.SalespersonID,
		&r.InvoiceID, &r.PaymentID,
		&r.BaseAmount, &r.Rate, &r.CalculatedAmount, &r.HoldbackAmount, &r.PayableAmount,
		&r.Status, &r.AccruedAt, &r.ApprovedBy, &r.ApprovedAt, &r.PaidAt,
		&r.PaidByPayrollID, &r.PayoutReference,
		&r.ClawbackRecordID, &r.IsClawback, &r.ClawbackReason,
		&r.Notes, &r.CreatedAt, &r.UpdatedAt,
	)
	return &r, err
}

func (r *RecordRepo) Create(ctx context.Context, rec *domain.CommissionRecord) (*domain.CommissionRecord, error) {
	const q = `INSERT INTO commission_records
		(engagement_commission_id, engagement_id, salesperson_id, invoice_id, payment_id,
		 base_amount, rate, calculated_amount, holdback_amount, payable_amount,
		 status, clawback_record_id, is_clawback, clawback_reason, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING ` + recCols
	row := r.pool.QueryRow(ctx, q,
		rec.EngagementCommissionID, rec.EngagementID, rec.SalespersonID,
		rec.InvoiceID, rec.PaymentID,
		rec.BaseAmount, rec.Rate, rec.CalculatedAmount, rec.HoldbackAmount, rec.PayableAmount,
		string(rec.Status), rec.ClawbackRecordID, rec.IsClawback, rec.ClawbackReason, rec.Notes,
	)
	created, err := scanRecord(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrDuplicateAccrual
		}
		return nil, fmt.Errorf("record.Create: %w", err)
	}
	return created, nil
}

func (r *RecordRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.CommissionRecord, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+recCols+` FROM commission_records WHERE id=$1`, id)
	rec, err := scanRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCommissionRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("record.FindByID: %w", err)
	}
	return rec, nil
}

func (r *RecordRepo) List(ctx context.Context, f domain.ListRecordsFilter, page, size int) ([]*domain.CommissionRecord, int64, error) {
	where := "WHERE 1=1"
	args := []any{}
	idx := 1

	if f.SalespersonID != nil {
		where += fmt.Sprintf(" AND salesperson_id=$%d", idx)
		args = append(args, f.SalespersonID)
		idx++
	}
	if f.EngagementID != nil {
		where += fmt.Sprintf(" AND engagement_id=$%d", idx)
		args = append(args, f.EngagementID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status=$%d", idx)
		args = append(args, string(f.Status))
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM commission_records `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("record.List count: %w", err)
	}

	offset := (page - 1) * size
	args = append(args, size, offset)
	q := fmt.Sprintf(`SELECT `+recCols+` FROM commission_records `+where+` ORDER BY accrued_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("record.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.CommissionRecord
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("record.List scan: %w", err)
		}
		list = append(list, rec)
	}
	if list == nil {
		list = []*domain.CommissionRecord{}
	}
	return list, total, rows.Err()
}

func (r *RecordRepo) ListForStatement(ctx context.Context, f domain.StatementFilter) ([]*domain.CommissionRecord, error) {
	q := `SELECT ` + recCols + ` FROM commission_records
		WHERE salesperson_id = $1 AND is_clawback = false
		  AND accrued_at >= $2 AND accrued_at < $3
		ORDER BY accrued_at ASC`
	rows, err := r.pool.Query(ctx, q, f.SalespersonID, f.From, f.To)
	if err != nil {
		return nil, fmt.Errorf("record.ListForStatement: %w", err)
	}
	defer rows.Close()
	var list []*domain.CommissionRecord
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("record.ListForStatement scan: %w", err)
		}
		list = append(list, rec)
	}
	if list == nil {
		list = []*domain.CommissionRecord{}
	}
	return list, rows.Err()
}

func (r *RecordRepo) ListByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*domain.CommissionRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+recCols+` FROM commission_records WHERE invoice_id=$1`, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("record.ListByInvoice: %w", err)
	}
	defer rows.Close()
	var list []*domain.CommissionRecord
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("record.ListByInvoice scan: %w", err)
		}
		list = append(list, rec)
	}
	if list == nil {
		list = []*domain.CommissionRecord{}
	}
	return list, rows.Err()
}

func (r *RecordRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.CommissionStatus, approvedBy *uuid.UUID, payoutRef string) (*domain.CommissionRecord, error) {
	const q = `UPDATE commission_records
		SET status=$2, approved_by=$3,
		    approved_at=CASE WHEN $2='approved' THEN NOW() ELSE approved_at END,
		    paid_at=CASE WHEN $2='paid' THEN NOW() ELSE paid_at END,
		    payout_reference=CASE WHEN $4 != '' THEN $4 ELSE payout_reference END,
		    updated_at=NOW()
		WHERE id=$1
		RETURNING ` + recCols
	row := r.pool.QueryRow(ctx, q, id, string(status), approvedBy, payoutRef)
	rec, err := scanRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCommissionRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("record.UpdateStatus: %w", err)
	}
	return rec, nil
}

func (r *RecordRepo) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status domain.CommissionStatus, approvedBy *uuid.UUID, payoutRef string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	placeholders := make([]string, len(ids))
	args := []any{string(status), approvedBy, payoutRef}
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+4)
		args = append(args, id)
	}
	q := fmt.Sprintf(`UPDATE commission_records
		SET status=$1, approved_by=$2,
		    approved_at=CASE WHEN $1='approved' THEN NOW() ELSE approved_at END,
		    paid_at=CASE WHEN $1='paid' THEN NOW() ELSE paid_at END,
		    payout_reference=CASE WHEN $3 != '' THEN $3 ELSE payout_reference END,
		    updated_at=NOW()
		WHERE id IN (%s)`, strings.Join(placeholders, ","))

	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return 0, fmt.Errorf("record.BulkUpdateStatus: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *RecordRepo) SummarySalesperson(ctx context.Context, salespersonID uuid.UUID) (*domain.SalespersonSummary, error) {
	const q = `
		SELECT
			COALESCE(SUM(payable_amount) FILTER (WHERE DATE_PART('year', accrued_at) = DATE_PART('year', NOW())), 0),
			COALESCE(SUM(payable_amount) FILTER (WHERE DATE_TRUNC('month', accrued_at) = DATE_TRUNC('month', NOW())), 0),
			COALESCE(SUM(payable_amount) FILTER (WHERE status = 'accrued'), 0),
			COALESCE(SUM(payable_amount) FILTER (WHERE status = 'on_hold'), 0),
			COALESCE(SUM(payable_amount) FILTER (WHERE status = 'approved'), 0),
			COALESCE(SUM(payable_amount) FILTER (WHERE status = 'paid'), 0)
		FROM commission_records
		WHERE salesperson_id = $1 AND is_clawback = false`
	var s domain.SalespersonSummary
	err := r.pool.QueryRow(ctx, q, salespersonID).Scan(
		&s.TotalYTD, &s.TotalMonth, &s.TotalPending, &s.TotalOnHold, &s.TotalApproved, &s.TotalPaid,
	)
	if err != nil {
		return nil, fmt.Errorf("record.SummarySalesperson: %w", err)
	}
	return &s, nil
}

func (r *RecordRepo) ListByTeam(ctx context.Context, managerID uuid.UUID, page, size int) ([]*domain.CommissionRecord, int64, error) {
	// Returns commission records for all salespersons who are team members of engagements
	// managed by the given manager_id.
	const countQ = `
		SELECT COUNT(cr.id)
		FROM commission_records cr
		JOIN engagements e ON e.id = cr.engagement_id
		WHERE e.manager_id = $1`
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, managerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("record.ListByTeam count: %w", err)
	}

	offset := (page - 1) * size
	q := `SELECT ` + recCols + `
		FROM commission_records cr
		JOIN engagements e ON e.id = cr.engagement_id
		WHERE e.manager_id = $1
		ORDER BY cr.accrued_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, q, managerID, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("record.ListByTeam: %w", err)
	}
	defer rows.Close()

	var list []*domain.CommissionRecord
	for rows.Next() {
		rec, err := scanRecord(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("record.ListByTeam scan: %w", err)
		}
		list = append(list, rec)
	}
	if list == nil {
		list = []*domain.CommissionRecord{}
	}
	return list, total, rows.Err()
}
