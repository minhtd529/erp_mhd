package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/tax/usecase"
)

type ComplianceHandler struct {
	uc *usecase.ComplianceUseCase
}

func NewComplianceHandler(uc *usecase.ComplianceUseCase) *ComplianceHandler {
	return &ComplianceHandler{uc: uc}
}

// GetComplianceStatus handles GET /clients/:client_id/tax/compliance-status
func (h *ComplianceHandler) GetComplianceStatus(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client_id"))
		return
	}
	status, err := h.uc.GetClientComplianceStatus(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, status)
}

// Dashboard handles GET /tax/dashboard
func (h *ComplianceHandler) Dashboard(c *gin.Context) {
	var from, to *time.Time
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = &t
		}
	}
	deadlines, err := h.uc.DashboardDeadlines(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"deadlines": deadlines, "count": len(deadlines)})
}

// OverdueAlerts handles GET /tax/overdue-alerts
func (h *ComplianceHandler) OverdueAlerts(c *gin.Context) {
	deadlines, err := h.uc.ListOverdueAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"overdue": deadlines, "count": len(deadlines)})
}
