package notification_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/pkg/notification"
	"github.com/mdh/erp-audit/api/pkg/push"
)

// ── Fakes ─────────────────────────────────────────────────────────────────────

type fakeDeviceLister struct {
	devices []push.PushDevice
	err     error
}

func (f *fakeDeviceLister) ListActiveByUser(_ context.Context, _ uuid.UUID) ([]push.PushDevice, error) {
	return f.devices, f.err
}

type fakeSender struct {
	online map[string]bool // device_token → online
	sent   []string        // tokens that received a send call
}

func (f *fakeSender) Send(token string, _ push.PushPayload) bool {
	if f.online[token] {
		f.sent = append(f.sent, token)
		return true
	}
	return false
}

// ── Tests ──────────────────────────────────────────────────────────────────────

func TestNotifier_NotifyUser_Online(t *testing.T) {
	token := "tok-online"
	lister := &fakeDeviceLister{
		devices: []push.PushDevice{{DeviceToken: token, IsActive: true}},
	}
	sender := &fakeSender{online: map[string]bool{token: true}}
	n := notification.New(lister, sender)

	delivered := n.NotifyUser(context.Background(), uuid.New(), push.PushPayload{Title: "Test"})
	if delivered != 1 {
		t.Errorf("want 1 delivered, got %d", delivered)
	}
	if len(sender.sent) != 1 || sender.sent[0] != token {
		t.Errorf("expected token %q sent, got %v", token, sender.sent)
	}
}

func TestNotifier_NotifyUser_Offline(t *testing.T) {
	token := "tok-offline"
	lister := &fakeDeviceLister{
		devices: []push.PushDevice{{DeviceToken: token, IsActive: true}},
	}
	sender := &fakeSender{online: map[string]bool{}} // not online
	n := notification.New(lister, sender)

	delivered := n.NotifyUser(context.Background(), uuid.New(), push.PushPayload{Title: "Test"})
	if delivered != 0 {
		t.Errorf("want 0 delivered (offline), got %d", delivered)
	}
}

func TestNotifier_NotifyUser_DeviceRepoError(t *testing.T) {
	lister := &fakeDeviceLister{err: errors.New("db error")}
	sender := &fakeSender{online: map[string]bool{}}
	n := notification.New(lister, sender)

	delivered := n.NotifyUser(context.Background(), uuid.New(), push.PushPayload{Title: "Test"})
	if delivered != 0 {
		t.Errorf("want 0 on repo error, got %d", delivered)
	}
}

func TestNotifier_NotifyUser_NoDevices(t *testing.T) {
	lister := &fakeDeviceLister{devices: nil}
	sender := &fakeSender{online: map[string]bool{}}
	n := notification.New(lister, sender)

	delivered := n.NotifyUser(context.Background(), uuid.New(), push.PushPayload{Title: "Test"})
	if delivered != 0 {
		t.Errorf("want 0 for user with no devices, got %d", delivered)
	}
}

func TestNotifier_NotifyUsers_MultipleUsers(t *testing.T) {
	tok1, tok2 := "tok-user1", "tok-user2"
	user1, user2 := uuid.New(), uuid.New()
	calls := 0
	lister := &multiUserLister{
		lookup: map[uuid.UUID][]push.PushDevice{
			user1: {{DeviceToken: tok1, IsActive: true}},
			user2: {{DeviceToken: tok2, IsActive: true}},
		},
		calls: &calls,
	}
	sender := &fakeSender{online: map[string]bool{tok1: true, tok2: true}}
	n := notification.New(lister, sender)

	delivered := n.NotifyUsers(context.Background(), []uuid.UUID{user1, user2}, push.PushPayload{Title: "Team"})
	if delivered != 2 {
		t.Errorf("want 2 delivered, got %d", delivered)
	}
}

func TestNotifier_NotifyUsers_Empty(t *testing.T) {
	lister := &fakeDeviceLister{}
	sender := &fakeSender{online: map[string]bool{}}
	n := notification.New(lister, sender)

	delivered := n.NotifyUsers(context.Background(), nil, push.PushPayload{Title: "x"})
	if delivered != 0 {
		t.Errorf("want 0 for empty user list, got %d", delivered)
	}
}

// multiUserLister returns different devices per user.
type multiUserLister struct {
	lookup map[uuid.UUID][]push.PushDevice
	calls  *int
}

func (m *multiUserLister) ListActiveByUser(_ context.Context, userID uuid.UUID) ([]push.PushDevice, error) {
	*m.calls++
	return m.lookup[userID], nil
}
