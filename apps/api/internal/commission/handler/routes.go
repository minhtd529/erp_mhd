package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/pkg/middleware"
)

func RegisterRoutes(
	v1 *gin.RouterGroup,
	planH *PlanHandler,
	ecH *EngCommissionHandler,
	recordH *RecordHandler,
	authMW gin.HandlerFunc,
) {
	requirePartner := middleware.RequireRole("FIRM_PARTNER")
	requireManager := middleware.RequireRole("FIRM_PARTNER", "AUDIT_MANAGER")
	requireStaff := middleware.RequireRole("FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF")

	// Commission plans
	plans := v1.Group("/commission-plans", authMW)
	{
		plans.GET("", requireStaff, planH.List)
		plans.POST("", requirePartner, planH.Create)
		plans.GET("/:id", requireStaff, planH.GetByID)
		plans.PUT("/:id", requirePartner, planH.Update)
		plans.POST("/:id/deactivate", requirePartner, planH.Deactivate)
	}

	// Engagement commissions
	ecs := v1.Group("/engagement-commissions", authMW)
	{
		ecs.GET("", requireManager, ecH.List)
		ecs.POST("", requireManager, ecH.Create)
		ecs.GET("/:id", requireManager, ecH.GetByID)
		ecs.POST("/:id/cancel", requireManager, ecH.Cancel)
		ecs.POST("/:id/approve", requirePartner, ecH.Approve)
	}

	// Commission records
	records := v1.Group("/commissions/records", authMW)
	{
		records.GET("", requireManager, recordH.List)
		records.POST("/bulk-approve", requirePartner, recordH.BulkApprove)
		records.POST("/bulk-pay", requireManager, recordH.BulkPay)
		records.POST("/:id/approve", requirePartner, recordH.Approve)
		records.POST("/:id/mark-paid", requireManager, recordH.MarkPaid)
		records.POST("/:id/clawback", requirePartner, recordH.Clawback)
	}

	// Team commissions (manager view)
	v1.GET("/commissions/team", authMW, requireManager, recordH.TeamCommissions)

	// Self-service commission views
	me := v1.Group("/me", authMW)
	{
		me.GET("/commissions", recordH.MyCommissions)
		me.GET("/commissions/summary", recordH.MyCommissionSummary)
		me.GET("/commissions/statement", recordH.MyCommissionStatement)
		me.GET("/commissions/export", recordH.ExportMyCommissions)
	}
}
