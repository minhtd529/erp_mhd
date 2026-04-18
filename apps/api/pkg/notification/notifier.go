// Package notification delivers push payloads to registered user devices.
package notification

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/pkg/push"
)

// Sender abstracts the push relay so the Notifier can be unit-tested without real WebSocket connections.
type Sender interface {
	Send(deviceToken string, payload push.PushPayload) bool
}

// DeviceLister abstracts the device repository query used by the Notifier.
type DeviceLister interface {
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]push.PushDevice, error)
}

// Notifier delivers push payloads to all active devices for one or more users.
// Devices that are offline are silently skipped.
type Notifier struct {
	devices DeviceLister
	sender  Sender
}

// New creates a Notifier backed by a device repository and a relay sender.
func New(devices DeviceLister, sender Sender) *Notifier {
	return &Notifier{devices: devices, sender: sender}
}

// NotifyUser sends payload to all active devices for userID.
// Returns the number of online devices reached.
func (n *Notifier) NotifyUser(ctx context.Context, userID uuid.UUID, payload push.PushPayload) int {
	devices, err := n.devices.ListActiveByUser(ctx, userID)
	if err != nil {
		return 0
	}
	delivered := 0
	for _, d := range devices {
		if n.sender.Send(d.DeviceToken, payload) {
			delivered++
		}
	}
	return delivered
}

// NotifyUsers sends payload to all active devices for each user in userIDs.
// Returns the total number of online devices reached across all users.
func (n *Notifier) NotifyUsers(ctx context.Context, userIDs []uuid.UUID, payload push.PushPayload) int {
	total := 0
	for _, id := range userIDs {
		total += n.NotifyUser(ctx, id, payload)
	}
	return total
}
