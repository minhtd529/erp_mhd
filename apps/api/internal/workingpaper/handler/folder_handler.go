package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/usecase"
)

// FolderHandler handles working paper folder endpoints.
type FolderHandler struct {
	uc *usecase.FolderUseCase
}

// NewFolderHandler constructs a FolderHandler.
func NewFolderHandler(uc *usecase.FolderUseCase) *FolderHandler {
	return &FolderHandler{uc: uc}
}

func (h *FolderHandler) Create(c *gin.Context) {
	engID, err := uuid.Parse(c.Param("engagement_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	var req usecase.FolderCreateRequest
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

func (h *FolderHandler) List(c *gin.Context) {
	engID, err := uuid.Parse(c.Param("engagement_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid engagement ID"))
		return
	}
	folders, err := h.uc.ListByEngagement(c.Request.Context(), engID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, folders)
}
