// Package ws provides the WebSocket hub and client primitives for real-time
// event streaming via GET /api/v1/events/stream.
package ws

import (
	"encoding/json"
	"time"
)

// Message is the JSON envelope sent to every subscribed WebSocket client.
type Message struct {
	// Type describes the event kind (e.g. "notification", "crm.client.created").
	Type    string          `json:"type"`
	// Channel is the logical topic the event belongs to (e.g. "global", "crm").
	Channel string          `json:"channel"`
	// Data carries the event payload — arbitrary JSON.
	Data    json.RawMessage `json:"data,omitempty"`
	// Timestamp is when the event was emitted (UTC).
	Timestamp time.Time     `json:"timestamp"`
}

// broadcastReq is an internal request to fan a message out to all clients
// subscribed to Channel.
type broadcastReq struct {
	channel string
	payload []byte // pre-marshalled Message
}
