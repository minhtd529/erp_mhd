package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdh/erp-audit/api/internal/billing/usecase"
)

// ARHandler handles accounts receivable endpoints.
type ARHandler struct {
	uc *usecase.ARUseCase
}

// NewARHandler constructs an ARHandler.
func NewARHandler(uc *usecase.ARUseCase) *ARHandler {
	return &ARHandler{uc: uc}
}

func (h *ARHandler) Aging(c *gin.Context) {
	result, err := h.uc.GetAging(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *ARHandler) Outstanding(c *gin.Context) {
	result, err := h.uc.GetOutstanding(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
