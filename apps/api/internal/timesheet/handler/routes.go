package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires Timesheet routes under /api/v1.
func RegisterRoutes(
	v1 *gin.RouterGroup,
	tsH *TimesheetHandler,
	entryH *EntryHandler,
	attendanceH *AttendanceHandler,
	authMW gin.HandlerFunc,
) {
	ts := v1.Group("/timesheets", authMW)
	{
		ts.GET("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), tsH.List)
		ts.GET("/:week_id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), tsH.Get)
		ts.POST("/:week_id/submit", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), tsH.Submit)
		ts.POST("/:week_id/approve", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), tsH.Approve)
		ts.POST("/:week_id/reject", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), tsH.Reject)
		ts.POST("/:week_id/lock", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), tsH.Lock)

		// Time entries
		ts.GET("/:week_id/entries", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), entryH.List)
		ts.POST("/:week_id/entries", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), entryH.Create)
		ts.PUT("/:week_id/entries/:entry_id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), entryH.Update)
		ts.DELETE("/:week_id/entries/:entry_id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), entryH.Delete)
	}

	att := v1.Group("/attendance", authMW)
	{
		att.POST("/check-in", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), attendanceH.CheckIn)
		att.POST("/check-out", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), attendanceH.CheckOut)
		att.GET("/my-records", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), attendanceH.MyRecords)
	}
}
