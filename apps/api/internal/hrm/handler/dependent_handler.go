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

// DependentHandler handles /api/v1/hrm/employees/:id/dependents endpoints.
type DependentHandler struct {
	uc *usecase.DependentUseCase
}

func NewDependentHandler(uc *usecase.DependentUseCase) *DependentHandler {
	return &DependentHandler{uc: uc}
}

// ListDependents handles GET /api/v1/hrm/employees/:id/dependents.
func (h *DependentHandler) ListDependents(c *gin.Context) {
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	items, err := h.uc.ListDependents(c.Request.Context(), employeeID)
	if err != nil {
		log.Printf("ERROR hrm.ListDependents: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// CreateDependent handles POST /api/v1/hrm/employees/:id/dependents.
func (h *DependentHandler) CreateDependent(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	var req usecase.CreateDependentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.CreateDependent(c.Request.Context(), employeeID, req, &caller, c.ClientIP())
	if err != nil {
		h.handleDepErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// UpdateDependent handles PUT /api/v1/hrm/employees/:id/dependents/:dep_id.
func (h *DependentHandler) UpdateDependent(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}
	depID, err := uuid.Parse(c.Param("dep_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid dependent ID"))
		return
	}

	var req usecase.UpdateDependentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.UpdateDependent(c.Request.Context(), depID, employeeID, req, &caller, c.ClientIP())
	if err != nil {
		h.handleDepErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// DeleteDependent handles DELETE /api/v1/hrm/employees/:id/dependents/:dep_id.
func (h *DependentHandler) DeleteDependent(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	employeeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}
	depID, err := uuid.Parse(c.Param("dep_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid dependent ID"))
		return
	}

	if err := h.uc.DeleteDependent(c.Request.Context(), depID, employeeID, &caller, c.ClientIP()); err != nil {
		h.handleDepErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *DependentHandler) handleDepErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrDependentNotFound):
		c.JSON(http.StatusNotFound, errResp("DEPENDENT_NOT_FOUND", "Dependent not found"))
	default:
		log.Printf("ERROR hrm.dependent: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
