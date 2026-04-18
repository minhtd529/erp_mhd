package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
)

type EngCommissionRepo struct{ pool *pgxpool.Pool }

func NewEngCommissionRepo(pool *pgxpool.Pool) *EngCommissionRepo {
	return &EngCommissionRepo{pool: pool}
}

const ecCols = `id, engagement_id, salesperson_id, role, plan_id,
	rate_type, rate, fixed_amount, tiers, apply_base, trigger_on,
	max_amount, holdback_pct, status, notes, approved_by, approved_at,
	created_by, created_at, updated_at`

func scanEC(row pgx.Row) (*domain.EngagementCommission, error) {
	var ec domain.EngagementCommission
	var tiersJSON []byte
	err := row.Scan(
		&ec.ID, &ec.EngagementID, &ec.SalespersonID, &ec.Role, &ec.PlanID,
		&ec.RateType, &ec.Rate, &ec.FixedAmount, &tiersJSON,
		&ec.ApplyBase, &ec.TriggerOn,
		&ec.MaxAmount, &ec.HoldbackPct, &ec.Status, &ec.Notes,
		&ec.ApprovedBy, &ec.ApprovedAt,
		&ec.CreatedBy, &ec.CreatedAt, &ec.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tiersJSON, &ec.Tiers); err != nil {
		return nil, fmt.Errorf("unmarshal tiers: %w", err)
	}
	if ec.Tiers == nil {
		ec.Tiers = []domain.CommissionTier{}
	}
	return &ec, nil
}

func (r *EngCommissionRepo) Create(ctx context.Context, p domain.CreateEngCommissionParams) (*domain.EngagementCommission, error) {
	tiersJSON, _ := json.Marshal(p.Tiers)
	if tiersJSON == nil {
		tiersJSON = []byte("[]")
	}
	const q = `INSERT INTO engagement_commissions
		(engagement_id, salesperson_id, role, plan_id, rate_type, rate, fixed_amount, tiers,
		 apply_base, trigger_on, max_amount, holdback_pct, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING ` + ecCols
	row := r.pool.QueryRow(ctx, q,
		p.EngagementID, p.SalespersonID, string(p.Role), p.PlanID,
		string(p.RateType), p.Rate, p.FixedAmount, tiersJSON,
		string(p.ApplyBase), string(p.TriggerOn),
		p.MaxAmount, p.HoldbackPct, p.Notes, p.CreatedBy,
	)
	ec, err := scanEC(row)
	if err != nil {
		return nil, fmt.Errorf("engCommission.Create: %w", err)
	}
	return ec, nil
}

func (r *EngCommissionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.EngagementCommission, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+ecCols+` FROM engagement_commissions WHERE id=$1`, id)
	ec, err := scanEC(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEngCommissionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engCommission.FindByID: %w", err)
	}
	return ec, nil
}

func (r *EngCommissionRepo) List(ctx context.Context, f domain.ListEngCommissionsFilter, page, size int) ([]*domain.EngagementCommission, int64, error) {
	where := "WHERE 1=1"
	args := []any{}
	idx := 1

	if f.EngagementID != nil {
		where += fmt.Sprintf(" AND engagement_id=$%d", idx)
		args = append(args, f.EngagementID)
		idx++
	}
	if f.SalespersonID != nil {
		where += fmt.Sprintf(" AND salesperson_id=$%d", idx)
		args = append(args, f.SalespersonID)
		idx++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status=$%d", idx)
		args = append(args, f.Status)
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM engagement_commissions `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("engCommission.List count: %w", err)
	}

	offset := (page - 1) * size
	args = append(args, size, offset)
	q := fmt.Sprintf(`SELECT `+ecCols+` FROM engagement_commissions `+where+` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("engCommission.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.EngagementCommission
	for rows.Next() {
		ec, err := scanEC(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("engCommission.List scan: %w", err)
		}
		list = append(list, ec)
	}
	if list == nil {
		list = []*domain.EngagementCommission{}
	}
	return list, total, rows.Err()
}

func (r *EngCommissionRepo) SumRateByEngagement(ctx context.Context, engagementID uuid.UUID) (float64, error) {
	var sum float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(rate),0) FROM engagement_commissions WHERE engagement_id=$1 AND status='active'`,
		engagementID,
	).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("engCommission.SumRate: %w", err)
	}
	return sum, nil
}

func (r *EngCommissionRepo) ListActiveByTrigger(ctx context.Context, engagementID uuid.UUID, trigger domain.CommissionTrigger) ([]*domain.EngagementCommission, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+ecCols+` FROM engagement_commissions WHERE engagement_id=$1 AND trigger_on=$2 AND status='active'`,
		engagementID, string(trigger),
	)
	if err != nil {
		return nil, fmt.Errorf("engCommission.ListActiveByTrigger: %w", err)
	}
	defer rows.Close()

	var list []*domain.EngagementCommission
	for rows.Next() {
		ec, err := scanEC(rows)
		if err != nil {
			return nil, fmt.Errorf("engCommission.ListActiveByTrigger scan: %w", err)
		}
		list = append(list, ec)
	}
	if list == nil {
		list = []*domain.EngagementCommission{}
	}
	return list, rows.Err()
}

func (r *EngCommissionRepo) SumHoldbackByEngagement(ctx context.Context, engagementID uuid.UUID) (int64, error) {
	var sum int64
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(holdback_amount),0)
		FROM commission_records
		WHERE engagement_id=$1 AND status NOT IN ('cancelled','clawback')
	`, engagementID).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("engCommission.SumHoldback: %w", err)
	}
	return sum, nil
}

func (r *EngCommissionRepo) Cancel(ctx context.Context, id uuid.UUID, _ uuid.UUID) (*domain.EngagementCommission, error) {
	const q = `UPDATE engagement_commissions SET status='cancelled', updated_at=NOW() WHERE id=$1 AND status='active' RETURNING ` + ecCols
	row := r.pool.QueryRow(ctx, q, id)
	ec, err := scanEC(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEngCommissionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engCommission.Cancel: %w", err)
	}
	return ec, nil
}

func (r *EngCommissionRepo) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) (*domain.EngagementCommission, error) {
	const q = `UPDATE engagement_commissions
		SET approved_by=$2, approved_at=NOW(), updated_at=NOW()
		WHERE id=$1
		RETURNING ` + ecCols
	row := r.pool.QueryRow(ctx, q, id, approvedBy)
	ec, err := scanEC(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEngCommissionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("engCommission.Approve: %w", err)
	}
	return ec, nil
}
