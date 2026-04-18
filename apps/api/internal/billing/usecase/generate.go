package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	engdomain "github.com/mdh/erp-audit/api/internal/engagement/domain"
	tsdomain "github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// GenerateFromEngagementRequest is the input for T&M invoice generation.
type GenerateFromEngagementRequest struct {
	EngagementID  uuid.UUID  `json:"engagement_id"  binding:"required"`
	InvoiceNumber string     `json:"invoice_number" binding:"required"`
	PeriodStart   time.Time  `json:"period_start"   binding:"required"`
	PeriodEnd     time.Time  `json:"period_end"     binding:"required"`
	DueDate       *time.Time `json:"due_date"`
	TaxRate       float64    `json:"tax_rate"` // e.g. 0.10 for 10%
}

// GenerateFromEngagementResponse carries the created invoice + line items.
type GenerateFromEngagementResponse struct {
	Invoice   InvoiceResponse    `json:"invoice"`
	LineItems []LineItemResponse `json:"line_items"`
}

// ─── Source interfaces (cross-domain reads) ───────────────────────────────────

// EngagementSource provides read-only access to engagement data needed for invoice generation.
type EngagementSource interface {
	FindByID(ctx context.Context, id uuid.UUID) (*engdomain.Engagement, error)
	ListByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*engdomain.EngagementMember, error)
	ListCostsByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*engdomain.DirectCost, error)
}

// TimesheetEntrySource provides read-only access to locked timesheet entries.
type TimesheetEntrySource interface {
	ListLockedByEngagement(ctx context.Context, engagementID uuid.UUID, start, end time.Time) ([]*tsdomain.TimesheetEntry, error)
}

// ─── GenerateUseCase ──────────────────────────────────────────────────────────

// GenerateUseCase handles invoice generation from engagement/timesheet data.
type GenerateUseCase struct {
	invoiceRepo domain.InvoiceRepository
	lineRepo    domain.LineItemRepository
	engSource   EngagementSource
	tsSource    TimesheetEntrySource
	auditLog    *audit.Logger
}

// NewGenerateUseCase constructs a GenerateUseCase.
func NewGenerateUseCase(
	invoiceRepo domain.InvoiceRepository,
	lineRepo domain.LineItemRepository,
	engSource EngagementSource,
	tsSource TimesheetEntrySource,
	auditLog *audit.Logger,
) *GenerateUseCase {
	return &GenerateUseCase{
		invoiceRepo: invoiceRepo,
		lineRepo:    lineRepo,
		engSource:   engSource,
		tsSource:    tsSource,
		auditLog:    auditLog,
	}
}

