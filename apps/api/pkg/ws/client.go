package ws

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// pongWait is the time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// pingPeriod — send pings to peer at this period (must be less than pongWait).
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize is the maximum message size allowed from peer.
	maxMessageSize = 512
	// sessionTimeout is the maximum duration of a single WebSocket session.
	sessionTimeout = 30 * time.Minute
	// sendBufSize is the buffered channel depth for outbound messages per client.
	sendBufSize = 64
)

// Client is a single WebSocket connection. It bridges the WebSocket connection
// to the Hub via the send channel.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	userID   uuid.UUID
	channels []string  // subscribed channels, immutable after registration
	send     chan []byte
}

// NewClient creates a Client and registers it with hub.
// The caller is responsible for starting the read and write pumps.
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, channels []string) *Client {
	c := &Client{
		hub:      hub,
		conn:     conn,
		userID:   userID,
		channels: channels,
		send:     make(chan []byte, sendBufSize),
	}
	hub.register <- c
	return c
}

// UserID returns the authenticated user's UUID.
func (c *Client) UserID() uuid.UUID { return c.userID }

// WritePump pumps messages from the hub to the WebSocket connection.
// It also enforces the 30-minute session timeout and handles pings.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	deadline := time.Now().Add(sessionTimeout)
	defer func() {
		ticker.Stop()
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if !ok {
				// Hub closed the channel — send a close message.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{}) //nolint:errcheck
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if time.Now().After(deadline) {
				// Session expired — close gracefully.
				c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
				c.conn.WriteMessage(websocket.CloseMessage,        //nolint:errcheck
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session timeout"))
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump reads incoming messages (we only expect pong frames) and refreshes
// the read deadline. When the read fails the client is removed from the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
		return nil
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
