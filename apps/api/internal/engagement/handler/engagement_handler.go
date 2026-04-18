// Package handler provides the HTTP layer for the Engagement bounded context.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/internal/engagement/usecase"
)

// EngagementHandler handles /api/v1/engagements/* endpoints.
type EngagementHandler struct {
	uc *usecase.EngagementUseCase
}

// NewEngagementHandler constructs an EngagementHandler.
func NewEngagementHandler(uc *usecase.EngagementUseCase) *EngagementHandler {
	return &EngagementHandler{uc: uc}
}

// List handles GET /engagements.
// When ?cursor= is present it returns cursor-paginated results; otherwise offset pagination.
func (h *EngagementHandler) List(c *gin.Context) {
	if c.Query("cursor") != "" || (c.Query("page") == "" && c.Query("cursor") != "") {
		h.listCursor(c)
		return
	}

	var req usecase.EngagementListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	result, err := h.uc.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *EngagementHandler) listCursor(c *gin.Context) {
	var req usecase.EngagementCursorListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	if req.Size == 0 {
		req.Size = 20
	}
	result, err := h.uc.ListCursor(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrEngagementNotFound) {
			c.JSON(http.StatusNotFound, errResp("ENGAGEMENT_NOT_FOUND", "Engagement not found"))
			return
		}
		c.JSON(http.StatusBadRequest, errResp("INVALID_CURSOR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *EngagementHandler) Create(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	var req usecase.EngagementCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *EngagementHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EngagementHandler) Update(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	var req usecase.EngagementUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Update(c.Request.Context(), id, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EngagementHandler) Delete(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id, caller, c.ClientIP()); err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *EngagementHandler) Activate(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	resp, err := h.uc.Activate(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EngagementHandler) Complete(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	resp, err := h.uc.Complete(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EngagementHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEngagementNotFound):
		c.JSON(http.StatusNotFound, errResp("ENGAGEMENT_NOT_FOUND", "Engagement not found"))
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_STATE_TRANSITION", "Invalid state transition"))
	case errors.Is(err, domain.ErrPartnerNotAssigned):
		c.JSON(http.StatusUnprocessableEntity, errResp("PARTNER_NOT_ASSIGNED", "A partner must be assigned before activating"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
