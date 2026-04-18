package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/internal/reporting/domain"
	"github.com/mdh/erp-audit/api/internal/reporting/usecase"
)

type ReportHandler struct {
	uc *usecase.ReportUseCase
}

func NewReportHandler(uc *usecase.ReportUseCase) *ReportHandler {
	return &ReportHandler{uc: uc}
}

func parseFilter(c *gin.Context) domain.ReportFilter {
	f := domain.ReportFilter{Format: c.DefaultQuery("format", "json")}
	if v, err := strconv.Atoi(c.Query("year")); err == nil {
		f.Year = v
	}
	if v, err := strconv.Atoi(c.Query("month")); err == nil {
		f.Month = v
	}
	return f
}

// Revenue handles GET /reports/revenue
func (h *ReportHandler) Revenue(c *gin.Context) {
	f := parseFilter(c)
	data, err := h.uc.RevenueReport(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate revenue report"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "format": f.Format})
}

// Utilization handles GET /reports/utilization
func (h *ReportHandler) Utilization(c *gin.Context) {
	f := parseFilter(c)
	data, err := h.uc.UtilizationReport(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate utilization report"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// ARaging handles GET /reports/ar-aging
func (h *ReportHandler) ARaging(c *gin.Context) {
	data, err := h.uc.ARAgingReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate AR aging report"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// EngagementStatus handles GET /reports/engagement-status
func (h *ReportHandler) EngagementStatus(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	data, err := h.uc.EngagementStatusReport(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate engagement status report"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// CommissionSummary handles GET /reports/commission-pending (and commission summary)
func (h *ReportHandler) CommissionSummary(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "12"))
	data, err := h.uc.CommissionSummaryReport(c.Request.Context(), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate commission report"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "months": months})
}

// RevenueByStaff handles GET /reports/revenue-by-salesperson
func (h *ReportHandler) RevenueByStaff(c *gin.Context) {
	f := parseFilter(c)
	data, err := h.uc.RevenueByStaffReport(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate revenue-by-salesperson report"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// MVRefreshStatus handles GET /reports/mv-status
func (h *ReportHandler) MVRefreshStatus(c *gin.Context) {
	logs, err := h.uc.GetMVRefreshStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to get MV refresh status"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"views": logs})
}

// RefreshMVs handles POST /admin/refresh-materialized-views (SUPER_ADMIN)
func (h *ReportHandler) RefreshMVs(c *gin.Context) {
	if err := h.uc.RefreshMaterializedViews(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, errResp("MV_REFRESH_FAILED", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "materialized views refreshed"})
}

// CommissionPayout handles GET /reports/commission-payout
func (h *ReportHandler) CommissionPayout(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "12"))
	data, err := h.uc.CommissionPayoutReport(c.Request.Context(), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate commission payout report"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "months": months})
}

// CommissionByService handles GET /reports/commission-by-service
func (h *ReportHandler) CommissionByService(c *gin.Context) {
	f := parseFilter(c)
	data, err := h.uc.CommissionByServiceReport(c.Request.Context(), f.Year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate commission-by-service report"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "year": f.Year})
}

// CommissionPending handles GET /reports/commission-pending
func (h *ReportHandler) CommissionPending(c *gin.Context) {
	data, err := h.uc.CommissionPendingReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate commission pending report"))
		return
	}
	c.JSON(http.StatusOK, data)
}

// CommissionClawback handles GET /reports/commission-clawback
func (h *ReportHandler) CommissionClawback(c *gin.Context) {
	months, _ := strconv.Atoi(c.DefaultQuery("months", "3"))
	data, err := h.uc.CommissionClawbackReport(c.Request.Context(), months)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "Failed to generate commission clawback report"))
		return
	}
	c.JSON(http.StatusOK, data)
}
