// Package handler provides the HTTP layer for the WorkingPaper bounded context.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/internal/workingpaper/usecase"
)

// WPHandler handles working paper endpoints.
type WPHandler struct {
	uc *usecase.WorkingPaperUseCase
}

// NewWPHandler constructs a WPHandler.
func NewWPHandler(uc *usecase.WorkingPaperUseCase) *WPHandler {
	return &WPHandler{uc: uc}
}

func (h *WPHandler) List(c *gin.Context) {
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	var req usecase.WPListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.uc.List(c.Request.Context(), engID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *WPHandler) Create(c *gin.Context) {
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	var req usecase.WPCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), engID, req, caller, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *WPHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		mapWPErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WPHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	var req usecase.WPUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Update(c.Request.Context(), id, req, caller, c.ClientIP())
	if err != nil {
		mapWPErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WPHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id, caller, c.ClientIP()); err != nil {
		mapWPErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *WPHandler) SubmitForReview(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.SubmitForReview(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		mapWPErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WPHandler) Finalize(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Finalize(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		mapWPErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WPHandler) SignOff(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid working paper ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.SignOff(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		mapWPErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListAll returns all working papers across engagements (global view).
func (h *WPHandler) ListAll(c *gin.Context) {
	var req usecase.WPListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.uc.ListAll(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// PendingReview returns WPs awaiting action by the caller's reviewer role.
func (h *WPHandler) PendingReview(c *gin.Context) {
	var req usecase.PendingReviewRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.uc.PendingReview(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func mapWPErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrWorkingPaperNotFound):
		c.JSON(http.StatusNotFound, errResp("WORKING_PAPER_NOT_FOUND", "Working paper not found"))
	case errors.Is(err, domain.ErrWorkingPaperLocked):
		c.JSON(http.StatusUnprocessableEntity, errResp("WORKING_PAPER_LOCKED", "Working paper cannot be modified in current state"))
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_STATE_TRANSITION", "Invalid state transition"))
	case errors.Is(err, domain.ErrReviewChainIncomplete):
		c.JSON(http.StatusUnprocessableEntity, errResp("REVIEW_CHAIN_INCOMPLETE", "All reviews must be approved before finalizing"))
	case errors.Is(err, domain.ErrCommentsNotResolved):
		c.JSON(http.StatusUnprocessableEntity, errResp("COMMENTS_NOT_RESOLVED", "All comments must be resolved before finalizing"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
