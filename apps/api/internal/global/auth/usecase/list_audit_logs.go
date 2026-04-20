package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// ─── Query-side types ────────────────────────────────────────────────────────

// AuditLogListRequest holds validated query parameters for GET /api/v1/audit-logs.
type AuditLogListRequest struct {
	Page     int        `form:"page,default=1"  binding:"min=1"`
	Size     int        `form:"size,default=50" binding:"min=1,max=200"`
	Module   string     `form:"module"`
	Resource string     `form:"resource"`
	Action   string     `form:"action"`
	UserID   *uuid.UUID `form:"user_id"`
	From     *time.Time `form:"from" time_format:"2006-01-02"`
	To       *time.Time `form:"to"   time_format:"2006-01-02"`
}

// AuditLogFilter is the internal filter struct passed to the repository.
type AuditLogFilter struct {
	Page     int
	Size     int
	Module   string
	Resource string
	Action   string
	UserID   *uuid.UUID
	From     *time.Time
	To       *time.Time
}

// AuditLogEntry is one row in the audit log list response.
type AuditLogEntry struct {
	ID         uuid.UUID       `json:"id"`
	UserID     *uuid.UUID      `json:"user_id"`
	UserName   string          `json:"user_name"`
	Module     string          `json:"module"`
	Resource   string          `json:"resource"`
	ResourceID *uuid.UUID      `json:"resource_id"`
	Action     string          `json:"action"`
	IPAddress  string          `json:"ip_address"`
	CreatedAt  time.Time       `json:"created_at"`
}

// AuditLogQuerier is the query-side read interface for audit logs.
// Implemented by repository.AuditLogRepo.
type AuditLogQuerier interface {
	ListAuditLogs(ctx context.Context, f AuditLogFilter) ([]AuditLogEntry, int64, error)
}

// ─── UseCase ──────────────────────────────────────────────────────────────────

// ListAuditLogsUseCase lists audit log entries with optional filters.
type ListAuditLogsUseCase struct {
	repo AuditLogQuerier
}

func NewListAuditLogsUseCase(repo AuditLogQuerier) *ListAuditLogsUseCase {
	return &ListAuditLogsUseCase{repo: repo}
}

func (uc *ListAuditLogsUseCase) Execute(ctx context.Context, req AuditLogListRequest) (PaginatedResult[AuditLogEntry], error) {
	entries, total, err := uc.repo.ListAuditLogs(ctx, AuditLogFilter{
		Page:     req.Page,
		Size:     req.Size,
		Module:   req.Module,
		Resource: req.Resource,
		Action:   req.Action,
		UserID:   req.UserID,
		From:     req.From,
		To:       req.To,
	})
	if err != nil {
		return PaginatedResult[AuditLogEntry]{}, fmt.Errorf("audit.List: %w", err)
	}
	return pagination.NewOffsetResult(entries, total, req.Page, req.Size), nil
}
