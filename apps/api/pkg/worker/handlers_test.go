package worker_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/mdh/erp-audit/api/pkg/worker"
)

func makeTask(t *testing.T, payload worker.TimesheetEventPayload) *asynq.Task {
	t.Helper()
	raw, _ := json.Marshal(payload)
	return asynq.NewTask("test", raw)
}

func TestHandleTimesheetApproved_ValidPayload(t *testing.T) {
	t.Parallel()
	task := makeTask(t, worker.TimesheetEventPayload{
		TimesheetID: "ts-1", StaffID: "staff-1", Status: "APPROVED", CallerID: "mgr-1",
	})
	if err := worker.HandleTimesheetApproved(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleTimesheetApproved_BadPayload(t *testing.T) {
	t.Parallel()
	task := asynq.NewTask("test", []byte("not-json"))
	if err := worker.HandleTimesheetApproved(context.Background(), task); err == nil {
		t.Fatal("expected error for bad payload")
	}
}

func TestHandleTimesheetSubmitted_ValidPayload(t *testing.T) {
	t.Parallel()
	task := makeTask(t, worker.TimesheetEventPayload{
		TimesheetID: "ts-2", StaffID: "staff-2", Status: "SUBMITTED", CallerID: "staff-2",
	})
	if err := worker.HandleTimesheetSubmitted(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleTimesheetLocked_ValidPayload(t *testing.T) {
	t.Parallel()
	task := makeTask(t, worker.TimesheetEventPayload{
		TimesheetID: "ts-3", StaffID: "staff-3", Status: "LOCKED", CallerID: "mgr-1",
	})
	if err := worker.HandleTimesheetLocked(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleTimesheetRejected_ValidPayload(t *testing.T) {
	t.Parallel()
	task := makeTask(t, worker.TimesheetEventPayload{
		TimesheetID: "ts-4", StaffID: "staff-4", Status: "REJECTED", CallerID: "mgr-1",
	})
	if err := worker.HandleTimesheetRejected(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
