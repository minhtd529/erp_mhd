package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires Engagement routes under /api/v1.
func RegisterRoutes(
	v1 *gin.RouterGroup,
	engH *EngagementHandler,
	teamH *TeamHandler,
	taskH *TaskHandler,
	costH *CostHandler,
	authMW gin.HandlerFunc,
) {
	e := v1.Group("/engagements", authMW)
	{
		e.GET("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), engH.List)
		e.POST("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), engH.Create)
		e.GET("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), engH.GetByID)
		e.PUT("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), engH.Update)
		e.DELETE("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), engH.Delete)
		e.POST("/:id/activate", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), engH.Activate)
		e.POST("/:id/complete", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), engH.Complete)

		// Team members
		e.GET("/:id/members", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), teamH.List)
		e.POST("/:id/members", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), teamH.Assign)
		e.PUT("/:id/members/:member_id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), teamH.Update)
		e.DELETE("/:id/members/:member_id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), teamH.Unassign)

		// Tasks
		e.GET("/:id/tasks", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), taskH.List)
		e.POST("/:id/tasks", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), taskH.Create)
		e.PUT("/:id/tasks/:task_id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), taskH.Update)

		// Direct costs
		e.GET("/:id/costs", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), costH.List)
		e.POST("/:id/costs", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), costH.Create)
		e.POST("/:id/costs/:cost_id/submit", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), costH.Submit)
		e.POST("/:id/costs/:cost_id/approve", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), costH.Approve)
		e.POST("/:id/costs/:cost_id/reject", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), costH.Reject)
	}
}
