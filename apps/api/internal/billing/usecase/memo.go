package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/outbox"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// MemoUseCase handles billing memo (credit note / adjustment) operations.
type MemoUseCase struct {
	memoRepo    domain.MemoRepository
	invoiceRepo domain.InvoiceRepository
	auditLog    *audit.Logger
	publisher   *outbox.Publisher
}

// NewMemoUseCase constructs a MemoUseCase.
func NewMemoUseCase(
	memoRepo domain.MemoRepository,
	invoiceRepo domain.InvoiceRepository,
	auditLog *audit.Logger,
	publisher *outbox.Publisher,
) *MemoUseCase {
	return &MemoUseCase{
		memoRepo:    memoRepo,
		invoiceRepo: invoiceRepo,
		auditLog:    auditLog,
		publisher:   publisher,
	}
}

func (uc *MemoUseCase) Create(ctx context.Context, invoiceID uuid.UUID, req MemoCreateRequest, callerID uuid.UUID, ip string) (*MemoResponse, error) {
	// For credit notes, verify the related invoice exists
	if req.MemoType == domain.MemoCreditNote && req.RelatedInvoiceID != nil {
		if _, err := uc.invoiceRepo.FindByID(ctx, *req.RelatedInvoiceID); err != nil {
			return nil, err
		}
	}

	relatedID := req.RelatedInvoiceID
	if relatedID == nil && invoiceID != uuid.Nil {
		relatedID = &invoiceID
	}

	memo, err := uc.memoRepo.Create(ctx, domain.CreateMemoParams{
		RelatedInvoiceID: relatedID,
		MemoType:         req.MemoType,
		MemoNumber:       req.MemoNumber,
		Amount:           req.Amount,
		Reason:           req.Reason,
		CreatedBy:        callerID,
	})
	if err != nil {
		return nil, err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "billing_memos",
		ResourceID: &memo.ID, Action: "CREATE", IPAddress: ip,
	})

	if uc.publisher != nil && req.MemoType == domain.MemoCreditNote {
		_ = uc.publisher.Publish(ctx, "billing_memo", memo.ID, outbox.EventType("credit_note.issued"), map[string]string{
			"memo_id": memo.ID.String(),
		})
	}

	resp := toMemoResponse(memo)
	return &resp, nil
}

func (uc *MemoUseCase) List(ctx context.Context, page, size int) (*PaginatedResult[MemoResponse], error) {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 20
	}
	memos, total, err := uc.memoRepo.List(ctx, page, size)
	if err != nil {
		return nil, err
	}
	data := make([]MemoResponse, len(memos))
	for i, m := range memos {
		data[i] = toMemoResponse(m)
	}
	result := pagination.NewOffsetResult(data, total, page, size)
	return &result, nil
}
