package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/internal/workingpaper/usecase"
)

// TemplateHandler handles audit template endpoints.
type TemplateHandler struct {
	uc *usecase.TemplateUseCase
}

// NewTemplateHandler constructs a TemplateHandler.
func NewTemplateHandler(uc *usecase.TemplateUseCase) *TemplateHandler {
	return &TemplateHandler{uc: uc}
}

func (h *TemplateHandler) List(c *gin.Context) {
	var req usecase.TemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	result, err := h.uc.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *TemplateHandler) Create(c *gin.Context) {
	var req usecase.TemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), req, caller, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *TemplateHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid template ID"))
		return
	}
	var req usecase.TemplateUpdateRequest
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
		mapTemplateErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TemplateHandler) Retire(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid template ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	if err := h.uc.Retire(c.Request.Context(), id, caller, c.ClientIP()); err != nil {
		mapTemplateErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TemplateHandler) ApplyToEngagement(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid template ID"))
		return
	}
	engID, err := uuid.Parse(c.Param("engagement_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.ApplyToEngagement(c.Request.Context(), id, engID, caller, c.ClientIP())
	if err != nil {
		mapTemplateErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func mapTemplateErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTemplateNotFound):
		c.JSON(http.StatusNotFound, errResp("TEMPLATE_NOT_FOUND", "Template not found"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
