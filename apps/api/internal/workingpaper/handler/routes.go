package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires WorkingPaper routes under /api/v1.
func RegisterRoutes(
	v1 *gin.RouterGroup,
	wpH *WPHandler,
	reviewH *ReviewHandler,
	tmplH *TemplateHandler,
	folderH *FolderHandler,
	authMW gin.HandlerFunc,
) {
	// Folders nested under engagements
	eng := v1.Group("/engagements/:engagement_id", authMW)
	{
		eng.GET("/folders", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), folderH.List)
		eng.POST("/folders", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), folderH.Create)
		eng.GET("/working-papers", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), wpH.List)
		eng.POST("/working-papers", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), wpH.Create)
	}

	wp := v1.Group("/working-papers", authMW)
	{
		wp.GET("/pending-review", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), wpH.PendingReview)
		wp.GET("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), wpH.GetByID)
		wp.PUT("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), wpH.Update)
		wp.DELETE("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), wpH.Delete)
		wp.POST("/:id/submit", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), wpH.SubmitForReview)
		wp.POST("/:id/finalize", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), wpH.Finalize)
		wp.POST("/:id/sign-off", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), wpH.SignOff)

		// Reviews
		wp.GET("/:id/reviews", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), reviewH.ListReviews)
		wp.POST("/:id/reviews/:role/approve", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), reviewH.Approve)
		wp.POST("/:id/reviews/:role/request-changes", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), reviewH.RequestChanges)
		wp.GET("/:id/reviews/:role/comments", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), reviewH.ListComments)
		wp.POST("/:id/reviews/:role/comments", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), reviewH.AddComment)
		wp.POST("/:id/reviews/:role/comments/:comment_id/resolve", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), reviewH.ResolveComment)
	}

	// Audit templates
	tmpl := v1.Group("/audit-templates", authMW)
	{
		tmpl.GET("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER", "AUDIT_STAFF"), tmplH.List)
		tmpl.POST("", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), tmplH.Create)
		tmpl.PUT("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), tmplH.Update)
		tmpl.DELETE("/:id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER"), tmplH.Retire)
		tmpl.POST("/:id/apply/:engagement_id", mw.RequireRole("SUPER_ADMIN", "FIRM_PARTNER", "AUDIT_MANAGER"), tmplH.ApplyToEngagement)
	}
}
