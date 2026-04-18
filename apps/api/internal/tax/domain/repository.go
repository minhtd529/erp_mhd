package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type DeadlineRepository interface {
	Create(ctx context.Context, p CreateDeadlineParams) (*TaxDeadline, error)
	FindByID(ctx context.Context, id uuid.UUID) (*TaxDeadline, error)
	List(ctx context.Context, f ListDeadlinesFilter, page, size int) ([]*TaxDeadline, int64, error)
	Update(ctx context.Context, p UpdateDeadlineParams) (*TaxDeadline, error)
	MarkCompleted(ctx context.Context, id uuid.UUID, actualDate time.Time, updatedBy uuid.UUID) (*TaxDeadline, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status DeadlineStatus) error
	ListDueSoon(ctx context.Context, beforeDate time.Time) ([]*TaxDeadline, error)
	ListOverdue(ctx context.Context) ([]*TaxDeadline, error)
}

type AdvisoryRepository interface {
	Create(ctx context.Context, p CreateAdvisoryParams) (*AdvisoryRecord, error)
	FindByID(ctx context.Context, id uuid.UUID) (*AdvisoryRecord, error)
	List(ctx context.Context, f ListAdvisoryFilter, page, size int) ([]*AdvisoryRecord, int64, error)
	Update(ctx context.Context, p UpdateAdvisoryParams) (*AdvisoryRecord, error)
	Deliver(ctx context.Context, id uuid.UUID, updatedBy uuid.UUID) (*AdvisoryRecord, error)
	AttachFile(ctx context.Context, p AttachFileParams) (*AdvisoryFile, error)
	ListFiles(ctx context.Context, advisoryID uuid.UUID) ([]*AdvisoryFile, error)
}

type ComplianceRepository interface {
	GetComplianceStatus(ctx context.Context, clientID uuid.UUID) (*ComplianceStatus, error)
	ListAllOverdue(ctx context.Context) ([]*TaxDeadline, error)
	DashboardDeadlines(ctx context.Context, from, to time.Time) ([]*TaxDeadline, error)
}
