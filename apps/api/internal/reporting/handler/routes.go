package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/pkg/middleware"
)

func RegisterRoutes(
	v1 *gin.RouterGroup,
	dashH *DashboardHandler,
	reportH *ReportHandler,
	authMW gin.HandlerFunc,
) {
	requirePartner := middleware.RequireRole("FIRM_PARTNER")
	requireManager := middleware.RequireRole("FIRM_PARTNER", "AUDIT_MANAGER")
	requireStaff := middleware.RequireRole("FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF")
	requireAdmin := middleware.RequireRole("SUPER_ADMIN")

	// Dashboards
	dashboard := v1.Group("/dashboard", authMW)
	{
		dashboard.GET("/executive", requirePartner, dashH.Executive)
		dashboard.GET("/manager", requireManager, dashH.Manager)
		dashboard.GET("/personal", requireStaff, dashH.Personal)
	}

	// Reports
	reports := v1.Group("/reports", authMW)
	{
		reports.GET("/revenue", requirePartner, reportH.Revenue)
		reports.GET("/utilization", requirePartner, reportH.Utilization)
		reports.GET("/ar-aging", requirePartner, reportH.ARaging)
		reports.GET("/engagement-status", requireManager, reportH.EngagementStatus)
		reports.GET("/commission-pending", requirePartner, reportH.CommissionSummary)
		reports.GET("/revenue-by-salesperson", requirePartner, reportH.RevenueByStaff)
		reports.GET("/mv-status", requirePartner, reportH.MVRefreshStatus)
		// Commission reports (v1.2)
		reports.GET("/commission-payout", requirePartner, reportH.CommissionPayout)
		reports.GET("/commission-by-service", requirePartner, reportH.CommissionByService)
		reports.GET("/commission-pending-detail", requirePartner, reportH.CommissionPending)
		reports.GET("/commission-clawback", requirePartner, reportH.CommissionClawback)
	}

	// Admin
	admin := v1.Group("/admin", authMW, requireAdmin)
	{
		admin.POST("/refresh-materialized-views", reportH.RefreshMVs)
	}
}
