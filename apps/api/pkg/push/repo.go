package push

import (
	"context"

	"github.com/google/uuid"
)

// DeviceRepository manages push device persistence.
type DeviceRepository interface {
	Upsert(ctx context.Context, p RegisterDeviceParams) (*PushDevice, error)
	FindByToken(ctx context.Context, deviceToken string) (*PushDevice, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]PushDevice, error)
	Deactivate(ctx context.Context, userID uuid.UUID, deviceToken string) error
	UpdateLastActive(ctx context.Context, deviceToken string) error
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]PushDevice, error)
}
