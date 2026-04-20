package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/internal/billing/usecase"
)

// PaymentHandler handles payment endpoints.
type PaymentHandler struct {
	uc *usecase.PaymentUseCase
}

// NewPaymentHandler constructs a PaymentHandler.
func NewPaymentHandler(uc *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) List(c *gin.Context) {
	page, size := parsePage(c)
	result, err := h.uc.ListAll(c.Request.Context(), page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *PaymentHandler) ListByInvoice(c *gin.Context) {
	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	payments, err := h.uc.ListByInvoice(c.Request.Context(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
		return
	}
	c.JSON(http.StatusOK, payments)
}

func (h *PaymentHandler) Record(c *gin.Context) {
	invoiceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid invoice ID"))
		return
	}
	var req usecase.PaymentRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errResp("VALIDATION_ERROR", err.Error()))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.Record(c.Request.Context(), invoiceID, req, caller, c.ClientIP())
	if err != nil {
		mapPaymentErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *PaymentHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid payment ID"))
		return
	}
	var req usecase.PaymentUpdateRequest
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
		mapPaymentErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PaymentHandler) Reverse(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid payment ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	if err := h.uc.Reverse(c.Request.Context(), id, caller, c.ClientIP()); err != nil {
		mapPaymentErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PaymentHandler) Clear(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid payment ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.ClearPayment(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		mapPaymentErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PaymentHandler) Dispute(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("INVALID_ID", "Invalid payment ID"))
		return
	}
	caller, ok := mustCallerID(c)
	if !ok {
		return
	}
	resp, err := h.uc.DisputePayment(c.Request.Context(), id, caller, c.ClientIP())
	if err != nil {
		mapPaymentErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func mapPaymentErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrPaymentNotFound):
		c.JSON(http.StatusNotFound, errResp("PAYMENT_NOT_FOUND", "Payment not found"))
	case errors.Is(err, domain.ErrPaymentExceedsBalance):
		c.JSON(http.StatusUnprocessableEntity, errResp("PAYMENT_EXCEEDS_BALANCE", "Payment amount exceeds outstanding balance"))
	case errors.Is(err, domain.ErrPaymentNotRecorded):
		c.JSON(http.StatusUnprocessableEntity, errResp("PAYMENT_NOT_RECORDED", "Payment cannot be reversed in current state"))
	case errors.Is(err, domain.ErrPaymentNotCleared):
		c.JSON(http.StatusUnprocessableEntity, errResp("PAYMENT_NOT_CLEARED", "Payment must be CLEARED before disputing"))
	case errors.Is(err, domain.ErrPaymentAlreadyCleared):
		c.JSON(http.StatusUnprocessableEntity, errResp("PAYMENT_ALREADY_CLEARED", "Payment is already cleared"))
	case errors.Is(err, domain.ErrInvalidStateTransition):
		c.JSON(http.StatusUnprocessableEntity, errResp("INVALID_STATE_TRANSITION", "Invoice not in a billable state"))
	case errors.Is(err, domain.ErrInvoiceNotFound):
		c.JSON(http.StatusNotFound, errResp("INVOICE_NOT_FOUND", "Invoice not found"))
	default:
		c.JSON(http.StatusInternalServerError, errResp("INTERNAL_ERROR", "An internal error occurred"))
	}
}
