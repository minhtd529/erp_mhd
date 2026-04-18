package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/mdh/erp-audit/api/pkg/notification"
	"github.com/mdh/erp-audit/api/pkg/push"
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

// NewTimesheetApprovedHandler returns an Asynq handler that notifies the staff
// member when their timesheet is approved.
func NewTimesheetApprovedHandler(n *notification.Notifier) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p TimesheetEventPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("TimesheetApproved: bad payload: %w", err)
		}
		staffID, err := uuid.Parse(p.StaffID)
		if err != nil {
			return fmt.Errorf("TimesheetApproved: invalid staff_id %q: %w", p.StaffID, err)
		}
		n.NotifyUser(ctx, staffID, push.PushPayload{
			Title:    "Timesheet Approved",
			Body:     "Your timesheet has been approved.",
			Priority: "high",
			TTL:      86400,
			Data:     map[string]string{"type": "TIMESHEET_APPROVED", "timesheet_id": p.TimesheetID},
		})
		return nil
	}
}

// NewTimesheetRejectedHandler returns an Asynq handler that notifies the staff
// member when their timesheet is rejected.
func NewTimesheetRejectedHandler(n *notification.Notifier) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p TimesheetEventPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("TimesheetRejected: bad payload: %w", err)
		}
		staffID, err := uuid.Parse(p.StaffID)
		if err != nil {
			return fmt.Errorf("TimesheetRejected: invalid staff_id %q: %w", p.StaffID, err)
		}
		n.NotifyUser(ctx, staffID, push.PushPayload{
			Title:    "Timesheet Rejected",
			Body:     "Your timesheet requires changes.",
			Priority: "high",
			TTL:      86400,
			Data:     map[string]string{"type": "TIMESHEET_REJECTED", "timesheet_id": p.TimesheetID},
		})
		return nil
	}
}

// NewTimesheetSubmittedHandler returns an Asynq handler for timesheet submission.
// It notifies the submitting staff member that the submission was received.
func NewTimesheetSubmittedHandler(n *notification.Notifier) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p TimesheetEventPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("TimesheetSubmitted: bad payload: %w", err)
		}
		staffID, err := uuid.Parse(p.StaffID)
		if err != nil {
			return fmt.Errorf("TimesheetSubmitted: invalid staff_id %q: %w", p.StaffID, err)
		}
		n.NotifyUser(ctx, staffID, push.PushPayload{
			Title:    "Timesheet Submitted",
			Body:     "Your timesheet has been submitted for review.",
			Priority: "normal",
			TTL:      86400,
			Data:     map[string]string{"type": "TIMESHEET_SUBMITTED", "timesheet_id": p.TimesheetID},
		})
		return nil
	}
}

// NewTimesheetLockedHandler returns an Asynq handler for timesheet locking.
func NewTimesheetLockedHandler(n *notification.Notifier) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p TimesheetEventPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("TimesheetLocked: bad payload: %w", err)
		}
		staffID, err := uuid.Parse(p.StaffID)
		if err != nil {
			return fmt.Errorf("TimesheetLocked: invalid staff_id %q: %w", p.StaffID, err)
		}
		n.NotifyUser(ctx, staffID, push.PushPayload{
			Title:    "Timesheet Locked",
			Body:     "Your timesheet has been locked for billing.",
			Priority: "normal",
			TTL:      86400,
			Data:     map[string]string{"type": "TIMESHEET_LOCKED", "timesheet_id": p.TimesheetID},
		})
		return nil
	}
}

// NewEngagementActivatedHandler returns an Asynq handler that notifies the
// engagement owner when an engagement becomes active.
func NewEngagementActivatedHandler(n *notification.Notifier) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var p EngagementEventPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("EngagementActivated: bad payload: %w", err)
		}
		callerID, err := uuid.Parse(p.CallerID)
		if err != nil {
			return fmt.Errorf("EngagementActivated: invalid caller_id %q: %w", p.CallerID, err)
		}
		n.NotifyUser(ctx, callerID, push.PushPayload{
			Title:    "Engagement Activated",
			Body:     "An engagement has been activated.",
			Priority: "high",
			TTL:      86400,
			Data:     map[string]string{"type": "ENGAGEMENT_ACTIVATED", "engagement_id": p.EngagementID},
		})
		return nil
	}
}

// HandleTimesheetApproved is the legacy stub (kept for backward compatibility).
func HandleTimesheetApproved(_ context.Context, t *asynq.Task) error {
	var p TimesheetEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("TimesheetApproved: bad payload: %w", err)
	}
	return nil
}

// HandleTimesheetSubmitted is the legacy stub.
func HandleTimesheetSubmitted(_ context.Context, t *asynq.Task) error {
	var p TimesheetEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("TimesheetSubmitted: bad payload: %w", err)
	}
	return nil
}

// HandleTimesheetLocked is the legacy stub.
func HandleTimesheetLocked(_ context.Context, t *asynq.Task) error {
	var p TimesheetEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("TimesheetLocked: bad payload: %w", err)
	}
	return nil
}

// HandleTimesheetRejected is the legacy stub.
func HandleTimesheetRejected(_ context.Context, t *asynq.Task) error {
	var p TimesheetEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("TimesheetRejected: bad payload: %w", err)
	}
	return nil
}

// HandleEngagementActivated is the legacy stub.
func HandleEngagementActivated(_ context.Context, t *asynq.Task) error {
	var p EngagementEventPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("EngagementActivated: bad payload: %w", err)
	}
	return nil
}
