package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/internal/engagement/usecase"
)

// TaskHandler handles /api/v1/engagements/:id/tasks endpoints.
type TaskHandler struct {
	uc *usecase.TaskUseCase
}

// NewTaskHandler constructs a TaskHandler.
func NewTaskHandler(uc *usecase.TaskUseCase) *TaskHandler {
	return &TaskHandler{uc: uc}
}

func (h *TaskHandler) List(c *gin.Context) {
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	page, size := parsePageSize(c)
	result, err := h.uc.List(c.Request.Context(), engID, usecase.TaskListRequest{
		Phase: domain.TaskPhase(c.Query("phase")),
		Page:  page,
		Size:  size,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *TaskHandler) Create(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	var req usecase.TaskCreateRequest
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

func (h *TaskHandler) Update(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	engID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	taskID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid task ID"))
		return
	}
	var req usecase.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Update(c.Request.Context(), engID, taskID, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TaskHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEngagementNotFound):
		c.JSON(http.StatusNotFound, errResp("ENGAGEMENT_NOT_FOUND", "Engagement not found"))
	case errors.Is(err, domain.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, errResp("TASK_NOT_FOUND", "Task not found"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
