package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
)

// AuditLogRepo implements usecase.AuditLogQuerier using pgxpool.
type AuditLogRepo struct {
	pool *pgxpool.Pool
}

// NewAuditLogRepo constructs an AuditLogRepo.
func NewAuditLogRepo(pool *pgxpool.Pool) *AuditLogRepo {
	return &AuditLogRepo{pool: pool}
}

// ListAuditLogs returns a paginated, filtered list of audit log entries.
// Joins with users to include full_name alongside user_id.
func (r *AuditLogRepo) ListAuditLogs(ctx context.Context, f usecase.AuditLogFilter) ([]usecase.AuditLogEntry, int64, error) {
	offset := (f.Page - 1) * f.Size
	args := []any{}
	where := "WHERE 1=1"
	idx := 1

	if f.Module != "" {
		where += fmt.Sprintf(" AND a.module = $%d", idx)
		args = append(args, f.Module)
		idx++
	}
	if f.Resource != "" {
		where += fmt.Sprintf(" AND a.resource = $%d", idx)
		args = append(args, f.Resource)
		idx++
	}
	if f.Action != "" {
		where += fmt.Sprintf(" AND a.action = $%d", idx)
		args = append(args, f.Action)
		idx++
	}
	if f.UserID != nil {
		where += fmt.Sprintf(" AND a.user_id = $%d", idx)
		args = append(args, f.UserID)
		idx++
	}
	if f.From != nil {
		where += fmt.Sprintf(" AND a.created_at >= $%d", idx)
		args = append(args, f.From)
		idx++
	}
	if f.To != nil {
		where += fmt.Sprintf(" AND a.created_at < $%d", idx)
		args = append(args, f.To)
		idx++
	}

	countQ := `SELECT COUNT(*) FROM audit_logs a ` + where
	var total int64
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("audit.ListAuditLogs count: %w", err)
	}

	args = append(args, f.Size, offset)
	dataQ := fmt.Sprintf(`
		SELECT a.id, a.user_id, COALESCE(u.full_name, '') AS user_name,
		       a.module, a.resource, a.resource_id, a.action,
		       COALESCE(a.ip_address, '') AS ip_address, a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.user_id
		%s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d`, where, idx, idx+1)

	rows, err := r.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("audit.ListAuditLogs query: %w", err)
	}
	defer rows.Close()

	var entries []usecase.AuditLogEntry
	for rows.Next() {
		var e usecase.AuditLogEntry
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.UserName,
			&e.Module, &e.Resource, &e.ResourceID, &e.Action,
			&e.IPAddress, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("audit.ListAuditLogs scan: %w", err)
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []usecase.AuditLogEntry{}
	}
	return entries, total, rows.Err()
}
