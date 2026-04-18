package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/pkg/push"
)

// PushDeviceUseCase manages push device registration and heartbeat.
type PushDeviceUseCase struct {
	repo push.DeviceRepository
}

func NewPushDeviceUseCase(repo push.DeviceRepository) *PushDeviceUseCase {
	return &PushDeviceUseCase{repo: repo}
}

func (uc *PushDeviceUseCase) RegisterDevice(ctx context.Context, p push.RegisterDeviceParams) (*push.PushDevice, error) {
	return uc.repo.Upsert(ctx, p)
}

func (uc *PushDeviceUseCase) UnregisterDevice(ctx context.Context, userID uuid.UUID, deviceToken string) error {
	return uc.repo.Deactivate(ctx, userID, deviceToken)
}

func (uc *PushDeviceUseCase) ListDevices(ctx context.Context, userID uuid.UUID) ([]push.PushDevice, error) {
	return uc.repo.ListActiveByUser(ctx, userID)
}

func (uc *PushDeviceUseCase) Heartbeat(ctx context.Context, deviceToken string) error {
	return uc.repo.UpdateLastActive(ctx, deviceToken)
}
