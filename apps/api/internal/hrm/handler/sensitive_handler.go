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

// SensitiveHandler handles /api/v1/hrm/employees/:id/sensitive endpoints.
type SensitiveHandler struct {
	uc *usecase.SensitiveUseCase
}

func NewSensitiveHandler(uc *usecase.SensitiveUseCase) *SensitiveHandler {
	return &SensitiveHandler{uc: uc}
}

// GetSensitive handles GET /api/v1/hrm/employees/:id/sensitive.
// Decrypts and returns PII fields. Always writes audit log EMPLOYEE_PII_ACCESSED.
func (h *SensitiveHandler) GetSensitive(c *gin.Context) {
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
	resp, err := h.uc.GetSensitive(c.Request.Context(), id, caller, roles, c.ClientIP())
	if err != nil {
		h.handleSensitiveErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateSensitive handles PUT /api/v1/hrm/employees/:id/sensitive.
// Encrypts and persists PII fields. Writes audit log EMPLOYEE_PII_UPDATED.
func (h *SensitiveHandler) UpdateSensitive(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid employee ID"))
		return
	}

	var req usecase.UpdateSensitiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	roles := callerRoles(c)
	if err := h.uc.UpdateSensitive(c.Request.Context(), id, req, caller, roles, c.ClientIP()); err != nil {
		h.handleSensitiveErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Sensitive fields updated successfully"})
}

func (h *SensitiveHandler) handleSensitiveErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, errResp("EMPLOYEE_NOT_FOUND", "Employee not found"))
	case errors.Is(err, domain.ErrInsufficientPermission):
		c.JSON(http.StatusForbidden, errResp("INSUFFICIENT_PERMISSION", "You do not have permission to access sensitive data"))
	default:
		log.Printf("ERROR hrm.sensitive: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
