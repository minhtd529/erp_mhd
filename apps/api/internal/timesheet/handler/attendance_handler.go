package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/internal/timesheet/usecase"
)

// AttendanceHandler handles /api/v1/attendance/* endpoints.
type AttendanceHandler struct {
	uc *usecase.AttendanceUseCase
}

// NewAttendanceHandler constructs an AttendanceHandler.
func NewAttendanceHandler(uc *usecase.AttendanceUseCase) *AttendanceHandler {
	return &AttendanceHandler{uc: uc}
}

func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	var req usecase.CheckInRequest
	_ = c.ShouldBindJSON(&req) // optional body
	resp, err := h.uc.CheckIn(c.Request.Context(), req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.CheckOut(c.Request.Context(), caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AttendanceHandler) MyRecords(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	records, err := h.uc.MyRecords(c.Request.Context(), caller)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": records})
}

func (h *AttendanceHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrAlreadyCheckedIn):
		c.JSON(http.StatusConflict, errResp("ALREADY_CHECKED_IN", "You are already checked in"))
	case errors.Is(err, domain.ErrNotCheckedIn):
		c.JSON(http.StatusUnprocessableEntity, errResp("NOT_CHECKED_IN", "No active check-in found"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
