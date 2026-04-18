package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
	"github.com/mdh/erp-audit/api/internal/commission/usecase"
)

type RecordHandler struct {
	uc *usecase.RecordUseCase
}

func NewRecordHandler(uc *usecase.RecordUseCase) *RecordHandler { return &RecordHandler{uc: uc} }

// List handles GET /commissions/records
func (h *RecordHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var f domain.ListRecordsFilter
	if v := c.Query("status"); v != "" {
		f.Status = domain.CommissionStatus(v)
	}
	if v := c.Query("salesperson_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.SalespersonID = &id
		}
	}
	if v := c.Query("engagement_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.EngagementID = &id
		}
	}

	result, err := h.uc.List(c.Request.Context(), f, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// Approve handles POST /commissions/records/:id/approve
func (h *RecordHandler) Approve(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid record ID"))
		return
	}
	resp, err := h.uc.Approve(c.Request.Context(), id, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapRecordErr(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

// MarkPaid handles POST /commissions/records/:id/mark-paid
func (h *RecordHandler) MarkPaid(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid record ID"))
		return
	}
	var body struct {
		PayoutReference string `json:"payout_reference" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.MarkPaid(c.Request.Context(), id, body.PayoutReference, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapRecordErr(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

// BulkApprove handles POST /commissions/records/bulk-approve
func (h *RecordHandler) BulkApprove(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	var req usecase.BulkApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.uc.BulkApprove(c.Request.Context(), req, callerID, clientIP(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// BulkPay handles POST /commissions/records/bulk-pay
func (h *RecordHandler) BulkPay(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	var req usecase.BulkPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.uc.BulkPay(c.Request.Context(), req, callerID, clientIP(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// Clawback handles POST /commissions/records/:id/clawback
func (h *RecordHandler) Clawback(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid record ID"))
		return
	}
	var body struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Clawback(c.Request.Context(), id, body.Reason, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapRecordErr(err))
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// MyCommissionStatement handles GET /me/commissions/statement?period=2026-Q1
func (h *RecordHandler) MyCommissionStatement(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	period := c.Query("period")
	if period == "" {
		c.JSON(http.StatusBadRequest, errResp("MISSING_PERIOD", "period query param required (e.g. 2026-Q1, 2026-01, 2026)"))
		return
	}
	stmt, err := h.uc.GetStatement(c.Request.Context(), callerID, period)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_PERIOD", err.Error()))
		return
	}
	c.JSON(http.StatusOK, stmt)
}

// ExportMyCommissions handles GET /me/commissions/export?period=2026-Q1&format=csv
func (h *RecordHandler) ExportMyCommissions(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	period := c.DefaultQuery("period", "")
	if period == "" {
		c.JSON(http.StatusBadRequest, errResp("MISSING_PERIOD", "period query param required"))
		return
	}
	data, err := h.uc.ExportStatementCSV(c.Request.Context(), callerID, period)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_PERIOD", err.Error()))
		return
	}
	filename := "commission-statement-" + period + ".csv"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// MyCommissions handles GET /me/commissions
func (h *RecordHandler) MyCommissions(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := domain.CommissionStatus(c.Query("status"))

	result, err := h.uc.MyCommissions(c.Request.Context(), callerID, status, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// MyCommissionSummary handles GET /me/commissions/summary
func (h *RecordHandler) MyCommissionSummary(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	summary, err := h.uc.MyCommissionSummary(c.Request.Context(), callerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, summary)
}

// TeamCommissions handles GET /commissions/team
func (h *RecordHandler) TeamCommissions(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	managerID := callerID
	if v := c.Query("manager_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			managerID = id
		}
	}

	result, err := h.uc.TeamCommissions(c.Request.Context(), managerID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func mapRecordErr(err error) (int, gin.H) {
	switch {
	case errors.Is(err, domain.ErrCommissionRecordNotFound):
		return http.StatusNotFound, errResp("COMMISSION_RECORD_NOT_FOUND", "Commission record not found")
	case errors.Is(err, domain.ErrCommissionRecordImmutable):
		return http.StatusUnprocessableEntity, errResp("COMMISSION_RECORD_IMMUTABLE", "This record cannot be modified")
	case errors.Is(err, domain.ErrRecordNotApprovable):
		return http.StatusUnprocessableEntity, errResp("RECORD_NOT_APPROVABLE", "Only accrued records can be approved")
	case errors.Is(err, domain.ErrRecordNotPayable):
		return http.StatusUnprocessableEntity, errResp("RECORD_NOT_PAYABLE", "Only approved records can be marked as paid")
	default:
		return http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred")
	}
}
