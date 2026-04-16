package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
	"github.com/mdh/erp-audit/api/internal/crm/usecase"
)

// ContactHandler handles /api/v1/clients/:id/contacts/* endpoints.
type ContactHandler struct {
	uc *usecase.ContactUseCase
}

// NewContactHandler constructs a ContactHandler.
func NewContactHandler(uc *usecase.ContactUseCase) *ContactHandler {
	return &ContactHandler{uc: uc}
}

// List handles GET /api/v1/clients/:id/contacts.
func (h *ContactHandler) List(c *gin.Context) {
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client ID"))
		return
	}

	items, err := h.uc.ListByClient(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, items)
}

// Create handles POST /api/v1/clients/:id/contacts.
func (h *ContactHandler) Create(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client ID"))
		return
	}

	var req usecase.ContactCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.Create(c.Request.Context(), clientID, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// Update handles PUT /api/v1/clients/:id/contacts/:cid.
func (h *ContactHandler) Update(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client ID"))
		return
	}
	contactID, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid contact ID"))
		return
	}

	var req usecase.ContactUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}

	resp, err := h.uc.Update(c.Request.Context(), clientID, contactID, req, caller, c.ClientIP())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Delete handles DELETE /api/v1/clients/:id/contacts/:cid.
func (h *ContactHandler) Delete(c *gin.Context) {
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	clientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid client ID"))
		return
	}
	contactID, err := uuid.Parse(c.Param("cid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid contact ID"))
		return
	}

	if err := h.uc.Delete(c.Request.Context(), clientID, contactID, caller, c.ClientIP()); err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *ContactHandler) handleErr(c *gin.Context, err error) {
	if errors.Is(err, domain.ErrContactNotFound) {
		c.JSON(http.StatusNotFound, errResp("CONTACT_NOT_FOUND", "Contact not found"))
		return
	}
	c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
}
