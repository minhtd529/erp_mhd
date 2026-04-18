package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/tax/domain"
)

type ComplianceRepo struct{ pool *pgxpool.Pool }

func NewComplianceRepo(pool *pgxpool.Pool) *ComplianceRepo { return &ComplianceRepo{pool: pool} }

func (r *ComplianceRepo) GetComplianceStatus(ctx context.Context, clientID uuid.UUID) (*domain.ComplianceStatus, error) {
	const q = `SELECT client_id, total_deadlines, completed, overdue, due_soon, compliance_score
		FROM mv_tax_compliance_status WHERE client_id=$1`
	var s domain.ComplianceStatus
	err := r.pool.QueryRow(ctx, q, clientID).Scan(
		&s.ClientID, &s.TotalDeadlines, &s.Completed, &s.Overdue, &s.DueSoon, &s.ComplianceScore,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// Client exists but has no deadlines — return zero score
		return &domain.ComplianceStatus{ClientID: clientID, ComplianceScore: 100}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("compliance.GetComplianceStatus: %w", err)
	}
	return &s, nil
}

func (r *ComplianceRepo) ListAllOverdue(ctx context.Context) ([]*domain.TaxDeadline, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+deadlineCols+` FROM tax_deadlines WHERE status='OVERDUE' ORDER BY due_date ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("compliance.ListAllOverdue: %w", err)
	}
	defer rows.Close()
	return collectDeadlines(rows)
}

func (r *ComplianceRepo) DashboardDeadlines(ctx context.Context, from, to time.Time) ([]*domain.TaxDeadline, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+deadlineCols+` FROM tax_deadlines WHERE due_date >= $1 AND due_date <= $2 ORDER BY due_date ASC`,
		from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("compliance.DashboardDeadlines: %w", err)
	}
	defer rows.Close()
	return collectDeadlines(rows)
}
