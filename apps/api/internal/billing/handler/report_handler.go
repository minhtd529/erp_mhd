package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/internal/billing/usecase"
)

// ReportHandler handles billing report and export endpoints.
type ReportHandler struct {
	uc *usecase.ReportUseCase
}

// NewReportHandler constructs a ReportHandler.
func NewReportHandler(uc *usecase.ReportUseCase) *ReportHandler {
	return &ReportHandler{uc: uc}
}

// PeriodSummary handles GET /billing/reports/period-summary?start=&end=
func (h *ReportHandler) PeriodSummary(c *gin.Context) {
	start, end, ok := parsePeriod(c)
	if !ok {
		return
	}
	result, err := h.uc.GetPeriodSummary(c.Request.Context(), usecase.PeriodSummaryRequest{Start: start, End: end})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// PaymentSummary handles GET /billing/reports/payment-summary?start=&end=
func (h *ReportHandler) PaymentSummary(c *gin.Context) {
	start, end, ok := parsePeriod(c)
	if !ok {
		return
	}
	result, err := h.uc.GetPaymentSummary(c.Request.Context(), usecase.PeriodSummaryRequest{Start: start, End: end})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// ExportInvoices handles GET /invoices/export?format=csv|xlsx&status=ISSUED
// format defaults to "csv". When format=xlsx an Excel workbook is returned.
func (h *ReportHandler) ExportInvoices(c *gin.Context) {
	var f domain.ListInvoicesFilter
	if s := c.Query("status"); s != "" {
		f.Status = domain.InvoiceStatus(s)
	}
	if engStr := c.Query("engagement_id"); engStr != "" {
		if id, err := uuid.Parse(engStr); err == nil {
			f.EngagementID = &id
		}
	}
	if clientStr := c.Query("client_id"); clientStr != "" {
		if id, err := uuid.Parse(clientStr); err == nil {
			f.ClientID = &id
		}
	}

	format := c.DefaultQuery("format", "csv")
	switch format {
	case "xlsx":
		data, err := h.uc.ExportInvoicesXLSX(c.Request.Context(), f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
			return
		}
		filename := fmt.Sprintf("invoices_%s.xlsx", time.Now().Format("20060102"))
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
	default:
		data, err := h.uc.ExportInvoicesCSV(c.Request.Context(), f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
			return
		}
		filename := fmt.Sprintf("invoices_%s.csv", time.Now().Format("20060102"))
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Data(http.StatusOK, "text/csv", data)
	}
}

// ExportPeriodSummary handles GET /billing/reports/period-summary/export?format=xlsx&start=&end=
func (h *ReportHandler) ExportPeriodSummary(c *gin.Context) {
	start, end, ok := parsePeriod(c)
	if !ok {
		return
	}
	format := c.DefaultQuery("format", "xlsx")
	if format != "xlsx" {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", "Supported formats: xlsx"))
		return
	}
	data, err := h.uc.ExportPeriodSummaryXLSX(c.Request.Context(), usecase.PeriodSummaryRequest{Start: start, End: end})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	filename := fmt.Sprintf("billing_summary_%s_%s.xlsx", start.Format("20060102"), end.Format("20060102"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// parsePeriod extracts and validates start/end query params.
func parsePeriod(c *gin.Context) (start, end time.Time, ok bool) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", "start and end query params are required (YYYY-MM-DD)"))
		return
	}
	var err error
	start, err = time.Parse("2006-01-02", startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", "start must be YYYY-MM-DD format"))
		return
	}
	end, err = time.Parse("2006-01-02", endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", "end must be YYYY-MM-DD format"))
		return
	}
	if end.Before(start) {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", "end must be on or after start"))
		return
	}
	ok = true
	return
}
