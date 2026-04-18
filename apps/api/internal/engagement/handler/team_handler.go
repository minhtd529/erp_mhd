package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/internal/engagement/usecase"
)

// TeamHandler handles /api/v1/engagements/:id/members endpoints.
type TeamHandler struct {
	uc *usecase.TeamUseCase
}

// NewTeamHandler constructs a TeamHandler.
func NewTeamHandler(uc *usecase.TeamUseCase) *TeamHandler {
	return &TeamHandler{uc: uc}
}

func (h *TeamHandler) List(c *gin.Context) {
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	page, size := parsePageSize(c)
	result, err := h.uc.List(c.Request.Context(), engID, usecase.MemberListRequest{Page: page, Size: size})
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *TeamHandler) Assign(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	var req usecase.MemberAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Assign(c.Request.Context(), engID, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *TeamHandler) Update(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	memberID, err := uuid.Parse(c.Param("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid member ID"))
		return
	}
	var req usecase.MemberUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Update(c.Request.Context(), engID, memberID, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TeamHandler) Unassign(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	memberID, err := uuid.Parse(c.Param("member_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid member ID"))
		return
	}
	if err := h.uc.Unassign(c.Request.Context(), engID, memberID, caller, c.ClientIP()); err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *TeamHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEngagementNotFound):
		c.JSON(http.StatusNotFound, errResp("ENGAGEMENT_NOT_FOUND", "Engagement not found"))
	case errors.Is(err, domain.ErrMemberNotFound):
		c.JSON(http.StatusNotFound, errResp("MEMBER_NOT_FOUND", "Team member not found"))
	case errors.Is(err, domain.ErrTeamAllocationExceeds):
		c.JSON(http.StatusUnprocessableEntity, errResp("TEAM_ALLOCATION_EXCEEDS_100", "Total allocation would exceed 100%"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
