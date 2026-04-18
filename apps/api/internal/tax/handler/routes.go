package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/pkg/middleware"
)

func RegisterRoutes(
	v1 *gin.RouterGroup,
	deadlineH *DeadlineHandler,
	advisoryH *AdvisoryHandler,
	complianceH *ComplianceHandler,
	authMW gin.HandlerFunc,
) {
	requirePartner := middleware.RequireRole("FIRM_PARTNER")
	requireManager := middleware.RequireRole("FIRM_PARTNER", "AUDIT_MANAGER")
	requireStaff := middleware.RequireRole("FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF")

	// Tax deadlines (client-scoped)
	deadlines := v1.Group("/clients/:client_id/tax-deadlines", authMW)
	{
		deadlines.GET("", requireStaff, deadlineH.List)
		deadlines.POST("", requireManager, deadlineH.Create)
		deadlines.GET("/:id", requireStaff, deadlineH.GetByID)
		deadlines.PUT("/:id", requireManager, deadlineH.Update)
		deadlines.POST("/:id/mark-completed", requireStaff, deadlineH.MarkCompleted)
		deadlines.POST("/auto-generate", requirePartner, deadlineH.AutoGenerate)
	}

	// Advisory records (client-scoped creation, top-level access)
	v1.GET("/clients/:client_id/advisory-records", authMW, requireStaff, advisoryH.List)
	v1.POST("/clients/:client_id/advisory-records", authMW, requireManager, advisoryH.Create)

	advisory := v1.Group("/advisory-records", authMW)
	{
		advisory.GET("/:id", requireStaff, advisoryH.GetByID)
		advisory.PUT("/:id", requireManager, advisoryH.Update)
		advisory.POST("/:id/deliver", requireManager, advisoryH.Deliver)
		advisory.POST("/:id/attach-file", requireManager, advisoryH.AttachFile)
		advisory.GET("/:id/files", requireStaff, advisoryH.ListFiles)
	}

	// Compliance dashboard
	tax := v1.Group("/tax", authMW)
	{
		tax.GET("/dashboard", requireStaff, complianceH.Dashboard)
		tax.GET("/overdue-alerts", requireManager, complianceH.OverdueAlerts)
	}
	v1.GET("/clients/:client_id/tax/compliance-status", authMW, requireStaff, complianceH.GetComplianceStatus)
}
