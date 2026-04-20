package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires org (branches + departments) routes under /api/v1.
func RegisterRoutes(v1 *gin.RouterGroup, branches *BranchHandler, depts *DepartmentHandler, authMW gin.HandlerFunc) {
	b := v1.Group("/branches", authMW)
	{
		b.GET("", branches.List)
		b.POST("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), branches.Create)
		b.GET("/:id", branches.GetByID)
		b.PUT("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), branches.Update)
	}

	d := v1.Group("/departments", authMW)
	{
		d.GET("", depts.List)
		d.POST("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), depts.Create)
		d.GET("/:id", depts.GetByID)
		d.PUT("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), depts.Update)
	}
}
