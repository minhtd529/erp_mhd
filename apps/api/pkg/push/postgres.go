package push

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrDeviceNotFound is returned when a device token does not exist.
var ErrDeviceNotFound = errors.New("DEVICE_NOT_FOUND")

type postgresRepo struct{ pool *pgxpool.Pool }

// NewDeviceRepo returns a PostgreSQL-backed DeviceRepository.
func NewDeviceRepo(pool *pgxpool.Pool) DeviceRepository {
	return &postgresRepo{pool: pool}
}

func (r *postgresRepo) Upsert(ctx context.Context, p RegisterDeviceParams) (*PushDevice, error) {
	const q = `
		INSERT INTO push_devices (user_id, device_token, platform, device_name, app_version, os_version)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (device_token) DO UPDATE
		    SET user_id       = EXCLUDED.user_id,
		        platform      = EXCLUDED.platform,
		        device_name   = EXCLUDED.device_name,
		        app_version   = EXCLUDED.app_version,
		        os_version    = EXCLUDED.os_version,
		        is_active     = TRUE,
		        last_active_at = NOW()
		RETURNING id, user_id, platform, device_name, app_version, os_version, is_active, last_active_at, created_at`

	var d PushDevice
	err := r.pool.QueryRow(ctx, q,
		p.UserID, p.DeviceToken, p.Platform, p.DeviceName, p.AppVersion, p.OSVersion,
	).Scan(
		&d.ID, &d.UserID, &d.Platform, &d.DeviceName, &d.AppVersion, &d.OSVersion,
		&d.IsActive, &d.LastActiveAt, &d.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("push.Upsert: %w", err)
	}
	d.DeviceToken = p.DeviceToken
	return &d, nil
}

func (r *postgresRepo) FindByToken(ctx context.Context, deviceToken string) (*PushDevice, error) {
	const q = `
		SELECT id, user_id, device_token, platform, device_name, app_version, os_version, is_active, last_active_at, created_at
		FROM push_devices WHERE device_token = $1`
	var d PushDevice
	if err := r.pool.QueryRow(ctx, q, deviceToken).Scan(
		&d.ID, &d.UserID, &d.DeviceToken, &d.Platform, &d.DeviceName, &d.AppVersion,
		&d.OSVersion, &d.IsActive, &d.LastActiveAt, &d.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeviceNotFound
		}
		return nil, fmt.Errorf("push.FindByToken: %w", err)
	}
	return &d, nil
}

func (r *postgresRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]PushDevice, error) {
	return r.list(ctx, userID, false)
}

func (r *postgresRepo) ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]PushDevice, error) {
	return r.list(ctx, userID, true)
}

func (r *postgresRepo) list(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]PushDevice, error) {
	q := `
		SELECT id, user_id, device_token, platform, device_name, app_version, os_version, is_active, last_active_at, created_at
		FROM push_devices WHERE user_id = $1`
	if activeOnly {
		q += ` AND is_active = TRUE`
	}
	q += ` ORDER BY last_active_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("push.list: %w", err)
	}
	defer rows.Close()
	var list []PushDevice
	for rows.Next() {
		var d PushDevice
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.DeviceToken, &d.Platform, &d.DeviceName, &d.AppVersion,
			&d.OSVersion, &d.IsActive, &d.LastActiveAt, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func (r *postgresRepo) Deactivate(ctx context.Context, userID uuid.UUID, deviceToken string) error {
	const q = `UPDATE push_devices SET is_active = FALSE WHERE user_id = $1 AND device_token = $2`
	if _, err := r.pool.Exec(ctx, q, userID, deviceToken); err != nil {
		return fmt.Errorf("push.Deactivate: %w", err)
	}
	return nil
}

func (r *postgresRepo) UpdateLastActive(ctx context.Context, deviceToken string) error {
	const q = `UPDATE push_devices SET last_active_at = NOW() WHERE device_token = $1 AND is_active = TRUE`
	if _, err := r.pool.Exec(ctx, q, deviceToken); err != nil {
		return fmt.Errorf("push.UpdateLastActive: %w", err)
	}
	return nil
}
