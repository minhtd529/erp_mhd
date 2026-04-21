package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/internal/hrm/usecase"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// errResp builds a standard error body.
func errResp(code, msg string) gin.H {
	return gin.H{"error": gin.H{"code": code, "message": msg}}
}

// mustCallerID extracts the authenticated user UUID or writes 401 and returns false.
func mustCallerID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(pkgauth.CtxUserID)
	if !ok {
		c.JSON(http.StatusUnauthorized, errResp("UNAUTHORIZED", "Authentication required"))
		return uuid.UUID{}, false
	}
	id, ok := raw.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, errResp("UNAUTHORIZED", "Authentication required"))
		return uuid.UUID{}, false
	}
	return id, true
}

// OrgHandler handles /api/v1/hrm/organization/* endpoints.
type OrgHandler struct {
	uc *usecase.OrgUseCase
}

// NewOrgHandler constructs an OrgHandler.
func NewOrgHandler(uc *usecase.OrgUseCase) *OrgHandler {
	return &OrgHandler{uc: uc}
}

// ─── Branch endpoints ─────────────────────────────────────────────────────────

// ListBranches handles GET /api/v1/hrm/organization/branches.
func (h *OrgHandler) ListBranches(c *gin.Context) {
	var req usecase.ListBranchHRMRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	result, err := h.uc.ListBranches(c.Request.Context(), req)
	if err != nil {
		log.Printf("ERROR hrm.ListBranches: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetBranch handles GET /api/v1/hrm/organization/branches/:id.
func (h *OrgHandler) GetBranch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid branch ID"))
		return
	}

	resp, err := h.uc.GetBranch(c.Request.Context(), id)
	if err != nil {
		h.handleBranchErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateBranch handles PUT /api/v1/hrm/organization/branches/:id.
// Allowed roles: SA, CHAIRMAN, CEO, HEAD_OF_BRANCH (HoB scope enforced in usecase).
func (h *OrgHandler) UpdateBranch(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid branch ID"))
		return
	}

	var req usecase.UpdateBranchHRMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	var callerRoles []string
	if raw, exists := c.Get(pkgauth.CtxRoles); exists {
		if roles, ok2 := raw.([]string); ok2 {
			callerRoles = roles
		}
	}
	var callerBranchID *uuid.UUID
	if raw, exists := c.Get(pkgauth.CtxBranchID); exists {
		if bid, ok2 := raw.(*uuid.UUID); ok2 {
			callerBranchID = bid
		}
	}

	resp, err := h.uc.UpdateBranch(c.Request.Context(), id, req, &caller, c.ClientIP(), callerRoles, callerBranchID)
	if err != nil {
		h.handleBranchErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *OrgHandler) handleBranchErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrBranchNotFound):
		c.JSON(http.StatusNotFound, errResp("BRANCH_NOT_FOUND", "Branch not found"))
	case errors.Is(err, domain.ErrInsufficientPermission):
		c.JSON(http.StatusForbidden, errResp("INSUFFICIENT_PERMISSION", "You do not have permission to perform this action"))
	default:
		log.Printf("ERROR hrm.branch: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}

// AssignBranchHead handles PUT /api/v1/hrm/organization/branches/:id/assign-head.
func (h *OrgHandler) AssignBranchHead(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid branch ID"))
		return
	}

	var req usecase.AssignBranchHeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.AssignBranchHead(c.Request.Context(), id, req, &caller, c.ClientIP())
	if err != nil {
		h.handleBranchErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// DeactivateBranch handles PUT /api/v1/hrm/organization/branches/:id/deactivate.
func (h *OrgHandler) DeactivateBranch(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid branch ID"))
		return
	}

	if err := h.uc.DeactivateBranch(c.Request.Context(), id, &caller, c.ClientIP()); err != nil {
		h.handleBranchErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ─── Department endpoints ─────────────────────────────────────────────────────

// ListDepts handles GET /api/v1/hrm/organization/departments.
func (h *OrgHandler) ListDepts(c *gin.Context) {
	var req usecase.ListDeptHRMRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	result, err := h.uc.ListDepts(c.Request.Context(), req)
	if err != nil {
		log.Printf("ERROR hrm.ListDepts: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetDept handles GET /api/v1/hrm/organization/departments/:id.
func (h *OrgHandler) GetDept(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid department ID"))
		return
	}

	resp, err := h.uc.GetDept(c.Request.Context(), id)
	if err != nil {
		h.handleDeptErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateDept handles PUT /api/v1/hrm/organization/departments/:id.
// Allowed roles: SA, CHAIRMAN, CEO (enforced in routes via RequireRole).
func (h *OrgHandler) UpdateDept(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid department ID"))
		return
	}

	var req usecase.UpdateDeptHRMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.UpdateDept(c.Request.Context(), id, req, &caller, c.ClientIP())
	if err != nil {
		h.handleDeptErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *OrgHandler) handleDeptErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrDeptNotFound):
		c.JSON(http.StatusNotFound, errResp("DEPARTMENT_NOT_FOUND", "Department not found"))
	default:
		log.Printf("ERROR hrm.dept: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}

// AssignDeptHead handles PUT /api/v1/hrm/organization/departments/:id/assign-head.
func (h *OrgHandler) AssignDeptHead(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid department ID"))
		return
	}

	var req usecase.AssignDeptHeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.AssignDeptHead(c.Request.Context(), id, req, &caller, c.ClientIP())
	if err != nil {
		h.handleDeptErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// DeactivateDept handles PUT /api/v1/hrm/organization/departments/:id/deactivate.
func (h *OrgHandler) DeactivateDept(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid department ID"))
		return
	}

	if err := h.uc.DeactivateDept(c.Request.Context(), id, &caller, c.ClientIP()); err != nil {
		h.handleDeptErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ─── BranchDepartment endpoints ───────────────────────────────────────────────

// ListBranchDepts handles GET /api/v1/hrm/organization/branch-departments.
func (h *OrgHandler) ListBranchDepts(c *gin.Context) {
	var req usecase.ListBranchDeptRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	result, err := h.uc.ListBranchDepts(c.Request.Context(), req)
	if err != nil {
		log.Printf("ERROR hrm.ListBranchDepts: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// LinkBranchDept handles POST /api/v1/hrm/organization/branch-departments.
// Allowed roles: SA, CHAIRMAN.
func (h *OrgHandler) LinkBranchDept(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}

	var req usecase.LinkBranchDeptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.LinkBranchDept(c.Request.Context(), req, &caller, c.ClientIP())
	if err != nil {
		h.handleBranchDeptErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UnlinkBranchDept handles DELETE /api/v1/hrm/organization/branch-departments/:branch_id/:dept_id.
// Allowed roles: SA, CHAIRMAN.
// Note: SPEC §13.1 shows /:id but branch_departments has a composite PK; two path params used instead.
func (h *OrgHandler) UnlinkBranchDept(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}

	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid branch_id"))
		return
	}
	deptID, err := uuid.Parse(c.Param("dept_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid dept_id"))
		return
	}

	if err := h.uc.UnlinkBranchDept(c.Request.Context(), branchID, deptID, &caller, c.ClientIP()); err != nil {
		h.handleBranchDeptErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *OrgHandler) handleBranchDeptErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrDuplicateBranchDept):
		c.JSON(http.StatusConflict, errResp("DUPLICATE_BRANCH_DEPT", "Branch-department link already exists"))
	case errors.Is(err, domain.ErrBranchDeptNotFound):
		c.JSON(http.StatusNotFound, errResp("BRANCH_DEPT_NOT_FOUND", "Branch-department link not found or already inactive"))
	default:
		log.Printf("ERROR hrm.branchDept: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}

// ─── Org Chart endpoint ───────────────────────────────────────────────────────

// GetOrgChart handles GET /api/v1/hrm/organization/org-chart.
func (h *OrgHandler) GetOrgChart(c *gin.Context) {
	resp, err := h.uc.GetOrgChart(c.Request.Context())
	if err != nil {
		log.Printf("ERROR hrm.GetOrgChart: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}
