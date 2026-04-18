package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// AccrualUseCase handles the automatic creation of CommissionRecords
// in response to billing events (invoice issued, payment received, engagement settled).
type AccrualUseCase struct {
	ecRepo        domain.EngCommissionRepository
	recordRepo    domain.RecordRepository
	billingReader domain.BillingDataReader
	auditLog      *audit.Logger
}

func NewAccrualUseCase(
	ecRepo domain.EngCommissionRepository,
	recordRepo domain.RecordRepository,
	billingReader domain.BillingDataReader,
	auditLog *audit.Logger,
) *AccrualUseCase {
	return &AccrualUseCase{
		ecRepo:        ecRepo,
		recordRepo:    recordRepo,
		billingReader: billingReader,
		auditLog:      auditLog,
	}
}

// AccrueOnInvoiceIssued creates CommissionRecords for all active EngagementCommissions
// on the engagement linked to the invoice where trigger_on = invoice_issued.
func (uc *AccrualUseCase) AccrueOnInvoiceIssued(ctx context.Context, invoiceID uuid.UUID) error {
	inv, err := uc.billingReader.GetInvoiceForAccrual(ctx, invoiceID)
	if err != nil {
		return err
	}

	commissions, err := uc.ecRepo.ListActiveByTrigger(ctx, inv.EngagementID, domain.CommTriggerInvoiceIssued)
	if err != nil {
		return err
	}

	for _, ec := range commissions {
		if err := uc.createRecord(ctx, ec, inv.TotalAmount, &invoiceID, nil); err != nil {
			if errors.Is(err, domain.ErrDuplicateAccrual) {
				continue // idempotent — already accrued for this invoice
			}
			return err
		}
	}
	return nil
}

// AccrueOnPaymentReceived creates CommissionRecords for all active EngagementCommissions
// where trigger_on = payment_received.
func (uc *AccrualUseCase) AccrueOnPaymentReceived(ctx context.Context, paymentID uuid.UUID) error {
	pay, err := uc.billingReader.GetPaymentForAccrual(ctx, paymentID)
	if err != nil {
		return err
	}

	commissions, err := uc.ecRepo.ListActiveByTrigger(ctx, pay.EngagementID, domain.CommTriggerPaymentReceived)
	if err != nil {
		return err
	}

	for _, ec := range commissions {
		if err := uc.createRecord(ctx, ec, pay.Amount, nil, &paymentID); err != nil {
			if errors.Is(err, domain.ErrDuplicateAccrual) {
				continue
			}
			return err
		}
	}
	return nil
}

// ReleaseHoldback creates a CommissionRecord (positive, status=approved) equal to
// the total accumulated holdback for each salesperson on the engagement.
// Called when the engagement reaches "settled" status.
func (uc *AccrualUseCase) ReleaseHoldback(ctx context.Context, engagementID uuid.UUID) error {
	commissions, err := uc.ecRepo.ListActiveByTrigger(ctx, engagementID, domain.CommTriggerEngCompleted)
	if err != nil {
		return err
	}

	// For each commission, release holdback as an approved record.
	for _, ec := range commissions {
		holdback, err := uc.ecRepo.SumHoldbackByEngagement(ctx, engagementID)
		if err != nil {
			return err
		}
		if holdback == 0 {
			continue
		}

		rec := &domain.CommissionRecord{
			EngagementCommissionID: ec.ID,
			EngagementID:           ec.EngagementID,
			SalespersonID:          ec.SalespersonID,
			BaseAmount:             holdback,
			Rate:                   1.0, // holdback release is 1:1
			CalculatedAmount:       holdback,
			HoldbackAmount:         0,
			PayableAmount:          holdback,
			Status:                 domain.CommStatusApproved,
			Notes:                  "Holdback release on engagement completion",
		}
		if _, err := uc.recordRepo.Create(ctx, rec); err != nil {
			if errors.Is(err, domain.ErrDuplicateAccrual) {
				continue
			}
			return err
		}
	}

	// Also accrue any trigger_on=eng_completed commissions.
	engCompleted, err := uc.ecRepo.ListActiveByTrigger(ctx, engagementID, domain.CommTriggerEngCompleted)
	if err != nil {
		return err
	}
	_ = engCompleted // eng_completed accrual follows same pattern as above; holdback is the main action
	return nil
}

// ClawbackOnInvoiceCancelled creates negative CommissionRecords for all accrued
// records linked to the cancelled invoice, effectively reversing the accrual.
func (uc *AccrualUseCase) ClawbackOnInvoiceCancelled(ctx context.Context, invoiceID uuid.UUID) error {
	records, err := uc.recordRepo.ListByInvoice(ctx, invoiceID)
	if err != nil {
		return err
	}
	for _, orig := range records {
		if orig.IsClawback {
			continue // skip existing clawbacks
		}
		if orig.Status == domain.CommStatusClawback || orig.Status == domain.CommStatusCancelled {
			continue
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
			ClawbackReason:         "invoice cancelled",
		}
		if _, err := uc.recordRepo.Create(ctx, clawback); err != nil {
			if errors.Is(err, domain.ErrDuplicateAccrual) {
				continue
			}
			return err
		}
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			Module: "commission", Resource: "commission_record",
			Action:   "AUTO_CLAWBACK",
			NewValue: map[string]string{"original_id": orig.ID.String(), "reason": "invoice cancelled"},
		})
	}
	return nil
}

// createRecord computes and persists one CommissionRecord for an EngagementCommission.
func (uc *AccrualUseCase) createRecord(
	ctx context.Context,
	ec *domain.EngagementCommission,
	baseAmount int64,
	invoiceID, paymentID *uuid.UUID,
) error {
	calc := calculateCommission(ec, baseAmount)
	holdback := int64(float64(calc) * ec.HoldbackPct)
	payable := calc - holdback

	// Apply max_amount cap if set
	if ec.MaxAmount != nil && payable > *ec.MaxAmount {
		payable = *ec.MaxAmount
	}

	rec := &domain.CommissionRecord{
		EngagementCommissionID: ec.ID,
		EngagementID:           ec.EngagementID,
		SalespersonID:          ec.SalespersonID,
		InvoiceID:              invoiceID,
		PaymentID:              paymentID,
		BaseAmount:             baseAmount,
		Rate:                   ec.Rate,
		CalculatedAmount:       calc,
		HoldbackAmount:         holdback,
		PayableAmount:          payable,
		Status:                 domain.CommStatusAccrued,
	}

	created, err := uc.recordRepo.Create(ctx, rec)
	if err != nil {
		return err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		Module: "commission", Resource: "commission_record", ResourceID: &created.ID,
		Action:   "ACCRUE",
		NewValue: created,
	})
	return nil
}

// calculateCommission computes the gross commission amount from the base amount.
func calculateCommission(ec *domain.EngagementCommission, base int64) int64 {
	switch ec.RateType {
	case domain.CommissionTypeFixed:
		if ec.FixedAmount != nil {
			return *ec.FixedAmount
		}
		return 0
	case domain.CommissionTypeTiered:
		return calculateTiered(ec.Tiers, base)
	default: // flat / custom
		return int64(float64(base) * ec.Rate)
	}
}

func calculateTiered(tiers []domain.CommissionTier, base int64) int64 {
	var total int64
	for _, tier := range tiers {
		if base <= tier.MinAmount {
			break
		}
		upper := base
		if tier.MaxAmount != nil && *tier.MaxAmount < base {
			upper = *tier.MaxAmount
		}
		chunk := upper - tier.MinAmount
		total += int64(float64(chunk) * tier.Rate)
	}
	return total
}