// GenerateFromEngagement creates a DRAFT invoice from locked timesheet entries
// and approved direct costs for an engagement.
func (uc *GenerateUseCase) GenerateFromEngagement(ctx context.Context, req GenerateFromEngagementRequest, callerID uuid.UUID, ip string) (*GenerateFromEngagementResponse, error) {
	// 1. Fetch engagement
	eng, err := uc.engSource.FindByID(ctx, req.EngagementID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch engagement members (for hourly rates)
	members, err := uc.engSource.ListByEngagement(ctx, req.EngagementID)
	if err != nil {
		return nil, err
	}
	rateByStaff := map[uuid.UUID]float64{}
	for _, m := range members {
		if m.HourlyRate != nil {
			rateByStaff[m.StaffID] = *m.HourlyRate
		}
	}

	// 3. Create DRAFT invoice
	now := time.Now()
	dueDate := req.DueDate
	if dueDate == nil {
		d := now.AddDate(0, 0, 30)
		dueDate = &d
	}

	inv, err := uc.invoiceRepo.Create(ctx, domain.CreateInvoiceParams{
		InvoiceNumber: req.InvoiceNumber,
		ClientID:      eng.ClientID,
		EngagementID:  &req.EngagementID,
		InvoiceType:   resolveInvoiceType(eng.FeeType),
		IssueDate:     &now,
		DueDate:       dueDate,
		Notes:         eng.Description,
		CreatedBy:     callerID,
	})
	if err != nil {
		return nil, err
	}

	var lineItems []LineItemResponse

	// 4. Generate line items based on fee type
	switch eng.FeeType {
	case engdomain.FeeTimeAndMaterial:
		items, err := uc.generateTMLineItems(ctx, inv.ID, req, rateByStaff)
		if err != nil {
			return nil, err
		}
		lineItems = items

	case engdomain.FeeFixed:
		item, err := uc.addFixedFeeLineItem(ctx, inv.ID, eng)
		if err != nil {
			return nil, err
		}
		lineItems = []LineItemResponse{item}

	default:
		// RETAINER / SUCCESS: create single line item from fee_amount
		item, err := uc.addFixedFeeLineItem(ctx, inv.ID, eng)
		if err != nil {
			return nil, err
		}
		lineItems = []LineItemResponse{item}
	}

	// 5. Add approved direct costs as line items
	costs, err := uc.engSource.ListCostsByEngagement(ctx, req.EngagementID)
	if err == nil {
		for _, cost := range costs {
			if cost.Status != engdomain.CostApproved {
				continue
			}
			snap, _ := json.Marshal(map[string]any{
				"cost_type": cost.CostType, "original_id": cost.ID.String(),
			})
			item, addErr := uc.lineRepo.Add(ctx, domain.AddLineItemParams{
				InvoiceID:    inv.ID,
				Description:  fmt.Sprintf("[%s] %s", cost.CostType, cost.Description),
				Quantity:     1,
				UnitPrice:    cost.Amount,
				TaxAmount:    0,
				TotalAmount:  cost.Amount,
				SourceType:   domain.SourceDirectCost,
				SnapshotData: snap,
			})
			if addErr == nil {
				lineItems = append(lineItems, toLineItemResponse(item))
			}
		}
	}

	// 6. Update invoice total
	var total, taxTotal float64
	for _, li := range lineItems {
		total += li.TotalAmount
		taxTotal += li.TaxAmount
	}
	if req.TaxRate > 0 {
		taxTotal = total * req.TaxRate
	}
	if _, err := uc.invoiceRepo.Update(ctx, domain.UpdateInvoiceParams{
		ID:          inv.ID,
		IssueDate:   inv.IssueDate,
		DueDate:     inv.DueDate,
		TotalAmount: total + taxTotal,
		TaxAmount:   taxTotal,
		Notes:       inv.Notes,
		UpdatedBy:   callerID,
	}); err != nil {
		return nil, err
	}

	// Re-fetch to get final state
	finalInv, _ := uc.invoiceRepo.FindByID(ctx, inv.ID)
	if finalInv == nil {
		finalInv = inv
		finalInv.TotalAmount = total + taxTotal
		finalInv.TaxAmount = taxTotal
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "billing", Resource: "invoices",
		ResourceID: &inv.ID, Action: "GENERATE", IPAddress: ip,
		NewValue: map[string]any{"engagement_id": req.EngagementID, "line_items": len(lineItems)},
	})

	return &GenerateFromEngagementResponse{
		Invoice:   toInvoiceResponse(finalInv),
		LineItems: lineItems,
	}, nil
}

func (uc *GenerateUseCase) generateTMLineItems(ctx context.Context, invoiceID uuid.UUID, req GenerateFromEngagementRequest, rateByStaff map[uuid.UUID]float64) ([]LineItemResponse, error) {
	entries, err := uc.tsSource.ListLockedByEngagement(ctx, req.EngagementID, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, err
	}

	// Group hours by staff member
	hoursByStaff := map[uuid.UUID]float64{}
	for _, e := range entries {
		hoursByStaff[e.CreatedBy] += e.HoursWorked
	}

	var lineItems []LineItemResponse
	for staffID, totalHours := range hoursByStaff {
		rate := rateByStaff[staffID]
		if rate == 0 {
			rate = 0 // no rate → line item amount = 0 (partner can edit)
		}
		lineTotal := totalHours * rate
		snap, _ := json.Marshal(map[string]any{
			"staff_id": staffID.String(), "hours": totalHours, "rate": rate,
			"period": fmt.Sprintf("%s – %s", req.PeriodStart.Format("2006-01-02"), req.PeriodEnd.Format("2006-01-02")),
		})
		item, err := uc.lineRepo.Add(ctx, domain.AddLineItemParams{
			InvoiceID:    invoiceID,
			Description:  fmt.Sprintf("Professional services — %s (%.2f hrs @ %.0f/hr)", staffID, totalHours, rate),
			Quantity:     totalHours,
			UnitPrice:    rate,
			TaxAmount:    0,
			TotalAmount:  lineTotal,
			SourceType:   domain.SourceTimesheetHours,
			SnapshotData: snap,
		})
		if err != nil {
			return nil, err
		}
		lineItems = append(lineItems, toLineItemResponse(item))
	}
	return lineItems, nil
}

