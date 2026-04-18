package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/billing/usecase"
)

// MemoHandler handles credit note / billing memo endpoints.
type MemoHandler struct {
	uc *usecase.MemoUseCase
}

// NewMemoHandler constructs a MemoHandler.
func NewMemoHandler(uc *usecase.MemoUseCase) *MemoHandler {
	return &MemoHandler{uc: uc}
}

func (h *MemoHandler) Create(c *gin.Context) {
	var invoiceID uuid.UUID
	if raw := c.Param("id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
			return
		}
		invoiceID = parsed
	}

	var req usecase.MemoCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), invoiceID, req, caller, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *MemoHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := h.uc.List(c.Request.Context(), page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}
