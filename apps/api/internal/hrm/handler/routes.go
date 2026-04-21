package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires all HRM routes under /api/v1.
// SPEC §13.1 — organization (14) + employees (15) + profile (2) = 31 endpoints total.
func RegisterRoutes(
	v1 *gin.RouterGroup,
	org *OrgHandler,
	emp *EmployeeHandler,
	dep *DependentHandler,
	con *ContractHandler,
	prof *ProfileHandler,
	authMW gin.HandlerFunc,
) {
	// ── Organization ──────────────────────────────────────────────────────────
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

	// ── Employees ────────────────────────────────────────────────────────────
	e := v1.Group("/hrm/employees", authMW)
	{
		// Core CRUD — scope enforcement in usecase
		e.GET("", emp.ListEmployees)
		e.GET("/:id", emp.GetEmployee)
		e.POST("", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER"), emp.CreateEmployee)
		e.PUT("/:id", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER"), emp.UpdateEmployee)
		e.DELETE("/:id", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER"), emp.DeleteEmployee)

		// Dependents
		e.GET("/:id/dependents", dep.ListDependents)
		e.POST("/:id/dependents", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER", "HR_STAFF"), dep.CreateDependent)
		e.PUT("/:id/dependents/:dep_id", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER", "HR_STAFF"), dep.UpdateDependent)
		e.DELETE("/:id/dependents/:dep_id", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER"), dep.DeleteDependent)

		// Contracts
		e.GET("/:id/contracts", con.ListContracts)
		e.POST("/:id/contracts", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER"), con.CreateContract)
		e.PUT("/:id/contracts/:cid", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER"), con.UpdateContract)
		e.POST("/:id/contracts/:cid/terminate", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER"), con.TerminateContract)
	}

	// ── My profile ────────────────────────────────────────────────────────────
	me := v1.Group("/me", authMW)
	{
		me.GET("/hrm-profile", prof.GetMyProfile)
		me.PUT("/hrm-profile", prof.UpdateMyProfile)
	}
}
