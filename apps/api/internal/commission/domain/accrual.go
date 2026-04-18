package domain

import "github.com/google/uuid"

// InvoiceAccrualData carries the billing data needed to compute commission on invoice issuance.
type InvoiceAccrualData struct {
	InvoiceID    uuid.UUID
	EngagementID uuid.UUID
	TotalAmount  int64 // VND
}

// PaymentAccrualData carries the billing data needed to compute commission on payment receipt.
type PaymentAccrualData struct {
	PaymentID    uuid.UUID
	InvoiceID    uuid.UUID
	EngagementID uuid.UUID
	Amount       int64 // VND received
}
