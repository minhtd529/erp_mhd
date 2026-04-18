package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/tax/domain"
	"github.com/mdh/erp-audit/api/internal/tax/usecase"
)

type AdvisoryHandler struct {
	uc *usecase.AdvisoryUseCase
}

func NewAdvisoryHandler(uc *usecase.AdvisoryUseCase) *AdvisoryHandler {
	return &AdvisoryHandler{uc: uc}
}

// List handles GET /clients/:client_id/advisory-records
func (h *AdvisoryHandler) List(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client_id"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := domain.AdvisoryStatus(c.Query("status"))

	result, err := h.uc.List(c.Request.Context(), clientID, status, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

// Create handles POST /clients/:client_id/advisory-records
func (h *AdvisoryHandler) Create(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	clientID, err := uuid.Parse(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client_id"))
		return
	}
	var req usecase.CreateAdvisoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	a, err := h.uc.Create(c.Request.Context(), clientID, req, callerID, clientIP(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusCreated, a)
}

// GetByID handles GET /advisory-records/:id
func (h *AdvisoryHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	a, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(mapAdvisoryErr(err))
		return
	}
	c.JSON(http.StatusOK, a)
}

// Update handles PUT /advisory-records/:id
func (h *AdvisoryHandler) Update(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	var req usecase.UpdateAdvisoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	a, err := h.uc.Update(c.Request.Context(), id, req, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapAdvisoryErr(err))
		return
	}
	c.JSON(http.StatusOK, a)
}

// Deliver handles POST /advisory-records/:id/deliver
func (h *AdvisoryHandler) Deliver(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	a, err := h.uc.Deliver(c.Request.Context(), id, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapAdvisoryErr(err))
		return
	}
	c.JSON(http.StatusOK, a)
}

// AttachFile handles POST /advisory-records/:id/attach-file
func (h *AdvisoryHandler) AttachFile(c *gin.Context) {
	callerID, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	var req usecase.AttachFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	f, err := h.uc.AttachFile(c.Request.Context(), id, req, callerID, clientIP(c))
	if err != nil {
		c.JSON(mapAdvisoryErr(err))
		return
	}
	c.JSON(http.StatusCreated, f)
}

// ListFiles handles GET /advisory-records/:id/files
func (h *AdvisoryHandler) ListFiles(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid ID"))
		return
	}
	files, err := h.uc.ListFiles(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, files)
}

func mapAdvisoryErr(err error) (int, gin.H) {
	switch {
	case errors.Is(err, domain.ErrAdvisoryRecordNotFound):
		return http.StatusNotFound, errResp("ADVISORY_RECORD_NOT_FOUND", "Advisory record not found")
	case errors.Is(err, domain.ErrAdvisoryNotDeliverable):
		return http.StatusUnprocessableEntity, errResp("ADVISORY_NOT_DELIVERABLE", "Only drafted records can be delivered")
	default:
		return http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred")
	}
}
