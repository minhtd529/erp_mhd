package worker_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/mdh/erp-audit/api/pkg/notification"
	"github.com/mdh/erp-audit/api/pkg/push"
	"github.com/mdh/erp-audit/api/pkg/worker"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeDeviceLister struct {
	devices []push.PushDevice
}

func (f *fakeDeviceLister) ListActiveByUser(_ context.Context, _ uuid.UUID) ([]push.PushDevice, error) {
	return f.devices, nil
}

type fakeSender struct {
	online  map[string]bool
	sent    int
}

func (f *fakeSender) Send(token string, _ push.PushPayload) bool {
	if f.online[token] {
		f.sent++
		return true
	}
	return false
}

func buildNotifier(deviceToken string, online bool) *notification.Notifier {
	lister := &fakeDeviceLister{
		devices: []push.PushDevice{{DeviceToken: deviceToken, IsActive: true}},
	}
	sender := &fakeSender{online: map[string]bool{deviceToken: online}}
	return notification.New(lister, sender)
}

func makeEngTask(t *testing.T, payload worker.EngagementEventPayload) *asynq.Task {
	t.Helper()
	raw, _ := json.Marshal(payload)
	return asynq.NewTask("test", raw)
}

// ── NewTimesheetApprovedHandler ───────────────────────────────────────────────

func TestNewTimesheetApprovedHandler_ValidPayload(t *testing.T) {
	staffID := uuid.New()
	n := buildNotifier("tok-approved", true)
	h := worker.NewTimesheetApprovedHandler(n)

	p := worker.TimesheetEventPayload{
		TimesheetID: uuid.New().String(),
		StaffID:     staffID.String(),
		Status:      "APPROVED",
		CallerID:    uuid.New().String(),
	}
	raw, _ := json.Marshal(p)
	task := asynq.NewTask("test", raw)

	if err := h(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewTimesheetApprovedHandler_BadPayload(t *testing.T) {
	n := buildNotifier("tok", false)
	h := worker.NewTimesheetApprovedHandler(n)
	task := asynq.NewTask("test", []byte("not-json"))
	if err := h(context.Background(), task); err == nil {
		t.Fatal("expected error for bad payload")
	}
}

func TestNewTimesheetApprovedHandler_InvalidStaffID(t *testing.T) {
	n := buildNotifier("tok", false)
	h := worker.NewTimesheetApprovedHandler(n)

	p := worker.TimesheetEventPayload{TimesheetID: "ts-1", StaffID: "not-a-uuid", Status: "APPROVED"}
	raw, _ := json.Marshal(p)
	task := asynq.NewTask("test", raw)
	if err := h(context.Background(), task); err == nil {
		t.Fatal("expected error for invalid staff_id")
	}
}

// ── NewTimesheetRejectedHandler ───────────────────────────────────────────────

func TestNewTimesheetRejectedHandler_ValidPayload(t *testing.T) {
	staffID := uuid.New()
	n := buildNotifier("tok-rejected", true)
	h := worker.NewTimesheetRejectedHandler(n)

	p := worker.TimesheetEventPayload{
		TimesheetID: uuid.New().String(),
		StaffID:     staffID.String(),
		Status:      "REJECTED",
		CallerID:    uuid.New().String(),
	}
	raw, _ := json.Marshal(p)
	task := asynq.NewTask("test", raw)
	if err := h(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewTimesheetRejectedHandler_BadPayload(t *testing.T) {
	n := buildNotifier("tok", false)
	h := worker.NewTimesheetRejectedHandler(n)
	task := asynq.NewTask("test", []byte("bad"))
	if err := h(context.Background(), task); err == nil {
		t.Fatal("expected error for bad payload")
	}
}

// ── NewTimesheetSubmittedHandler ──────────────────────────────────────────────

func TestNewTimesheetSubmittedHandler_ValidPayload(t *testing.T) {
	staffID := uuid.New()
	n := buildNotifier("tok-submitted", true)
	h := worker.NewTimesheetSubmittedHandler(n)

	p := worker.TimesheetEventPayload{
		TimesheetID: uuid.New().String(),
		StaffID:     staffID.String(),
		Status:      "SUBMITTED",
		CallerID:    staffID.String(),
	}
	raw, _ := json.Marshal(p)
	task := asynq.NewTask("test", raw)
	if err := h(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── NewEngagementActivatedHandler ────────────────────────────────────────────

func TestNewEngagementActivatedHandler_ValidPayload(t *testing.T) {
	callerID := uuid.New()
	n := buildNotifier("tok-eng", true)
	h := worker.NewEngagementActivatedHandler(n)

	task := makeEngTask(t, worker.EngagementEventPayload{
		EngagementID: uuid.New().String(),
		Status:       "ACTIVE",
		CallerID:     callerID.String(),
	})
	if err := h(context.Background(), task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewEngagementActivatedHandler_BadPayload(t *testing.T) {
	n := buildNotifier("tok", false)
	h := worker.NewEngagementActivatedHandler(n)
	task := asynq.NewTask("test", []byte("bad"))
	if err := h(context.Background(), task); err == nil {
		t.Fatal("expected error for bad payload")
	}
}

func TestNewEngagementActivatedHandler_InvalidCallerID(t *testing.T) {
	n := buildNotifier("tok", false)
	h := worker.NewEngagementActivatedHandler(n)
	task := makeEngTask(t, worker.EngagementEventPayload{EngagementID: uuid.New().String(), CallerID: "bad-uuid"})
	if err := h(context.Background(), task); err == nil {
		t.Fatal("expected error for invalid caller_id")
	}
}
