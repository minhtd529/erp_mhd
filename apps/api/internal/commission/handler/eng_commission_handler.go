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

type EngCommissionHandler struct {
	uc *usecase.EngCommissionUseCase
}

func NewEngCommissionHandler(uc *usecase.EngCommissionUseCase) *EngCommissionHandler {
	return &EngCommissionHandler{uc: uc}
}

func (h *EngCommissionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	var f domain.ListEngCommissionsFilter
	if v := c.Query("engagement_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.EngagementID = &id
		}
	}
	if v := c.Query("salesperson_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.SalespersonID = &id
		}
	}
	if v := c.Query("status"); v != "" {
		f.Status = v
	}

	result, err := h.uc.List(c.Request.Context(), f, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *EngCommissionHandler) Create(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	var req usecase.EngCommissionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), req, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapECErr(err))
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *EngCommissionHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(mapECErr(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EngCommissionHandler) Cancel(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	resp, err := h.uc.Cancel(c.Request.Context(), id, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapECErr(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *EngCommissionHandler) Approve(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	resp, err := h.uc.Approve(c.Request.Context(), id, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapECErr(err))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func mapECErr(err error) (int, gin.H) {
	switch {
	case errors.Is(err, domain.ErrEngCommissionNotFound):
		return http.StatusNotFound, errResp("ENG_COMMISSION_NOT_FOUND", "Engagement commission not found")
	case errors.Is(err, domain.ErrEngCommissionRateExceeds):
		return http.StatusUnprocessableEntity, errResp("ENG_COMMISSION_RATE_EXCEEDS_100", "Total commission rate would exceed 100%")
	default:
		return http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred")
	}
}
