package ws

// Broadcaster is the interface use cases depend on for real-time fan-out.
// *Hub satisfies it; a nil *Hub is safe (no-op) for unit tests.
type Broadcaster interface {
	Broadcast(channel string, eventType string, data any) error
}

// NopBroadcaster is a no-op implementation for use in tests.
type NopBroadcaster struct{}

func (NopBroadcaster) Broadcast(_ string, _ string, _ any) error { return nil }
