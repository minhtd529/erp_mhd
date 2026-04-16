package handler

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all global module routes under the given router group
func (h *GlobalHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/health", h.Health)
}
