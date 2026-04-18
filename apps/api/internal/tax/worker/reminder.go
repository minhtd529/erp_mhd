// Package worker provides scheduled Asynq jobs for the tax advisory module.
package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/mdh/erp-audit/api/internal/tax/usecase"
)

const TaskDeadlineReminder = "tax:deadline-reminder"

// NewDeadlineReminderHandler returns an Asynq handler that refreshes deadline statuses.
// Scheduled daily — marks past-due deadlines as OVERDUE and near-due as DUE_SOON.
func NewDeadlineReminderHandler(deadlineUC *usecase.TaxDeadlineUseCase) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		if err := deadlineUC.RefreshOverdue(ctx); err != nil {
			return fmt.Errorf("deadline-reminder: refresh overdue: %w", err)
		}
		if err := deadlineUC.RefreshDueSoon(ctx); err != nil {
			return fmt.Errorf("deadline-reminder: refresh due-soon: %w", err)
		}
		return nil
	}
}
