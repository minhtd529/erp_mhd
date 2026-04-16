// Package handler provides the HTTP layer for the CRM bounded context.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
	"github.com/mdh/erp-audit/api/internal/crm/usecase"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// ClientHandler handles /api/v1/clients/* endpoints.
type ClientHandler struct {
	uc *usecase.ClientUseCase
}

// NewClientHandler constructs a ClientHandler.
func NewClientHandler(uc *usecase.ClientUseCase) *ClientHandler {
	return &ClientHandler{uc: uc}
}

// List handles GET /api/v1/clients.
func (h *ClientHandler) List(c *gin.Context) {
	var req usecase.ClientListRequest
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

// Create handles POST /api/v1/clients.
func (h *ClientHandler) Create(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	var req usecase.ClientCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.Create(c.Request.Context(), req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetByID handles GET /api/v1/clients/:id.
func (h *ClientHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client ID"))
		return
	}

	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Update handles PUT /api/v1/clients/:id.
func (h *ClientHandler) Update(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client ID"))
		return
	}

	var req usecase.ClientUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.Update(c.Request.Context(), id, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Delete handles DELETE /api/v1/clients/:id.
func (h *ClientHandler) Delete(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client ID"))
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id, caller, c.ClientIP()); err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (h *ClientHandler) handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrClientNotFound):
		c.JSON(http.StatusNotFound, errResp("CLIENT_NOT_FOUND", "Client not found"))
	case errors.Is(err, domain.ErrDuplicateTaxCode):
		c.JSON(http.StatusConflict, errResp("DUPLICATE_TAX_CODE", "Tax code already registered"))
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_STATE_TRANSITION", "Invalid state transition"))
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

// mustCallerID returns the authenticated user ID or writes 401 and returns false.
func mustCallerID(c *gin.Context) (uuid.UUID, bool) {
	id := callerID(c)
	if id == nil {
		c.JSON(http.StatusUnauthorized, errResp("UNAUTHORIZED", "Authentication required"))
		return uuid.UUID{}, false
	}
	return *id, true
}
