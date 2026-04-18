// Package push provides self-hosted push notification relay for mobile devices.
package push

import (
	"time"

	"github.com/google/uuid"
)

type DevicePlatform string

const (
	PlatformIOS     DevicePlatform = "ios"
	PlatformAndroid DevicePlatform = "android"
	PlatformWebPush DevicePlatform = "web_push"
)

// PushDevice is a registered mobile / browser device for receiving push messages.
type PushDevice struct {
	ID           uuid.UUID      `json:"id"`
	UserID       uuid.UUID      `json:"user_id"`
	DeviceToken  string         `json:"-"`
	Platform     DevicePlatform `json:"platform"`
	DeviceName   string         `json:"device_name"`
	AppVersion   string         `json:"app_version"`
	OSVersion    string         `json:"os_version"`
	IsActive     bool           `json:"is_active"`
	LastActiveAt time.Time      `json:"last_active_at"`
	CreatedAt    time.Time      `json:"created_at"`
}

// PushPayload is the message sent to a device.
type PushPayload struct {
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	Data        map[string]string `json:"data,omitempty"`
	Priority    string            `json:"priority"`  // "high" | "normal"
	TTL         int               `json:"ttl"`       // seconds
	CollapseKey string            `json:"collapse_key,omitempty"`
}

// RegisterDeviceParams holds the fields needed to register a device.
type RegisterDeviceParams struct {
	UserID      uuid.UUID
	DeviceToken string
	Platform    DevicePlatform
	DeviceName  string
	AppVersion  string
	OSVersion   string
}
