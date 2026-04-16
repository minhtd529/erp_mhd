package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires CRM routes under /api/v1.
func RegisterRoutes(v1 *gin.RouterGroup, clients *ClientHandler, authMW gin.HandlerFunc) {
	c := v1.Group("/clients", authMW)
	{
		c.GET("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), clients.List)
		c.POST("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), clients.Create)
		c.GET("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), clients.GetByID)
		c.PUT("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), clients.Update)
		c.DELETE("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), clients.Delete)
	}
}
