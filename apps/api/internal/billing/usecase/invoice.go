// Package usecase implements the Billing application layer.
package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/outbox"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// InvoiceUseCase handles invoice lifecycle operations.
type InvoiceUseCase struct {
	invoiceRepo  domain.InvoiceRepository
	lineRepo     domain.LineItemRepository
	auditLog     *audit.Logger
	publisher    *outbox.Publisher
}

// NewInvoiceUseCase constructs an InvoiceUseCase.
func NewInvoiceUseCase(
	invoiceRepo domain.InvoiceRepository,
	lineRepo domain.LineItemRepository,
	auditLog *audit.Logger,
	publisher *outbox.Publisher,
) *InvoiceUseCase {
	return &InvoiceUseCase{
		invoiceRepo: invoiceRepo,
		lineRepo:    lineRepo,
		auditLog:    auditLog,
		publisher:   publisher,
	}
}

func (uc *InvoiceUseCase) Create(ctx context.Context, req InvoiceCreateRequest, callerID uuid.UUID, ip string) (*InvoiceResponse, error) {
	inv, err := uc.invoiceRepo.Create(ctx, domain.CreateInvoiceParams{
		InvoiceNumber: req.InvoiceNumber,
		ClientID:      req.ClientID,
		EngagementID:  req.EngagementID,
		InvoiceType:   req.InvoiceType,
		IssueDate:     req.IssueDate,
		DueDate:       req.DueDate,
		TotalAmount:   req.TotalAmount,
		TaxAmount:     req.TaxAmount,
		Notes:         req.Notes,
		CreatedBy:     callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "invoices",
		ResourceID: &inv.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toInvoiceResponse(inv)
	return &resp, nil
}

func (uc *InvoiceUseCase) GetByID(ctx context.Context, id uuid.UUID) (*InvoiceResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toInvoiceResponse(inv)
	return &resp, nil
}

func (uc *InvoiceUseCase) Update(ctx context.Context, id uuid.UUID, req InvoiceUpdateRequest, callerID uuid.UUID, ip string) (*InvoiceResponse, error) {
	inv, err := uc.invoiceRepo.Update(ctx, domain.UpdateInvoiceParams{
		ID:          id,
		IssueDate:   req.IssueDate,
		DueDate:     req.DueDate,
		TotalAmount: req.TotalAmount,
		TaxAmount:   req.TaxAmount,
		Notes:       req.Notes,
		UpdatedBy:   callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "invoices",
		ResourceID: &id, Action: "UPDATE", IPAddress: ip,
	})

	resp := toInvoiceResponse(inv)
	return &resp, nil
}

func (uc *InvoiceUseCase) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) error {
	if err := uc.invoiceRepo.SoftDelete(ctx, id, callerID); err != nil {
		return err
	}
	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "invoices",
		ResourceID: &id, Action: "DELETE", IPAddress: ip,
	})
	return nil
}

// Send transitions DRAFT → SENT.
func (uc *InvoiceUseCase) Send(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*InvoiceResponse, error) {
	return uc.transition(ctx, id, domain.InvoiceStatusSent, callerID, ip)
}

// Confirm transitions SENT → CONFIRMED.
func (uc *InvoiceUseCase) Confirm(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*InvoiceResponse, error) {
	return uc.transition(ctx, id, domain.InvoiceStatusConfirmed, callerID, ip)
}

// Issue transitions CONFIRMED → ISSUED, freezes snapshot, publishes outbox event.
func (uc *InvoiceUseCase) Issue(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) (*InvoiceResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if inv.Status != domain.InvoiceStatusConfirmed && inv.Status != domain.InvoiceStatusSent {
		return nil, domain.ErrInvalidStateTransition
	}

	// Freeze snapshot with timestamp
	snapshot, _ := json.Marshal(map[string]any{
		"invoice_number": inv.InvoiceNumber,
		"total_amount":   inv.TotalAmount,
		"tax_amount":     inv.TaxAmount,
		"issued_at":      time.Now().UTC(),
	})
	now := time.Now()
	inv.IssueDate = &now

	if _, err := uc.invoiceRepo.UpdateSnapshot(ctx, id, snapshot, callerID); err != nil {
		return nil, err
	}
	updated, err := uc.invoiceRepo.UpdateStatus(ctx, id, domain.InvoiceStatusIssued, callerID)
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "invoices",
		ResourceID: &id, Action: "STATE_TRANSITION", IPAddress: ip,
		NewValue: map[string]string{"status": "ISSUED"},
	})

	if uc.publisher != nil {
		_ = uc.publisher.Publish(ctx, "invoice", id, outbox.EventType("invoice.issued"), map[string]string{
			"invoice_id": id.String(), "client_id": inv.ClientID.String(),
		})
	}

	resp := toInvoiceResponse(updated)
	return &resp, nil
}

