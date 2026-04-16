package ws_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	pkgws "github.com/mdh/erp-audit/api/pkg/ws"
)

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

// newConnPair creates a connected WebSocket pair for testing.
//
// Returns:
//   - serverConn: the *websocket.Conn the Hub side receives (server perspective)
//   - clientConn: the *websocket.Conn the test uses to send/receive (browser perspective)
//
// The returned httptest.Server is registered for cleanup automatically.
func newConnPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()

	serverConnCh := make(chan *websocket.Conn, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		serverConnCh <- conn
	}))
	t.Cleanup(ts.Close)

	url := "ws" + strings.TrimPrefix(ts.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	t.Cleanup(func() { clientConn.Close() })

	serverConn := <-serverConnCh
	return serverConn, clientConn
}

// makeClient creates a Client registered with hub using a real WebSocket pair.
func makeClient(t *testing.T, hub *pkgws.Hub, channels []string) (*pkgws.Client, *websocket.Conn) {
	t.Helper()
	serverConn, clientConn := newConnPair(t)
	client := pkgws.NewClient(hub, serverConn, uuid.New(), channels)
	return client, clientConn
}
