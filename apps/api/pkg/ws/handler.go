package ws

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// TokenClaims carries the subset of JWT claims needed by the hub.
type TokenClaims struct {
	UserID uuid.UUID
	Roles  []string
}

// TokenValidator is a minimal interface satisfied by pkg/auth.JWTService.
type TokenValidator interface {
	ValidateAccessToken(tokenStr string) (*pkgauth.TokenClaims, error)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	// Allow all origins in dev; restrict in production via CORS middleware.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handler handles GET /api/v1/events/stream?token=<JWT>&channels=crm,global
type Handler struct {
	hub    *Hub
	jwtSvc TokenValidator
}

// NewHandler constructs a WebSocket upgrade handler.
func NewHandler(hub *Hub, jwtSvc TokenValidator) *Handler {
	return &Handler{hub: hub, jwtSvc: jwtSvc}
}

// ServeWS upgrades the HTTP connection to WebSocket.
//
//   GET /api/v1/events/stream?token=<access_token>&channels=global,crm
//
// Token is passed as a query param because the WebSocket handshake does not
// support custom request headers in browsers.
func (h *Handler) ServeWS(c *gin.Context) {
	// 1. Authenticate via query-param JWT
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "MISSING_TOKEN", "message": "token query param required"})
		return
	}

	claims, err := h.jwtSvc.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "INVALID_TOKEN", "message": "token is invalid or expired"})
		return
	}

	// 2. Parse requested channels (default: "global")
	channelsParam := c.Query("channels")
	channels := ParseChannels(channelsParam)

	// 3. Upgrade the HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// upgrader already wrote the error response
		return
	}

	// 4. Create client and start pumps
	client := NewClient(h.hub, conn, claims.UserID, channels)
	go client.WritePump()
	go client.ReadPump()
}

// RegisterRoutes registers the WebSocket endpoint on the given router group.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler) {
	v1.GET("/events/stream", h.ServeWS)
}

// ParseChannels splits a comma-separated channel list.  Returns ["global"]
// when the input is empty or blank.  Exported so tests and external callers can
// use the same normalisation logic.
func ParseChannels(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"global"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if ch := strings.TrimSpace(p); ch != "" {
			out = append(out, ch)
		}
	}
	if len(out) == 0 {
		return []string{"global"}
	}
	return out
}
