// Package handler provides the HTTP layer for the Billing bounded context.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/internal/billing/usecase"
)

// InvoiceHandler handles /api/v1/invoices/* endpoints.
type InvoiceHandler struct {
	uc      *usecase.InvoiceUseCase
	genUC   *usecase.GenerateUseCase
}

// NewInvoiceHandler constructs an InvoiceHandler.
func NewInvoiceHandler(uc *usecase.InvoiceUseCase, genUC *usecase.GenerateUseCase) *InvoiceHandler {
	return &InvoiceHandler{uc: uc, genUC: genUC}
}

// GenerateFromEngagement handles POST /invoices/generate-from-engagement.
func (h *InvoiceHandler) GenerateFromEngagement(c *gin.Context) {
	var req usecase.GenerateFromEngagementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.genUC.GenerateFromEngagement(c.Request.Context(), req, caller, c.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvoiceNumberDuplicate):
			c.JSON(http.StatusConflict, errResp("INVOICE_NUMBER_DUPLICATE", "Invoice number already exists"))
		default:
			mapInvoiceErr(c, err)
		}
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ApprovalQueue returns invoices in SENT or CONFIRMED status awaiting action.
func (h *InvoiceHandler) ApprovalQueue(c *gin.Context) {
	var req struct {
		Page int `form:"page"`
		Size int `form:"size"`
	}
	_ = c.ShouldBindQuery(&req)
	result, err := h.uc.ApprovalQueue(c.Request.Context(), req.Page, req.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *InvoiceHandler) List(c *gin.Context) {
	var req usecase.InvoiceListRequest
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

func (h *InvoiceHandler) Create(c *gin.Context) {
	var req usecase.InvoiceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), req, caller, c.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvoiceNumberDuplicate):
			c.JSON(http.StatusConflict, errResp("INVOICE_NUMBER_DUPLICATE", "Invoice number already exists"))
		default:
			c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		}
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *InvoiceHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	resp, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrInvoiceNotFound) {
			c.JSON(http.StatusNotFound, errResp("INVOICE_NOT_FOUND", "Invoice not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvoiceHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	var req usecase.InvoiceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Update(c.Request.Context(), id, req, caller, c.ClientIP())
	if err != nil {
		mapInvoiceErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvoiceHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id, caller, c.ClientIP()); err != nil {
		mapInvoiceErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *InvoiceHandler) Send(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Send(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		mapInvoiceErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvoiceHandler) Confirm(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Confirm(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		mapInvoiceErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvoiceHandler) Issue(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Issue(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		mapInvoiceErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InvoiceHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Cancel(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		mapInvoiceErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ── Line Items ────────────────────────────────────────────────────────────────

func (h *InvoiceHandler) ListLineItems(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	items, err := h.uc.ListLineItems(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *InvoiceHandler) AddLineItem(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	var req usecase.LineItemAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.AddLineItem(c.Request.Context(), id, req, caller, c.ClientIP())
	if err != nil {
		mapInvoiceErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *InvoiceHandler) DeleteLineItem(c *gin.Context) {
	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid item ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	if err := h.uc.DeleteLineItem(c.Request.Context(), invoiceID, itemID, caller, c.ClientIP()); err != nil {
		mapInvoiceErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func mapInvoiceErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvoiceNotFound):
		c.JSON(http.StatusNotFound, errResp("INVOICE_NOT_FOUND", "Invoice not found"))
	case errors.Is(err, domain.ErrInvoiceLocked):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVOICE_LOCKED", "Invoice cannot be modified in current state"))
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_STATE_TRANSITION", "Invalid state transition"))
	case errors.Is(err, domain.ErrLineItemNotFound):
		c.JSON(http.StatusNotFound, errResp("LINE_ITEM_NOT_FOUND", "Line item not found"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
