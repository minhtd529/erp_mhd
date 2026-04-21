package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires HRM organization routes under /api/v1.
// SPEC §13.1 — 14 endpoints total.
func RegisterRoutes(v1 *gin.RouterGroup, org *OrgHandler, authMW gin.HandlerFunc) {
	o := v1.Group("/hrm/organization", authMW)
	{
		// Branches — all roles can read; write roles per SPEC §13.1
		o.GET("/branches", org.ListBranches)
		o.GET("/branches/:id", org.GetBranch)
		// HEAD_OF_BRANCH included; scope (own branch, non-critical fields) enforced in usecase.
		o.PUT("/branches/:id", mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO", "HEAD_OF_BRANCH"), org.UpdateBranch)
		o.PUT("/branches/:id/assign-head", mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO"), org.AssignBranchHead)
		o.PUT("/branches/:id/deactivate", mw.RequireRole("SUPER_ADMIN", "CHAIRMAN"), org.DeactivateBranch)

		// Departments — all roles can read; write roles per SPEC §13.1
		o.GET("/departments", org.ListDepts)
		o.GET("/departments/:id", org.GetDept)
		o.PUT("/departments/:id", mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO"), org.UpdateDept)
		o.PUT("/departments/:id/assign-head", mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO", "HR_MANAGER"), org.AssignDeptHead)
		o.PUT("/departments/:id/deactivate", mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO"), org.DeactivateDept)

		// Branch-Department matrix — all roles can read; write roles per SPEC §13.1
		// DELETE uses two path params (composite PK — no surrogate id on branch_departments).
		o.GET("/branch-departments", org.ListBranchDepts)
		o.POST("/branch-departments", mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO"), org.LinkBranchDept)
		o.DELETE("/branch-departments/:branch_id/:dept_id", mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO"), org.UnlinkBranchDept)

		// Org chart — all authenticated roles
		o.GET("/org-chart", org.GetOrgChart)
	}
}
