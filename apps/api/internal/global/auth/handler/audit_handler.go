package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
)

// AuditHandler handles GET /api/v1/audit-logs.
type AuditHandler struct {
	uc *usecase.ListAuditLogsUseCase
}

// NewAuditHandler constructs an AuditHandler.
func NewAuditHandler(uc *usecase.ListAuditLogsUseCase) *AuditHandler {
	return &AuditHandler{uc: uc}
}

// List handles GET /api/v1/audit-logs with optional query filters.
func (h *AuditHandler) List(c *gin.Context) {
	var req usecase.AuditLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 50
	}

	result, err := h.uc.Execute(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}
