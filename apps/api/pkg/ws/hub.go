package ws

import (
	"encoding/json"
	"sync"
	"time"
)

// Hub maintains the set of active WebSocket clients and fans broadcast
// messages out to all clients subscribed to a given channel.
//
// Call Run() in a dedicated goroutine once, then use Broadcast() from
// anywhere in the application to push real-time events.
type Hub struct {
	mu       sync.RWMutex
	// clients registered with this hub, keyed by client pointer
	clients  map[*Client]bool
	// channelClients maps channel name → set of clients subscribed to it
	channelClients map[string]map[*Client]bool

	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastReq
	done       chan struct{}
}

// NewHub creates an initialised Hub ready to Run.
func NewHub() *Hub {
	return &Hub{
		clients:        make(map[*Client]bool),
		channelClients: make(map[string]map[*Client]bool),
		register:       make(chan *Client, 64),
		unregister:     make(chan *Client, 64),
		broadcast:      make(chan broadcastReq, 256),
		done:           make(chan struct{}),
	}
}

// Run processes register / unregister / broadcast events.  It blocks until
// Stop is called, so run it in its own goroutine.
func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			for _, ch := range client.channels {
				if h.channelClients[ch] == nil {
					h.channelClients[ch] = make(map[*Client]bool)
				}
				h.channelClients[ch][client] = true
			}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if h.clients[client] {
				delete(h.clients, client)
				for _, ch := range client.channels {
					if subs, ok := h.channelClients[ch]; ok {
						delete(subs, client)
						if len(subs) == 0 {
							delete(h.channelClients, ch)
						}
					}
				}
				close(client.send)
			}
			h.mu.Unlock()

		case req := <-h.broadcast:
			h.mu.RLock()
			subs := h.channelClients[req.channel]
			for client := range subs {
				select {
				case client.send <- req.payload:
				default:
					// slow client — drop the message to avoid blocking the hub
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Stop shuts the hub down gracefully.
func (h *Hub) Stop() {
	close(h.done)
}

// Broadcast sends msg to all clients subscribed to msg.Channel.
// Safe to call from any goroutine.
func (h *Hub) Broadcast(channel string, eventType string, data any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msg := Message{
		Type:      eventType,
		Channel:   channel,
		Data:      json.RawMessage(raw),
		Timestamp: time.Now().UTC(),
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	select {
	case h.broadcast <- broadcastReq{channel: channel, payload: payload}:
	default:
		// broadcast channel full — drop
	}
	return nil
}

// ClientCount returns the number of currently connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// SubscriberCount returns the number of clients subscribed to a specific channel.
func (h *Hub) SubscriberCount(channel string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.channelClients[channel])
}
