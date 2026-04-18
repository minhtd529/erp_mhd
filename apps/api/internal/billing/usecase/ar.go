package usecase

import (
	"context"

	"github.com/mdh/erp-audit/api/internal/billing/domain"
)

// ARUseCase handles accounts receivable analytics.
type ARUseCase struct {
	arRepo domain.ARRepository
}

// NewARUseCase constructs an ARUseCase.
func NewARUseCase(arRepo domain.ARRepository) *ARUseCase {
	return &ARUseCase{arRepo: arRepo}
}

// GetAging returns AR aging breakdown grouped by client.
func (uc *ARUseCase) GetAging(ctx context.Context) ([]ARAgingResponse, error) {
	rows, err := uc.arRepo.GetAging(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ARAgingResponse, len(rows))
	for i, r := range rows {
		result[i] = toARAgingResponse(r)
	}
	return result, nil
}

// GetOutstanding returns outstanding balance totals grouped by client.
func (uc *ARUseCase) GetOutstanding(ctx context.Context) ([]AROutstandingResponse, error) {
	rows, err := uc.arRepo.GetOutstanding(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]AROutstandingResponse, len(rows))
	for i, r := range rows {
		result[i] = toAROutstandingResponse(r)
	}
	return result, nil
}
