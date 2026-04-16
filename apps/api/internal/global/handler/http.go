package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GlobalHandler handles global/health routes
type GlobalHandler struct{}

func NewGlobalHandler() *GlobalHandler {
	return &GlobalHandler{}
}

// Health returns a simple health check response
func (h *GlobalHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": "v1",
	})
}
