// Package handler provides the HTTP layer for the Timesheet bounded context.
package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/internal/timesheet/usecase"
)

// TimesheetHandler handles /api/v1/timesheets/* endpoints.
type TimesheetHandler struct {
	uc *usecase.TimesheetUseCase
}

// NewTimesheetHandler constructs a TimesheetHandler.
func NewTimesheetHandler(uc *usecase.TimesheetUseCase) *TimesheetHandler {
	return &TimesheetHandler{uc: uc}
}

// List handles GET /api/v1/timesheets.
// Switches to cursor mode when ?cursor= is present.
func (h *TimesheetHandler) List(c *gin.Context) {
	if c.Query("cursor") != "" {
		h.listCursor(c)
		return
	}

	var req usecase.TimesheetListRequest
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

func (h *TimesheetHandler) listCursor(c *gin.Context) {
	var req usecase.TimesheetCursorListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	if req.Size == 0 {
		req.Size = 20
	}
	result, err := h.uc.ListCursor(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_CURSOR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, result)
}

// Get handles GET /api/v1/timesheets/:week_id.
// week_id is either a UUID (timesheet ID) or YYYY-MM-DD (week start date).
func (h *TimesheetHandler) Get(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	weekParam := c.Param("week_id")

	// Try UUID first, then date.
	if id, err := uuid.Parse(weekParam); err == nil {
		resp, err := h.uc.GetByID(c.Request.Context(), id)
		if err != nil {
			h.handleErr(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	weekStart, err := time.Parse("2006-01-02", weekParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_WEEK", "week_id must be UUID or YYYY-MM-DD"))
		return
	}
	resp, err := h.uc.GetOrCreate(c.Request.Context(), caller, weekStart, caller)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Submit handles POST /api/v1/timesheets/:week_id/submit.
func (h *TimesheetHandler) Submit(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("week_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid timesheet ID"))
		return
	}
	resp, err := h.uc.Submit(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Approve handles POST /api/v1/timesheets/:week_id/approve.
func (h *TimesheetHandler) Approve(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("week_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid timesheet ID"))
		return
	}
	resp, err := h.uc.Approve(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Reject handles POST /api/v1/timesheets/:week_id/reject.
func (h *TimesheetHandler) Reject(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("week_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid timesheet ID"))
		return
	}
	var req usecase.RejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Reject(c.Request.Context(), id, req.Reason, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Lock handles POST /api/v1/timesheets/:week_id/lock.
func (h *TimesheetHandler) Lock(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("week_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid timesheet ID"))
		return
	}
	resp, err := h.uc.Lock(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TimesheetHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTimesheetNotFound):
		c.JSON(http.StatusNotFound, errResp("TIMESHEET_NOT_FOUND", "Timesheet not found"))
	case errors.Is(err, domain.ErrTimesheetLocked):
		c.JSON(http.StatusConflict, errResp("TIMESHEET_LOCKED", "Timesheet is locked"))
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_STATE_TRANSITION", "Invalid state transition"))
	case errors.Is(err, domain.ErrLockNotAcquired):
		c.JSON(http.StatusConflict, errResp("LOCK_NOT_ACQUIRED", "Another operation is in progress, try again"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