func (uc *GenerateUseCase) addFixedFeeLineItem(ctx context.Context, invoiceID uuid.UUID, eng *engdomain.Engagement) (LineItemResponse, error) {
	snap, _ := json.Marshal(map[string]any{
		"engagement_id": eng.ID.String(),
		"fee_type":      string(eng.FeeType),
		"fee_amount":    eng.FeeAmount,
	})
	item, err := uc.lineRepo.Add(ctx, domain.AddLineItemParams{
		InvoiceID:    invoiceID,
		Description:  fmt.Sprintf("Professional services — %s engagement fee", eng.ServiceType),
		Quantity:     1,
		UnitPrice:    eng.FeeAmount,
		TaxAmount:    0,
		TotalAmount:  eng.FeeAmount,
		SourceType:   domain.SourceEngagementFee,
		SnapshotData: snap,
	})
	if err != nil {
		return LineItemResponse{}, err
	}
	return toLineItemResponse(item), nil
}

func resolveInvoiceType(feeType engdomain.FeeType) domain.InvoiceType {
	switch feeType {
	case engdomain.FeeTimeAndMaterial:
		return domain.InvoiceTypeTimeAndMaterial
	case engdomain.FeeRetainer:
		return domain.InvoiceTypeRetainer
	default:
		return domain.InvoiceTypeFixedFee
	}
}

// ─── Adapter: implements EngagementSource using engagement repositories ────────

// EngagementAdapter wraps engagement repositories into a EngagementSource.
type EngagementAdapter struct {
	engRepo    engdomain.EngagementRepository
	memberRepo engdomain.MemberRepository
	costRepo   engdomain.CostRepository
}

// NewEngagementAdapter creates an EngagementAdapter.
func NewEngagementAdapter(
	engRepo engdomain.EngagementRepository,
	memberRepo engdomain.MemberRepository,
	costRepo engdomain.CostRepository,
) *EngagementAdapter {
	return &EngagementAdapter{engRepo: engRepo, memberRepo: memberRepo, costRepo: costRepo}
}

func (a *EngagementAdapter) FindByID(ctx context.Context, id uuid.UUID) (*engdomain.Engagement, error) {
	return a.engRepo.FindByID(ctx, id)
}
func (a *EngagementAdapter) ListByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*engdomain.EngagementMember, error) {
	return a.memberRepo.ListByEngagement(ctx, engagementID)
}
func (a *EngagementAdapter) ListCostsByEngagement(ctx context.Context, engagementID uuid.UUID) ([]*engdomain.DirectCost, error) {
	return a.costRepo.ListByEngagement(ctx, engagementID)
}

// ─── Adapter: implements TimesheetEntrySource using timesheet EntryRepository ──

// TimesheetEntryAdapter wraps the timesheet entry repository.
type TimesheetEntryAdapter struct {
	repo tsdomain.EntryRepository
}

// NewTimesheetEntryAdapter creates a TimesheetEntryAdapter.
func NewTimesheetEntryAdapter(repo tsdomain.EntryRepository) *TimesheetEntryAdapter {
	return &TimesheetEntryAdapter{repo: repo}
}

func (a *TimesheetEntryAdapter) ListLockedByEngagement(ctx context.Context, engagementID uuid.UUID, start, end time.Time) ([]*tsdomain.TimesheetEntry, error) {
	return a.repo.ListLockedByEngagement(ctx, engagementID, start, end)
}

