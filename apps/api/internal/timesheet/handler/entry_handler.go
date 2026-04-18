package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/internal/timesheet/usecase"
)

// EntryHandler handles /api/v1/timesheets/:week_id/entries endpoints.
type EntryHandler struct {
	uc *usecase.EntryUseCase
}

// NewEntryHandler constructs an EntryHandler.
func NewEntryHandler(uc *usecase.EntryUseCase) *EntryHandler {
	return &EntryHandler{uc: uc}
}

func (h *EntryHandler) List(c *gin.Context) {
	tsID, err := uuid.Parse(c.Param("week_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid timesheet ID"))
		return
	}
	entries, err := h.uc.List(c.Request.Context(), tsID)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entries})
}

func (h *EntryHandler) Create(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	tsID, err := uuid.Parse(c.Param("week_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid timesheet ID"))
		return
	}
	var req usecase.EntryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), tsID, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *EntryHandler) Update(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	tsID, entryID, ok2 := parseEntryIDs(c)
	if !ok2 {
		return
	}
	var req usecase.EntryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Update(c.Request.Context(), tsID, entryID, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EntryHandler) Delete(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	tsID, entryID, ok2 := parseEntryIDs(c)
	if !ok2 {
		return
	}
	if err := h.uc.Delete(c.Request.Context(), tsID, entryID, caller, c.ClientIP()); err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func parseEntryIDs(c *gin.Context) (uuid.UUID, uuid.UUID, bool) {
	tsID, err := uuid.Parse(c.Param("week_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid timesheet ID"))
		return uuid.UUID{}, uuid.UUID{}, false
	}
	entryID, err := uuid.Parse(c.Param("entry_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid entry ID"))
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return tsID, entryID, true
}

func (h *EntryHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTimesheetNotFound):
		c.JSON(http.StatusNotFound, errResp("TIMESHEET_NOT_FOUND", "Timesheet not found"))
	case errors.Is(err, domain.ErrEntryNotFound):
		c.JSON(http.StatusNotFound, errResp("ENTRY_NOT_FOUND", "Entry not found"))
	case errors.Is(err, domain.ErrEntryDateOutsidePeriod):
		c.JSON(http.StatusUnprocessableEntity, errResp("ENTRY_DATE_OUTSIDE_PERIOD", "Entry date is not within the timesheet week"))
	case errors.Is(err, domain.ErrTimesheetNotEditable):
		c.JSON(http.StatusConflict, errResp("TIMESHEET_NOT_EDITABLE", "Timesheet cannot be edited in its current state"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
