package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/tax/domain"
	"github.com/mdh/erp-audit/api/internal/tax/usecase"
)

type DeadlineHandler struct {
	uc *usecase.TaxDeadlineUseCase
}

func NewDeadlineHandler(uc *usecase.TaxDeadlineUseCase) *DeadlineHandler {
	return &DeadlineHandler{uc: uc}
}

// List handles GET /clients/:client_id/tax-deadlines
func (h *DeadlineHandler) List(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client_id"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := domain.DeadlineStatus(c.Query("status"))

	result, err := h.uc.List(c.Request.Context(), clientID, status, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// Create handles POST /clients/:client_id/tax-deadlines
func (h *DeadlineHandler) Create(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client_id"))
		return
	}
	var req usecase.CreateDeadlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	d, err := h.uc.Create(c.Request.Context(), clientID, req, callerID, clientIP(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusCreated, d)
}

// GetByID handles GET /clients/:client_id/tax-deadlines/:id
func (h *DeadlineHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	d, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(mapDeadlineErr(err))
		return
	}
	c.JSON(http.StatusOK, d)
}

// Update handles PUT /clients/:client_id/tax-deadlines/:id
func (h *DeadlineHandler) Update(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	var req usecase.UpdateDeadlineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	d, err := h.uc.Update(c.Request.Context(), id, req, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapDeadlineErr(err))
		return
	}
	c.JSON(http.StatusOK, d)
}

// MarkCompleted handles POST /clients/:client_id/tax-deadlines/:id/mark-completed
func (h *DeadlineHandler) MarkCompleted(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	var req usecase.MarkCompletedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	d, err := h.uc.MarkCompleted(c.Request.Context(), id, req, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapDeadlineErr(err))
		return
	}
	c.JSON(http.StatusOK, d)
}

// AutoGenerate handles POST /clients/:client_id/tax-deadlines/auto-generate
func (h *DeadlineHandler) AutoGenerate(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client_id"))
		return
	}
	var body struct {
		FiscalYearEnd string `json:"fiscal_year_end" binding:"required"`
		Year          int    `json:"year" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	deadlines, err := h.uc.AutoGenerate(c.Request.Context(), clientID, body.FiscalYearEnd, body.Year, callerID, clientIP(c))
	if err != nil {
		if errors.Is(err, domain.ErrFiscalYearNotConfigured) {
			c.JSON(http.StatusUnprocessableEntity, errResp("FISCAL_YEAR_NOT_CONFIGURED", "Fiscal year must be configured before auto-generating deadlines"))
			return
		}
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"created": deadlines, "count": len(deadlines)})
}

func mapDeadlineErr(err error) (int, gin.H) {
	switch {
	case errors.Is(err, domain.ErrTaxDeadlineNotFound):
		return http.StatusNotFound, errResp("TAX_DEADLINE_NOT_FOUND", "Tax deadline not found")
	case errors.Is(err, domain.ErrInvalidStateTransition):
		return http.StatusUnprocessableEntity, errResp("INVALID_STATE_TRANSITION", "Cannot perform this state transition")
	case errors.Is(err, domain.ErrFiscalYearNotConfigured):
		return http.StatusUnprocessableEntity, errResp("FISCAL_YEAR_NOT_CONFIGURED", "Fiscal year not configured")
	default:
		return http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred")
	}
}
