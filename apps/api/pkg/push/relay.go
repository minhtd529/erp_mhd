package push

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
	maxMsgSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:    func(r *http.Request) bool { return true },
}

// Relay maintains persistent WebSocket connections from mobile devices.
// Keyed by device_token, it delivers push payloads to online devices.
type Relay struct {
	mu          sync.RWMutex
	connections map[string]*deviceConn // device_token → conn
}

type deviceConn struct {
	conn  *websocket.Conn
	send  chan []byte
	token string
}

// NewRelay creates a new push relay.
func NewRelay() *Relay {
	return &Relay{connections: make(map[string]*deviceConn)}
}

// ServeDevice upgrades an HTTP request to WebSocket and registers the device.
// The caller must authenticate the device token before calling this.
func (r *Relay) ServeDevice(w http.ResponseWriter, req *http.Request, deviceToken string) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	dc := &deviceConn{conn: conn, send: make(chan []byte, 16), token: deviceToken}
	r.mu.Lock()
	r.connections[deviceToken] = dc
	r.mu.Unlock()

	go dc.writePump()
	dc.readPump(func() {
		r.mu.Lock()
		delete(r.connections, deviceToken)
		r.mu.Unlock()
		close(dc.send)
	})
}

// Send delivers a payload to a device if it is currently connected.
// Returns false when the device is offline (message should be queued).
func (r *Relay) Send(deviceToken string, payload PushPayload) bool {
	r.mu.RLock()
	dc, ok := r.connections[deviceToken]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	select {
	case dc.send <- raw:
		return true
	default:
		return false
	}
}

// IsOnline returns true if the device has an active WebSocket connection.
func (r *Relay) IsOnline(deviceToken string) bool {
	r.mu.RLock()
	_, ok := r.connections[deviceToken]
	r.mu.RUnlock()
	return ok
}

// OnlineCount returns the number of currently connected devices.
func (r *Relay) OnlineCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.connections)
}

func (dc *deviceConn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() { ticker.Stop(); dc.conn.Close() }()
	for {
		select {
		case msg, ok := <-dc.send:
			dc.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if !ok {
				dc.conn.WriteMessage(websocket.CloseMessage, []byte{}) //nolint:errcheck
				return
			}
			if err := dc.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			dc.conn.SetWriteDeadline(time.Now().Add(writeWait)) //nolint:errcheck
			if err := dc.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (dc *deviceConn) readPump(onClose func()) {
	defer onClose()
	dc.conn.SetReadLimit(maxMsgSize)
	dc.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
	dc.conn.SetPongHandler(func(string) error {
		dc.conn.SetReadDeadline(time.Now().Add(pongWait)) //nolint:errcheck
		return nil
	})
	for {
		_, _, err := dc.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
