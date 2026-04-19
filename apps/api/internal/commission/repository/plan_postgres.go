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

type PlanRepo struct{ pool *pgxpool.Pool }

func NewPlanRepo(pool *pgxpool.Pool) *PlanRepo { return &PlanRepo{pool: pool} }

const planCols = `id, code, name, description, type, default_rate, tiers, apply_base, trigger_on, service_types, is_active, created_by, created_at, updated_at, updated_by`

func scanPlan(row pgx.Row) (*domain.CommissionPlan, error) {
	var p domain.CommissionPlan
	var tiersJSON, serviceTypesJSON []byte
	err := row.Scan(
		&p.ID, &p.Code, &p.Name, &p.Description,
		&p.Type, &p.DefaultRate, &tiersJSON,
		&p.ApplyBase, &p.TriggerOn, &serviceTypesJSON,
		&p.IsActive, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &p.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tiersJSON, &p.Tiers); err != nil {
		return nil, fmt.Errorf("unmarshal tiers: %w", err)
	}
	if err := json.Unmarshal(serviceTypesJSON, &p.ServiceTypes); err != nil {
		return nil, fmt.Errorf("unmarshal service_types: %w", err)
	}
	if p.Tiers == nil {
		p.Tiers = []domain.CommissionTier{}
	}
	if p.ServiceTypes == nil {
		p.ServiceTypes = []string{}
	}
	return &p, nil
}

func (r *PlanRepo) Create(ctx context.Context, p domain.CreatePlanParams) (*domain.CommissionPlan, error) {
	tiersJSON, _ := json.Marshal(p.Tiers)
	stJSON, _ := json.Marshal(p.ServiceTypes)
	if tiersJSON == nil {
		tiersJSON = []byte("[]")
	}
	if stJSON == nil {
		stJSON = []byte("[]")
	}

	const q = `INSERT INTO commission_plans
		(code, name, description, type, default_rate, tiers, apply_base, trigger_on, service_types, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING ` + planCols
	row := r.pool.QueryRow(ctx, q,
		p.Code, p.Name, p.Description, string(p.Type), p.DefaultRate,
		tiersJSON, string(p.ApplyBase), string(p.TriggerOn), stJSON, p.CreatedBy,
	)
	plan, err := scanPlan(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrPlanCodeConflict
		}
		return nil, fmt.Errorf("plan.Create: %w", err)
	}
	return plan, nil
}

func (r *PlanRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.CommissionPlan, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+planCols+` FROM commission_plans WHERE id=$1`, id)
	plan, err := scanPlan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPlanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("plan.FindByID: %w", err)
	}
	return plan, nil
}

func (r *PlanRepo) FindByCode(ctx context.Context, code string) (*domain.CommissionPlan, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+planCols+` FROM commission_plans WHERE code=$1`, code)
	plan, err := scanPlan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPlanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("plan.FindByCode: %w", err)
	}
	return plan, nil
}

func (r *PlanRepo) Update(ctx context.Context, p domain.UpdatePlanParams) (*domain.CommissionPlan, error) {
	tiersJSON, _ := json.Marshal(p.Tiers)
	stJSON, _ := json.Marshal(p.ServiceTypes)
	if tiersJSON == nil {
		tiersJSON = []byte("[]")
	}
	if stJSON == nil {
		stJSON = []byte("[]")
	}
	const q = `UPDATE commission_plans
		SET name=$2, description=$3, default_rate=$4, tiers=$5, apply_base=$6, trigger_on=$7, service_types=$8, updated_by=$9, updated_at=NOW()
		WHERE id=$1 AND is_active=true
		RETURNING ` + planCols
	row := r.pool.QueryRow(ctx, q,
		p.ID, p.Name, p.Description, p.DefaultRate,
		tiersJSON, string(p.ApplyBase), string(p.TriggerOn), stJSON, p.UpdatedBy,
	)
	plan, err := scanPlan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPlanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("plan.Update: %w", err)
	}
	return plan, nil
}

func (r *PlanRepo) Deactivate(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) (*domain.CommissionPlan, error) {
	const q = `UPDATE commission_plans SET is_active=false, updated_by=$2, updated_at=NOW() WHERE id=$1 RETURNING ` + planCols
	row := r.pool.QueryRow(ctx, q, id, updatedBy)
	plan, err := scanPlan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrPlanNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("plan.Deactivate: %w", err)
	}
	return plan, nil
}

func (r *PlanRepo) List(ctx context.Context, f domain.ListPlansFilter, page, size int) ([]*domain.CommissionPlan, int64, error) {
	where := "WHERE 1=1"
	args := []any{}
	idx := 1

	if f.IsActive != nil {
		where += fmt.Sprintf(" AND is_active=$%d", idx)
		args = append(args, *f.IsActive)
		idx++
	}
	if f.Type != "" {
		where += fmt.Sprintf(" AND type=$%d", idx)
		args = append(args, string(f.Type))
		idx++
	}

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM commission_plans `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("plan.List count: %w", err)
	}

	offset := (page - 1) * size
	args = append(args, size, offset)
	q := fmt.Sprintf(`SELECT `+planCols+` FROM commission_plans `+where+` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("plan.List: %w", err)
	}
	defer rows.Close()

	var list []*domain.CommissionPlan
	for rows.Next() {
		p, err := scanPlan(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("plan.List scan: %w", err)
		}
		list = append(list, p)
	}
	if list == nil {
		list = []*domain.CommissionPlan{}
	}
	return list, total, rows.Err()
}
