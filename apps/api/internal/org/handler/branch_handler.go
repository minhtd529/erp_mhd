package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/org/domain"
	"github.com/mdh/erp-audit/api/internal/org/usecase"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

type BranchHandler struct {
	uc *usecase.BranchUseCase
}

func NewBranchHandler(uc *usecase.BranchUseCase) *BranchHandler {
	return &BranchHandler{uc: uc}
}

func (h *BranchHandler) List(c *gin.Context) {
	var req usecase.BranchListRequest
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

func (h *BranchHandler) Create(c *gin.Context) {
	var req usecase.BranchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), req, callerID(c), c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *BranchHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid branch ID"))
		return
	}
	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BranchHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid branch ID"))
		return
	}
	var req usecase.BranchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	resp, err := h.uc.Update(c.Request.Context(), id, req, callerID(c), c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BranchHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrBranchNotFound):
		c.JSON(http.StatusNotFound, errResp("BRANCH_NOT_FOUND", "Branch not found"))
	case errors.Is(err, domain.ErrDuplicateCode):
		c.JSON(http.StatusConflict, errResp("DUPLICATE_CODE", "Branch code already exists"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}

func errResp(code, msg string) gin.H {
	return gin.H{"error": code, "message": msg}
}

func callerID(c *gin.Context) *uuid.UUID {
	raw, ok := c.Get(pkgauth.CtxUserID)
	if !ok {
		return nil
	}
	id, ok := raw.(uuid.UUID)
	if !ok {
		return nil
	}
	return &id
}
