package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
	"github.com/mdh/erp-audit/api/internal/commission/usecase"
)

type PlanHandler struct {
	uc *usecase.PlanUseCase
}

func NewPlanHandler(uc *usecase.PlanUseCase) *PlanHandler { return &PlanHandler{uc: uc} }

func (h *PlanHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var f domain.ListPlansFilter
	if v := c.Query("type"); v != "" {
		f.Type = domain.CommissionType(v)
	}
	if v := c.Query("is_active"); v != "" {
		active := v == "true"
		f.IsActive = &active
	}

	result, err := h.uc.List(c.Request.Context(), f, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *PlanHandler) Create(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	var req usecase.PlanCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), req, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapPlanErr(err))
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *PlanHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid plan ID"))
		return
	}
	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(mapPlanErr(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PlanHandler) Update(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid plan ID"))
		return
	}
	var req usecase.PlanUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Update(c.Request.Context(), id, req, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapPlanErr(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PlanHandler) Deactivate(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid plan ID"))
		return
	}
	resp, err := h.uc.Deactivate(c.Request.Context(), id, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapPlanErr(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func mapPlanErr(err error) (int, gin.H) {
	switch {
	case errors.Is(err, domain.ErrPlanNotFound):
		return http.StatusNotFound, errResp("PLAN_NOT_FOUND", "Commission plan not found")
	case errors.Is(err, domain.ErrPlanCodeConflict):
		return http.StatusConflict, errResp("PLAN_CODE_CONFLICT", "A plan with this code already exists")
	case errors.Is(err, domain.ErrPlanInactive):
		return http.StatusUnprocessableEntity, errResp("PLAN_INACTIVE", "Commission plan is inactive")
	default:
		return http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred")
	}
}
