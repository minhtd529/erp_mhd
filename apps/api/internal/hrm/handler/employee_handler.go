// Package handler provides the HTTP layer for the HRM bounded context.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/internal/hrm/usecase"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// EmployeeHandler handles /api/v1/employees/* endpoints.
type EmployeeHandler struct {
	uc *usecase.EmployeeUseCase
}

// NewEmployeeHandler constructs an EmployeeHandler.
func NewEmployeeHandler(uc *usecase.EmployeeUseCase) *EmployeeHandler {
	return &EmployeeHandler{uc: uc}
}

// List handles GET /api/v1/employees.
func (h *EmployeeHandler) List(c *gin.Context) {
	var req usecase.EmployeeListRequest
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

	result, err := h.uc.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// Create handles POST /api/v1/employees.
func (h *EmployeeHandler) Create(c *gin.Context) {
	var req usecase.EmployeeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.Create(c.Request.Context(), req, callerID(c), c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetByID handles GET /api/v1/employees/:id.
func (h *EmployeeHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Update handles PUT /api/v1/employees/:id.
func (h *EmployeeHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	var req usecase.EmployeeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.Update(c.Request.Context(), id, req, callerID(c), c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateBankDetails handles PUT /api/v1/employees/:id/bank-details.
// Requires SUPER_ADMIN or FIRM_PARTNER. The plaintext account number is
// encrypted in the usecase layer before persistence.
func (h *EmployeeHandler) UpdateBankDetails(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	var req usecase.EmployeeBankDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	if err := h.uc.UpdateBankDetails(c.Request.Context(), id, req, callerID(c), c.ClientIP()); err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// Delete handles DELETE /api/v1/employees/:id.
func (h *EmployeeHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id, callerID(c), c.ClientIP()); err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (h *EmployeeHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, errResp("EMPLOYEE_NOT_FOUND", "Employee not found"))
	case errors.Is(err, domain.ErrDuplicateEmail):
		c.JSON(http.StatusConflict, errResp("DUPLICATE_EMAIL", "Email already registered"))
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_STATE_TRANSITION", "Invalid state transition"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}

func errResp(code, msg string) gin.H {
	return gin.H{"error": code, "message": msg}
}

func callerID(c *gin.Context) *uuid.UUID {
	raw, ok := c.Get(pkgauth.CtxUserID)
	if !ok {
		return nil
	}
	id, ok := raw.(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}
