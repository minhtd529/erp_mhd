package handler

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the notification endpoints under /api/v1.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMW gin.HandlerFunc) {
	n := v1.Group("/notifications", authMW)
	n.GET("", h.List)
	n.POST("/:id/read", h.MarkRead)
}