func (uc *InvoiceUseCase) transition(ctx context.Context, id uuid.UUID, newStatus domain.InvoiceStatus, callerID uuid.UUID, ip string) (*InvoiceResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !validTransition(inv.Status, newStatus) {
		return nil, domain.ErrInvalidStateTransition
	}
	updated, err := uc.invoiceRepo.UpdateStatus(ctx, id, newStatus, callerID)
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "invoices",
		ResourceID: &id, Action: "STATE_TRANSITION", IPAddress: ip,
		NewValue: map[string]string{"status": string(newStatus)},
	})

	resp := toInvoiceResponse(updated)
	return &resp, nil
}

func validTransition(current, next domain.InvoiceStatus) bool {
	allowed := map[domain.InvoiceStatus][]domain.InvoiceStatus{
		domain.InvoiceStatusDraft:     {domain.InvoiceStatusSent, domain.InvoiceStatusCancelled},
		domain.InvoiceStatusSent:      {domain.InvoiceStatusConfirmed, domain.InvoiceStatusCancelled},
		domain.InvoiceStatusConfirmed: {domain.InvoiceStatusIssued, domain.InvoiceStatusCancelled},
		domain.InvoiceStatusIssued:    {domain.InvoiceStatusPaid},
	}
	for _, s := range allowed[current] {
		if s == next {
			return true
		}
	}
	return false
}

func (uc *InvoiceUseCase) List(ctx context.Context, req InvoiceListRequest) (*PaginatedResult[InvoiceResponse], error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}
	invoices, total, err := uc.invoiceRepo.List(ctx, domain.ListInvoicesFilter{
		Page:         req.Page,
		Size:         req.Size,
		ClientID:     req.ClientID,
		EngagementID: req.EngagementID,
		Status:       req.Status,
		Q:            req.Q,
	})
	if err != nil {
		return nil, err
	}
	data := make([]InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		data[i] = toInvoiceResponse(inv)
	}
	result := pagination.NewOffsetResult(data, total, req.Page, req.Size)
	return &result, nil
}

// ApprovalQueue returns invoices in SENT or CONFIRMED status — i.e., those waiting for a human action.
func (uc *InvoiceUseCase) ApprovalQueue(ctx context.Context, page, size int) (*PaginatedResult[InvoiceResponse], error) {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 20
	}
	invoices, total, err := uc.invoiceRepo.List(ctx, domain.ListInvoicesFilter{
		Page:     page,
		Size:     size,
		Statuses: []domain.InvoiceStatus{domain.InvoiceStatusSent, domain.InvoiceStatusConfirmed},
	})
	if err != nil {
		return nil, err
	}
	data := make([]InvoiceResponse, len(invoices))
	for i, inv := range invoices {
		data[i] = toInvoiceResponse(inv)
	}
	result := pagination.NewOffsetResult(data, total, page, size)
	return &result, nil
}

// ── Line Items ────────────────────────────────────────────────────────────────

func (uc *InvoiceUseCase) AddLineItem(ctx context.Context, invoiceID uuid.UUID, req LineItemAddRequest, callerID uuid.UUID, ip string) (*LineItemResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if inv.Status != domain.InvoiceStatusDraft {
		return nil, domain.ErrInvoiceLocked
	}

	totalAmount := req.Quantity * req.UnitPrice
	srcType := req.SourceType
	if srcType == "" {
		srcType = domain.SourceManual
	}
	item, err := uc.lineRepo.Add(ctx, domain.AddLineItemParams{
		InvoiceID:   invoiceID,
		Description: req.Description,
		Quantity:    req.Quantity,
		UnitPrice:   req.UnitPrice,
		TaxAmount:   req.TaxAmount,
		TotalAmount: totalAmount,
		SourceType:  srcType,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "invoice_line_items",
		ResourceID: &item.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toLineItemResponse(item)
	return &resp, nil
}

func (uc *InvoiceUseCase) ListLineItems(ctx context.Context, invoiceID uuid.UUID) ([]LineItemResponse, error) {
	items, err := uc.lineRepo.ListByInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	data := make([]LineItemResponse, len(items))
	for i, item := range items {
		data[i] = toLineItemResponse(item)
	}
	return data, nil
}

func (uc *InvoiceUseCase) DeleteLineItem(ctx context.Context, invoiceID, itemID uuid.UUID, callerID uuid.UUID, ip string) error {
	inv, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return err
	}
	if inv.Status != domain.InvoiceStatusDraft {
		return domain.ErrInvoiceLocked
	}
	if err := uc.lineRepo.Delete(ctx, itemID); err != nil {
		return err
	}
	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "invoice_line_items",
		ResourceID: &itemID, Action: "DELETE", IPAddress: ip,
	})
	return nil
}
