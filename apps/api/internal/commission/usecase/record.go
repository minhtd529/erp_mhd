package usecase

import (
	"context"
	"net"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// RecordUseCase handles commission record lifecycle: approve, pay, clawback, queries.
type RecordUseCase struct {
	recordRepo domain.RecordRepository
	auditLog   *audit.Logger
}

func NewRecordUseCase(recordRepo domain.RecordRepository, auditLog *audit.Logger) *RecordUseCase {
	return &RecordUseCase{recordRepo: recordRepo, auditLog: auditLog}
}

func (uc *RecordUseCase) List(ctx context.Context, f domain.ListRecordsFilter, page, size int) (pagination.OffsetResult[RecordResponse], error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	records, total, err := uc.recordRepo.List(ctx, f, page, size)
	if err != nil {
		return pagination.OffsetResult[RecordResponse]{}, err
	}
	items := make([]RecordResponse, len(records))
	for i, r := range records {
		items[i] = toRecordResponse(r)
	}
	return pagination.NewOffsetResult(items, total, page, size), nil
}

func (uc *RecordUseCase) Approve(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip net.IP) (*RecordResponse, error) {
	rec, err := uc.recordRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec.Status != domain.CommStatusAccrued {
		return nil, domain.ErrRecordNotApprovable
	}

	updated, err := uc.recordRepo.UpdateStatus(ctx, id, domain.CommStatusApproved, &callerID, "")
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "commission_record", ResourceID: &updated.ID,
		Action: "APPROVE", NewValue: map[string]string{"status": "approved"},
	})
	r := toRecordResponse(updated)
	return &r, nil
}

func (uc *RecordUseCase) MarkPaid(ctx context.Context, id uuid.UUID, payoutRef string, callerID uuid.UUID, ip net.IP) (*RecordResponse, error) {
	rec, err := uc.recordRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rec.Status != domain.CommStatusApproved {
		return nil, domain.ErrRecordNotPayable
	}

	updated, err := uc.recordRepo.UpdateStatus(ctx, id, domain.CommStatusPaid, &callerID, payoutRef)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "commission_record", ResourceID: &updated.ID,
		Action: "MARK_PAID", NewValue: map[string]string{"status": "paid", "payout_reference": payoutRef},
	})
	r := toRecordResponse(updated)
	return &r, nil
}

type BulkApproveRequest struct {
	IDs []uuid.UUID `json:"ids" binding:"required"`
}

type BulkPayRequest struct {
	IDs            []uuid.UUID `json:"ids" binding:"required"`
	PayoutReference string     `json:"payout_reference"`
}

type BulkResult struct {
	AffectedCount int64 `json:"affected_count"`
}

func (uc *RecordUseCase) BulkApprove(ctx context.Context, req BulkApproveRequest, callerID uuid.UUID, ip net.IP) (*BulkResult, error) {
	n, err := uc.recordRepo.BulkUpdateStatus(ctx, req.IDs, domain.CommStatusApproved, &callerID, "")
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "commission_record",
		Action: "BULK_APPROVE", NewValue: map[string]any{"count": n},
	})
	return &BulkResult{AffectedCount: n}, nil
}

func (uc *RecordUseCase) BulkPay(ctx context.Context, req BulkPayRequest, callerID uuid.UUID, ip net.IP) (*BulkResult, error) {
	n, err := uc.recordRepo.BulkUpdateStatus(ctx, req.IDs, domain.CommStatusPaid, &callerID, req.PayoutReference)
	if err != nil {
		return nil, err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "commission_record",
		Action: "BULK_PAY", NewValue: map[string]any{"count": n, "payout_reference": req.PayoutReference},
	})
	return &BulkResult{AffectedCount: n}, nil
}

// Clawback creates a negative CommissionRecord linked to an existing record.
func (uc *RecordUseCase) Clawback(ctx context.Context, id uuid.UUID, reason string, callerID uuid.UUID, ip net.IP) (*RecordResponse, error) {
	orig, err := uc.recordRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if orig.Status == domain.CommStatusApproved || orig.Status == domain.CommStatusPaid {
		// Immutable — only clawback is allowed, not edit
	}

	clawback := &domain.CommissionRecord{
		EngagementCommissionID: orig.EngagementCommissionID,
		EngagementID:           orig.EngagementID,
		SalespersonID:          orig.SalespersonID,
		InvoiceID:              orig.InvoiceID,
		PaymentID:              orig.PaymentID,
		BaseAmount:             -orig.BaseAmount,
		Rate:                   orig.Rate,
		CalculatedAmount:       -orig.CalculatedAmount,
		HoldbackAmount:         -orig.HoldbackAmount,
		PayableAmount:          -orig.PayableAmount,
		Status:                 domain.CommStatusClawback,
		ClawbackRecordID:       &orig.ID,
		IsClawback:             true,
		ClawbackReason:         reason,
	}
	created, err := uc.recordRepo.Create(ctx, clawback)
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, IPAddress: ip.String(),
		Module: "commission", Resource: "commission_record", ResourceID: &created.ID,
		Action: "CLAWBACK", NewValue: map[string]string{"original_id": id.String(), "reason": reason},
	})
	r := toRecordResponse(created)
	return &r, nil
}

// MyCommissions returns records for the authenticated salesperson.
func (uc *RecordUseCase) MyCommissions(ctx context.Context, salespersonID uuid.UUID, status domain.CommissionStatus, page, size int) (pagination.OffsetResult[RecordResponse], error) {
	f := domain.ListRecordsFilter{SalespersonID: &salespersonID, Status: status}
	return uc.List(ctx, f, page, size)
}

// MyCommissionSummary returns aggregated commission totals for the salesperson.
func (uc *RecordUseCase) MyCommissionSummary(ctx context.Context, salespersonID uuid.UUID) (*domain.SalespersonSummary, error) {
	return uc.recordRepo.SummarySalesperson(ctx, salespersonID)
}

// TeamCommissions returns commission records for all salespersons under the manager's engagements.
func (uc *RecordUseCase) TeamCommissions(ctx context.Context, managerID uuid.UUID, page, size int) (pagination.OffsetResult[RecordResponse], error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	records, total, err := uc.recordRepo.ListByTeam(ctx, managerID, page, size)
	if err != nil {
		return pagination.OffsetResult[RecordResponse]{}, err
	}
	items := make([]RecordResponse, len(records))
	for i, r := range records {
		items[i] = toRecordResponse(r)
	}
	return pagination.NewOffsetResult(items, total, page, size), nil
}
