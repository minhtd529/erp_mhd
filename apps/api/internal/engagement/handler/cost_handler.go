package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/internal/engagement/usecase"
)

// CostHandler handles /api/v1/engagements/:id/costs endpoints.
type CostHandler struct {
	uc *usecase.CostUseCase
}

// NewCostHandler constructs a CostHandler.
func NewCostHandler(uc *usecase.CostUseCase) *CostHandler {
	return &CostHandler{uc: uc}
}

func (h *CostHandler) List(c *gin.Context) {
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	page, size := parsePageSize(c)
	result, err := h.uc.List(c.Request.Context(), engID, usecase.CostListRequest{Page: page, Size: size})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *CostHandler) Create(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	var req usecase.CostCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), engID, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *CostHandler) Submit(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	engID, costID, ok2 := parseCostIDs(c)
	if !ok2 {
		return
	}
	resp, err := h.uc.Submit(c.Request.Context(), engID, costID, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *CostHandler) Approve(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	engID, costID, ok2 := parseCostIDs(c)
	if !ok2 {
		return
	}
	resp, err := h.uc.Approve(c.Request.Context(), engID, costID, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *CostHandler) Reject(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	engID, costID, ok2 := parseCostIDs(c)
	if !ok2 {
		return
	}
	var req usecase.CostRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Reject(c.Request.Context(), engID, costID, req.Reason, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func parseCostIDs(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return uuid.UUID{}, uuid.UUID{}, false
	}
	costID, err := uuid.Parse(c.Param("cost_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid cost ID"))
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return engID, costID, true
}

func (h *CostHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEngagementNotFound):
		c.JSON(http.StatusNotFound, errResp("ENGAGEMENT_NOT_FOUND", "Engagement not found"))
	case errors.Is(err, domain.ErrCostNotFound):
		c.JSON(http.StatusNotFound, errResp("COST_NOT_FOUND", "Cost not found"))
	case errors.Is(err, domain.ErrCostApprovalRequired):
		c.JSON(http.StatusUnprocessableEntity, errResp("COST_APPROVAL_REQUIRED", "Cost must be in SUBMITTED status for approval"))
	case errors.Is(err, domain.ErrInvalidCostTransition):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_COST_TRANSITION", "Invalid cost status transition"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
