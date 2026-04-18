package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// TimesheetEventPayload is the payload for all timesheet domain events.
type TimesheetEventPayload struct {
	TimesheetID string `json:"timesheet_id"`
	StaffID     string `json:"staff_id"`
	Status      string `json:"status"`
	CallerID    string `json:"caller_id"`
}

// EngagementEventPayload is the payload for engagement domain events.
type EngagementEventPayload struct {
	EngagementID string `json:"engagement_id"`
	Status       string `json:"status"`
	CallerID     string `json:"caller_id"`
}

// HandleTimesheetApproved is called when a timesheet is approved.
// Phase 3 will replace this stub with billing intake logic.
func HandleTimesheetApproved(_ context.Context, t *asynq.Task) error {
	var p TimesheetEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("TimesheetApproved: bad payload: %w", err)
	}
	// TODO Phase 3: trigger billing intake for timesheet p.TimesheetID
	return nil
}

// HandleTimesheetSubmitted is called when a staff member submits a timesheet.
// Phase 3 will add manager notification logic here.
func HandleTimesheetSubmitted(_ context.Context, t *asynq.Task) error {
	var p TimesheetEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("TimesheetSubmitted: bad payload: %w", err)
	}
	// TODO Phase 3: notify manager for timesheet p.TimesheetID
	return nil
}

// HandleTimesheetLocked is called when a timesheet is locked for billing.
func HandleTimesheetLocked(_ context.Context, t *asynq.Task) error {
	var p TimesheetEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("TimesheetLocked: bad payload: %w", err)
	}
	// TODO Phase 3: create billing line items for timesheet p.TimesheetID
	return nil
}

// HandleTimesheetRejected is called when a timesheet is rejected.
func HandleTimesheetRejected(_ context.Context, t *asynq.Task) error {
	var p TimesheetEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("TimesheetRejected: bad payload: %w", err)
	}
	// TODO Phase 3: notify staff of rejection for timesheet p.TimesheetID
	return nil
}

// HandleEngagementActivated handles engagement state transitions.
func HandleEngagementActivated(_ context.Context, t *asynq.Task) error {
	var p EngagementEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("EngagementActivated: bad payload: %w", err)
	}
	// TODO Phase 3: notify team members for engagement p.EngagementID
	return nil
}
