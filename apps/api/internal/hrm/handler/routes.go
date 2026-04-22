package handler

import (
	"github.com/gin-gonic/gin"
	mw "github.com/mdh/erp-audit/api/pkg/middleware"
)

// RegisterRoutes wires all HRM routes under /api/v1.
// SPEC §13.1–§13.13 — organization, employees, sensitive PII, salary history, profile,
// user provisioning (§13.12), offboarding (§13.13).
func RegisterRoutes(
	v1 *gin.RouterGroup,
	org *OrgHandler,
	emp *EmployeeHandler,
	dep *DependentHandler,
	con *ContractHandler,
	prof *ProfileHandler,
	sens *SensitiveHandler,
	sal *SalaryHistoryHandler,
	prov *ProvisioningHandler,
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

		// Sensitive PII — SPEC §13.2: HR_MANAGER, CEO, CHAIRMAN, SUPER_ADMIN only
		e.GET("/:id/sensitive",
			mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO", "HR_MANAGER"),
			sens.GetSensitive,
		)
		e.PUT("/:id/sensitive",
			mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO", "HR_MANAGER"),
			sens.UpdateSensitive,
		)

		// Salary history — immutable (no PUT/DELETE per SPEC §13.15)
		// GET: SA, CHAIRMAN, CEO, HR_MANAGER; POST: SA, CEO, HR_MANAGER (CHAIRMAN DENY)
		e.GET("/:id/salary-history",
			mw.RequireRole("SUPER_ADMIN", "CHAIRMAN", "CEO", "HR_MANAGER"),
			sal.ListSalaryHistory,
		)
		e.POST("/:id/salary-history",
			mw.RequireRole("SUPER_ADMIN", "CEO", "HR_MANAGER"),
			sal.CreateSalaryHistory,
		)
	}

	// ── My profile ────────────────────────────────────────────────────────────
	me := v1.Group("/me", authMW)
	{
		me.GET("/hrm-profile", prof.GetMyProfile)
		me.PUT("/hrm-profile", prof.UpdateMyProfile)
	}

	// ── User Provisioning Requests — SPEC §13.12 ──────────────────────────────
	p := v1.Group("/hrm/user-provisioning-requests", authMW)
	{
		p.GET("", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER", "HEAD_OF_BRANCH"), prov.ListProvisioningRequests)
		p.POST("", mw.RequireRole("HR_MANAGER", "HEAD_OF_BRANCH", "CEO"), prov.CreateProvisioningRequest)
		p.GET("/:id", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER", "HEAD_OF_BRANCH"), prov.GetProvisioningRequest)
		p.POST("/:id/branch-approve", mw.RequireRole("HEAD_OF_BRANCH"), prov.BranchApprove)
		p.POST("/:id/branch-reject", mw.RequireRole("HEAD_OF_BRANCH"), prov.BranchReject)
		p.POST("/:id/hr-approve", mw.RequireRole("HR_MANAGER"), prov.HRApprove)
		p.POST("/:id/hr-reject", mw.RequireRole("HR_MANAGER"), prov.HRReject)
		p.POST("/:id/execute", mw.RequireRole("SUPER_ADMIN"), prov.ExecuteProvisioning)
		// Cancel: requester or SA — role check is permissive; ownership enforced in usecase
		p.POST("/:id/cancel", prov.CancelProvisioningRequest)
	}

	// ── Offboarding Checklists — SPEC §13.13 ─────────────────────────────────
	ob := v1.Group("/hrm/offboarding", authMW)
	{
		ob.GET("", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER", "CEO"), prov.ListOffboarding)
		ob.POST("", mw.RequireRole("HR_MANAGER"), prov.CreateOffboarding)
		ob.GET("/:id", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER", "CEO"), prov.GetOffboarding)
		// items/:key — HR, IT (no dedicated IT role in RBAC so HR_STAFF), Finance roles
		ob.PUT("/:id/items/:key", mw.RequireRole("SUPER_ADMIN", "HR_MANAGER", "HR_STAFF", "ACCOUNTANT"), prov.UpdateOffboardingItem)
		ob.POST("/:id/complete", mw.RequireRole("HR_MANAGER"), prov.CompleteOffboarding)
	}
}
