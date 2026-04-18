package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/pkg/push"
)

// PushHandler manages device registration and the push relay WebSocket.
type PushHandler struct {
	deviceUC *usecase.PushDeviceUseCase
	push2FAUC *usecase.Push2FAUseCase
	relay    *push.Relay
	deviceRepo push.DeviceRepository
}

func NewPushHandler(
	deviceUC *usecase.PushDeviceUseCase,
	push2FAUC *usecase.Push2FAUseCase,
	relay *push.Relay,
	deviceRepo push.DeviceRepository,
) *PushHandler {
	return &PushHandler{deviceUC: deviceUC, push2FAUC: push2FAUC, relay: relay, deviceRepo: deviceRepo}
}

// RegisterDevice handles POST /push/devices/register
func (h *PushHandler) RegisterDevice(c *gin.Context) {
	userID := callerUserIDPush(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "authentication required"))
		return
	}
	var req struct {
		DeviceToken string `json:"device_token" binding:"required"`
		Platform    string `json:"platform" binding:"required,oneof=ios android web_push"`
		DeviceName  string `json:"device_name"`
		AppVersion  string `json:"app_version"`
		OSVersion   string `json:"os_version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}
	dev, err := h.deviceUC.RegisterDevice(c.Request.Context(), push.RegisterDeviceParams{
		UserID:      *userID,
		DeviceToken: req.DeviceToken,
		Platform:    push.DevicePlatform(req.Platform),
		DeviceName:  req.DeviceName,
		AppVersion:  req.AppVersion,
		OSVersion:   req.OSVersion,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("REGISTER_FAILED", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"device": dev})
}

// UnregisterDevice handles DELETE /push/devices
func (h *PushHandler) UnregisterDevice(c *gin.Context) {
	userID := callerUserIDPush(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "authentication required"))
		return
	}
	var req struct {
		DeviceToken string `json:"device_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}
	if err := h.deviceUC.UnregisterDevice(c.Request.Context(), *userID, req.DeviceToken); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("UNREGISTER_FAILED", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "device unregistered"})
}

// ListDevices handles GET /push/devices
func (h *PushHandler) ListDevices(c *gin.Context) {
	userID := callerUserIDPush(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, errorResponse("UNAUTHORIZED", "authentication required"))
		return
	}
	devices, err := h.deviceUC.ListDevices(c.Request.Context(), *userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

// Heartbeat handles POST /push/devices/heartbeat
func (h *PushHandler) Heartbeat(c *gin.Context) {
	var req struct {
		DeviceToken string `json:"device_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}
	if err := h.deviceUC.Heartbeat(c.Request.Context(), req.DeviceToken); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("HEARTBEAT_FAILED", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// PushRelayWS handles WS /ws/push?token={deviceToken}
// The device token authenticates the device; the JWT validates it came from a legit user.
func (h *PushHandler) PushRelayWS(c *gin.Context) {
	deviceToken := c.Query("token")
	if deviceToken == "" {
		c.JSON(http.StatusBadRequest, errorResponse("MISSING_DEVICE_TOKEN", "device_token required"))
		return
	}
	// Validate device exists and is active
	dev, err := h.deviceRepo.FindByToken(c.Request.Context(), deviceToken)
	if err != nil || !dev.IsActive {
		c.JSON(http.StatusUnauthorized, errorResponse("INVALID_DEVICE_TOKEN", "device not registered"))
		return
	}
	// Update last active
	_ = h.deviceRepo.UpdateLastActive(c.Request.Context(), deviceToken)
	// Upgrade to WebSocket and register with relay
	h.relay.ServeDevice(c.Writer, c.Request, deviceToken)
}

// PushResponse handles POST /auth/2fa/push-response (mobile app approve/reject)
func (h *PushHandler) PushResponse(c *gin.Context) {
	var req struct {
		ChallengeID string `json:"challenge_id" binding:"required"`
		Approved    bool   `json:"approved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}
	if err := h.push2FAUC.RespondToPush(c.Request.Context(), req.ChallengeID, req.Approved); err != nil {
		if err == domain.ErrChallengeNotFound {
			c.JSON(http.StatusNotFound, errorResponse("CHALLENGE_NOT_FOUND", "challenge not found or already responded"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("RESPOND_FAILED", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "response recorded"})
}

// PushStatus handles GET /auth/2fa/push-status/:challengeId (browser polling)
func (h *PushHandler) PushStatus(c *gin.Context) {
	challengeID := c.Param("challengeId")
	status, tokens, err := h.push2FAUC.GetPushStatus(c.Request.Context(), challengeID)
	if err != nil {
		if err == domain.ErrChallengeNotFound {
			c.JSON(http.StatusNotFound, errorResponse("CHALLENGE_NOT_FOUND", "challenge not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", err.Error()))
		return
	}
	if tokens != nil {
		c.JSON(http.StatusOK, gin.H{"status": status, "tokens": tokens})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": status})
}

// ResendPush handles POST /auth/2fa/resend-push
// Returns 200 always; device may or may not be online.
func (h *PushHandler) ResendPush(c *gin.Context) {
	var req struct {
		ChallengeID string `json:"challenge_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_REQUEST", err.Error()))
		return
	}
	status, _, err := h.push2FAUC.GetPushStatus(c.Request.Context(), req.ChallengeID)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("CHALLENGE_NOT_FOUND", "challenge not found"))
		return
	}
	if status != domain.PushChallengePending {
		c.JSON(http.StatusConflict, errorResponse("CHALLENGE_NOT_PENDING", "challenge is no longer pending"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "push resent if device is online"})
}

// callerUserIDPush extracts the authenticated user UUID from the Gin context.
func callerUserIDPush(c *gin.Context) *uuid.UUID {
	v, ok := c.Get(string(pkgauth.CtxUserID))
	if !ok {
		return nil
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}
