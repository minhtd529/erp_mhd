package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/internal/hrm/usecase"
)

// ProfileHandler handles /api/v1/me/profile endpoints for HRM.
type ProfileHandler struct {
	uc *usecase.ProfileUseCase
}

func NewProfileHandler(uc *usecase.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{uc: uc}
}

// GetMyProfile handles GET /api/v1/me/hrm-profile.
func (h *ProfileHandler) GetMyProfile(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}

	resp, err := h.uc.GetMyProfile(c.Request.Context(), caller)
	if err != nil {
		h.handleProfileErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// UpdateMyProfile handles PUT /api/v1/me/hrm-profile.
func (h *ProfileHandler) UpdateMyProfile(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}

	var req usecase.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.UpdateMyProfile(c.Request.Context(), caller, req, c.ClientIP())
	if err != nil {
		h.handleProfileErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *ProfileHandler) handleProfileErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, errResp("EMPLOYEE_NOT_FOUND", "No employee record found for current user"))
	default:
		log.Printf("ERROR hrm.profile: %v", err)
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
