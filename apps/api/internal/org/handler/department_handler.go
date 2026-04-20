package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/org/domain"
	"github.com/mdh/erp-audit/api/internal/org/usecase"
)

type DepartmentHandler struct {
	uc *usecase.DepartmentUseCase
}

func NewDepartmentHandler(uc *usecase.DepartmentUseCase) *DepartmentHandler {
	return &DepartmentHandler{uc: uc}
}

func (h *DepartmentHandler) List(c *gin.Context) {
	var req usecase.DepartmentListRequest
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

func (h *DepartmentHandler) Create(c *gin.Context) {
	var req usecase.DepartmentCreateRequest
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

func (h *DepartmentHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid department ID"))
		return
	}
	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *DepartmentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid department ID"))
		return
	}
	var req usecase.DepartmentUpdateRequest
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

func (h *DepartmentHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrDepartmentNotFound):
		c.JSON(http.StatusNotFound, errResp("DEPARTMENT_NOT_FOUND", "Department not found"))
	case errors.Is(err, domain.ErrDuplicateCode):
		c.JSON(http.StatusConflict, errResp("DUPLICATE_CODE", "Department code already exists"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
