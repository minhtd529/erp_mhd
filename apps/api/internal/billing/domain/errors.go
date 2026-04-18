package domain

import "errors"

var (
	ErrInvoiceNotFound              = errors.New("INVOICE_NOT_FOUND")
	ErrInvoiceLocked                = errors.New("INVOICE_LOCKED")
	ErrInvoiceNotDraft              = errors.New("INVOICE_NOT_DRAFT")
	ErrInvoiceNumberDuplicate       = errors.New("INVOICE_NUMBER_DUPLICATE")
	ErrInvalidStateTransition       = errors.New("INVALID_STATE_TRANSITION")
	ErrPaymentExceedsBalance        = errors.New("PAYMENT_EXCEEDS_BALANCE")
	ErrPaymentNotFound              = errors.New("PAYMENT_NOT_FOUND")
	ErrPaymentNotRecorded           = errors.New("PAYMENT_NOT_RECORDED")
	ErrMemoNotFound                 = errors.New("MEMO_NOT_FOUND")
	ErrLineItemNotFound             = errors.New("LINE_ITEM_NOT_FOUND")
	ErrEngagementNotApprovedForBill = errors.New("ENGAGEMENT_NOT_APPROVED_FOR_BILLING")
	ErrPaymentAlreadyCleared        = errors.New("PAYMENT_ALREADY_CLEARED")
	ErrPaymentNotCleared            = errors.New("PAYMENT_NOT_CLEARED")
)
