// Package worker provides scheduled Asynq jobs for the reporting module.
package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/mdh/erp-audit/api/internal/reporting/usecase"
)

const TaskRefreshMVs = "reporting:refresh-views"

// NewMVRefreshHandler returns an Asynq handler for nightly MV refresh.
func NewMVRefreshHandler(uc *usecase.ReportUseCase) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		if err := uc.RefreshMaterializedViews(ctx); err != nil {
			return fmt.Errorf("reporting:refresh-views: %w", err)
		}
		return nil
	}
}
