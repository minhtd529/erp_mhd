package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/internal/hrm/usecase"
)

// SalaryHistoryHandler handles /api/v1/hrm/employees/:id/salary-history endpoints.
type SalaryHistoryHandler struct {
	uc *usecase.SalaryHistoryUseCase
}

func NewSalaryHistoryHandler(uc *usecase.SalaryHistoryUseCase) *SalaryHistoryHandler {
	return &SalaryHistoryHandler{uc: uc}
}

// ListSalaryHistory handles GET /api/v1/hrm/employees/:id/salary-history.
// Writes audit log SALARY_HISTORY_ACCESSED on every call.
func (h *SalaryHistoryHandler) ListSalaryHistory(c *gin.Context) {
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
	items, err := h.uc.ListSalaryHistory(c.Request.Context(), id, caller, roles, c.ClientIP())
	if err != nil {
		h.handleSalaryErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateSalaryHistory handles POST /api/v1/hrm/employees/:id/salary-history.
// Appends an immutable salary record. Writes audit log SALARY_HISTORY_CREATED.
func (h *SalaryHistoryHandler) CreateSalaryHistory(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	var req usecase.CreateSalaryHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	roles := callerRoles(c)
	item, err := h.uc.CreateSalaryHistory(c.Request.Context(), id, req, caller, roles, c.ClientIP())
	if err != nil {
		h.handleSalaryErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *SalaryHistoryHandler) handleSalaryErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, errResp("EMPLOYEE_NOT_FOUND", "Employee not found"))
	case errors.Is(err, domain.ErrInsufficientPermission):
		c.JSON(http.StatusForbidden, errResp("INSUFFICIENT_PERMISSION", "You do not have permission to access salary data"))
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
	default:
		log.Printf("ERROR hrm.salaryHistory: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
