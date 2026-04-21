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

// EmployeeHandler handles /api/v1/hrm/employees/* endpoints.
type EmployeeHandler struct {
	uc *usecase.EmployeeUseCase
}

func NewEmployeeHandler(uc *usecase.EmployeeUseCase) *EmployeeHandler {
	return &EmployeeHandler{uc: uc}
}

// ListEmployees handles GET /api/v1/hrm/employees.
func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}

	var req usecase.ListEmployeeRequest
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

	roles := callerRoles(c)
	branchID := callerBranchID(c)

	result, err := h.uc.ListEmployees(c.Request.Context(), req, caller, roles, branchID)
	if err != nil {
		log.Printf("ERROR hrm.ListEmployees: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetEmployee handles GET /api/v1/hrm/employees/:id.
func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	roles := callerRoles(c)
	branchID := callerBranchID(c)

	resp, err := h.uc.GetEmployee(c.Request.Context(), id, caller, roles, branchID)
	if err != nil {
		h.handleEmployeeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// CreateEmployee handles POST /api/v1/hrm/employees.
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}

	var req usecase.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.CreateEmployee(c.Request.Context(), req, &caller, c.ClientIP())
	if err != nil {
		h.handleEmployeeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// UpdateEmployee handles PUT /api/v1/hrm/employees/:id.
func (h *EmployeeHandler) UpdateEmployee(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	var req usecase.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.UpdateEmployee(c.Request.Context(), id, req, &caller, c.ClientIP())
	if err != nil {
		h.handleEmployeeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// DeleteEmployee handles DELETE /api/v1/hrm/employees/:id.
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	if err := h.uc.DeleteEmployee(c.Request.Context(), id, &caller, c.ClientIP()); err != nil {
		h.handleEmployeeErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *EmployeeHandler) handleEmployeeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, errResp("EMPLOYEE_NOT_FOUND", "Employee not found"))
	case errors.Is(err, domain.ErrEmployeeEmailConflict):
		c.JSON(http.StatusConflict, errResp("EMPLOYEE_EMAIL_CONFLICT", "Email already exists"))
	case errors.Is(err, domain.ErrInsufficientPermission):
		c.JSON(http.StatusForbidden, errResp("INSUFFICIENT_PERMISSION", "You do not have permission to perform this action"))
	default:
		log.Printf("ERROR hrm.employee: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func callerRoles(c *gin.Context) []string {
	raw, _ := c.Get(pkgauth.CtxRoles)
	if roles, ok := raw.([]string); ok {
		return roles
	}
	return nil
}

func callerBranchID(c *gin.Context) *uuid.UUID {
	raw, _ := c.Get(pkgauth.CtxBranchID)
	if bid, ok := raw.(*uuid.UUID); ok {
		return bid
	}
	return nil
}
