package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires HRM routes under /api/v1.
func RegisterRoutes(v1 *gin.RouterGroup, employees *EmployeeHandler, authMW gin.HandlerFunc) {
	e := v1.Group("/employees", authMW)
	{
		e.GET("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), employees.List)
		e.POST("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), employees.Create)
		e.GET("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), employees.GetByID)
		e.PUT("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), employees.Update)
		e.PUT("/:id/bank-details", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), employees.UpdateBankDetails)
		e.DELETE("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), employees.Delete)
	}
}
