// Package handler exposes the notification HTTP endpoints.
package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
	"github.com/mdh/erp-audit/api/internal/notification/domain"
	"github.com/mdh/erp-audit/api/internal/notification/usecase"
)

// Handler serves GET /api/v1/notifications and POST /api/v1/notifications/:id/read.
type Handler struct{ uc *usecase.UseCase }

func New(uc *usecase.UseCase) *Handler { return &Handler{uc: uc} }

func errResp(code, msg string) gin.H {
	return gin.H{"error": gin.H{"code": code, "message": msg}}
}

func callerID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(pkgauth.CtxUserID)
	if !ok {
		c.JSON(http.StatusUnauthorized, errResp("UNAUTHORIZED", "Authentication required"))
		return uuid.UUID{}, false
	}
	id, ok := raw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, errResp("UNAUTHORIZED", "Authentication required"))
		return uuid.UUID{}, false
	}
	return id, true
}

// List handles GET /api/v1/notifications?page=1&size=20.
func (h *Handler) List(c *gin.Context) {
	userID, ok := callerID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	result, err := h.uc.List(c.Request.Context(), userID, page, size)
	if err != nil {
		log.Printf("ERROR notification.List: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// MarkRead handles POST /api/v1/notifications/:id/read.
func (h *Handler) MarkRead(c *gin.Context) {
	userID, ok := callerID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid notification ID"))
		return
	}

	if err := h.uc.MarkRead(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, domain.ErrNotificationNotFound) {
			c.JSON(http.StatusNotFound, errResp("NOTIFICATION_NOT_FOUND", "Notification not found"))
			return
		}
		log.Printf("ERROR notification.MarkRead: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
