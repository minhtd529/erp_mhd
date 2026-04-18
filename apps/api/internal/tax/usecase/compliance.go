package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/tax/domain"
)

type ComplianceUseCase struct {
	repo domain.ComplianceRepository
}

func NewComplianceUseCase(repo domain.ComplianceRepository) *ComplianceUseCase {
	return &ComplianceUseCase{repo: repo}
}

func (uc *ComplianceUseCase) GetClientComplianceStatus(ctx context.Context, clientID uuid.UUID) (*domain.ComplianceStatus, error) {
	return uc.repo.GetComplianceStatus(ctx, clientID)
}

func (uc *ComplianceUseCase) ListOverdueAlerts(ctx context.Context) ([]*domain.TaxDeadline, error) {
	return uc.repo.ListAllOverdue(ctx)
}

// DashboardDeadlines returns all deadlines in the next 90 days by default.
func (uc *ComplianceUseCase) DashboardDeadlines(ctx context.Context, from, to *time.Time) ([]*domain.TaxDeadline, error) {
	now := time.Now().UTC()
	start := now
	end := now.AddDate(0, 3, 0) // 90 days
	if from != nil {
		start = *from
	}
	if to != nil {
		end = *to
	}
	return uc.repo.DashboardDeadlines(ctx, start, end)
}
