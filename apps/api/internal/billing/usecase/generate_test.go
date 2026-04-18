package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	engdomain "github.com/mdh/erp-audit/api/internal/engagement/domain"
	tsdomain "github.com/mdh/erp-audit/api/internal/timesheet/domain"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/internal/billing/usecase"
)

// ── mock engagement source ────────────────────────────────────────────────────

type mockEngSource struct {
	eng     *engdomain.Engagement
	engErr  error
	members []*engdomain.EngagementMember
	costs   []*engdomain.DirectCost
}

func (m *mockEngSource) FindByID(_ context.Context, _ uuid.UUID) (*engdomain.Engagement, error) {
	return m.eng, m.engErr
}
func (m *mockEngSource) ListByEngagement(_ context.Context, _ uuid.UUID) ([]*engdomain.EngagementMember, error) {
	return m.members, nil
}
func (m *mockEngSource) ListCostsByEngagement(_ context.Context, _ uuid.UUID) ([]*engdomain.DirectCost, error) {
	return m.costs, nil
}

// ── mock timesheet entry source ───────────────────────────────────────────────

type mockTSSource struct {
	entries []*tsdomain.TimesheetEntry
	err     error
}

func (m *mockTSSource) ListLockedByEngagement(_ context.Context, _ uuid.UUID, _, _ time.Time) ([]*tsdomain.TimesheetEntry, error) {
	return m.entries, m.err
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newEngagement(feeType engdomain.FeeType, feeAmount float64) *engdomain.Engagement {
	clientID := uuid.New()
	return &engdomain.Engagement{
		ID:          uuid.New(),
		ClientID:    clientID,
		ServiceType: engdomain.ServiceAudit,
		FeeType:     feeType,
		FeeAmount:   feeAmount,
		Status:      engdomain.StatusActive,
		CreatedBy:   uuid.New(),
		UpdatedBy:   uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestGenerateUseCase_FixedFee(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	eng := newEngagement(engdomain.FeeFixed, 50_000_000)

	lineItem := &domain.InvoiceLineItem{
		ID:          uuid.New(),
		Description: "Professional services",
		Quantity:    1, UnitPrice: 50_000_000, TotalAmount: 50_000_000,
		SourceType: domain.SourceEngagementFee,
		CreatedAt:  time.Now(),
	}
	inv := newInvoice(domain.InvoiceStatusDraft)
	inv.EngagementID = &eng.ID
	inv.ClientID = eng.ClientID
	inv.InvoiceType = domain.InvoiceTypeFixedFee

	invRepo := &mockInvoiceRepo{created: inv, updated: inv, found: inv}
	lineRepo := &mockLineItemRepo{added: lineItem}
	engSource := &mockEngSource{eng: eng}
	tsSource := &mockTSSource{}

	uc := usecase.NewGenerateUseCase(invRepo, lineRepo, engSource, tsSource, nil)
	resp, err := uc.GenerateFromEngagement(context.Background(), usecase.GenerateFromEngagementRequest{
		EngagementID:  eng.ID,
		InvoiceNumber: "INV-2026-001",
		PeriodStart:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	}, callerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.LineItems) != 1 {
		t.Errorf("want 1 line item (fixed fee), got %d", len(resp.LineItems))
	}
	if resp.LineItems[0].SourceType != domain.SourceEngagementFee {
		t.Errorf("want source ENGAGEMENT_FEE, got %s", resp.LineItems[0].SourceType)
	}
}

func TestGenerateUseCase_TimeAndMaterial_GroupsByStaff(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	eng := newEngagement(engdomain.FeeTimeAndMaterial, 0)

	staffA := uuid.New()
	staffB := uuid.New()
	rate := 500_000.0

	eng2 := *eng
	members := []*engdomain.EngagementMember{
		{StaffID: staffA, HourlyRate: &rate},
		{StaffID: staffB, HourlyRate: &rate},
	}

	// 3 entries: staffA = 8h, staffB = 4h (using CreatedBy as staff proxy in test)
	entries := []*tsdomain.TimesheetEntry{
		{ID: uuid.New(), CreatedBy: staffA, HoursWorked: 8, EngagementID: eng2.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), CreatedBy: staffB, HoursWorked: 4, EngagementID: eng2.ID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	lineItem := &domain.InvoiceLineItem{
		ID:        uuid.New(),
		Quantity:  8, UnitPrice: rate, TotalAmount: 8 * rate,
		SourceType: domain.SourceTimesheetHours, CreatedAt: time.Now(),
	}
	inv := newInvoice(domain.InvoiceStatusDraft)
	inv.EngagementID = &eng.ID
	inv.ClientID = eng.ClientID
	inv.InvoiceType = domain.InvoiceTypeTimeAndMaterial

	invRepo := &mockInvoiceRepo{created: inv, updated: inv, found: inv}
	lineRepo := &mockLineItemRepo{added: lineItem}
	engSource := &mockEngSource{eng: &eng2, members: members, costs: nil}
	tsSource := &mockTSSource{entries: entries}

	uc := usecase.NewGenerateUseCase(invRepo, lineRepo, engSource, tsSource, nil)
	resp, err := uc.GenerateFromEngagement(context.Background(), usecase.GenerateFromEngagementRequest{
		EngagementID:  eng.ID,
		InvoiceNumber: "INV-2026-002",
		PeriodStart:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	}, callerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2 staff members → 2 line items
	if len(resp.LineItems) != 2 {
		t.Errorf("want 2 line items (one per staff), got %d", len(resp.LineItems))
	}
}

func TestGenerateUseCase_IncludesApprovedDirectCosts(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	eng := newEngagement(engdomain.FeeFixed, 10_000_000)

	approvedCost := &engdomain.DirectCost{
		ID:           uuid.New(),
		EngagementID: eng.ID,
		CostType:     engdomain.CostTravel,
		Description:  "Flight to client site",
		Amount:       2_000_000,
		Status:       engdomain.CostApproved,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		CreatedBy:    callerID,
		UpdatedBy:    callerID,
	}
	draftCost := &engdomain.DirectCost{
		ID:     uuid.New(), Status: engdomain.CostDraft, Amount: 500_000,
		Description: "Meals", CostType: engdomain.CostMeals,
		CreatedAt: time.Now(), UpdatedAt: time.Now(), CreatedBy: callerID, UpdatedBy: callerID,
	}

	inv := newInvoice(domain.InvoiceStatusDraft)
	inv.EngagementID = &eng.ID
	inv.ClientID = eng.ClientID

	callCount := 0
	lineItems := []*domain.InvoiceLineItem{
		{ID: uuid.New(), SourceType: domain.SourceEngagementFee, TotalAmount: 10_000_000, CreatedAt: time.Now()},
		{ID: uuid.New(), SourceType: domain.SourceDirectCost, TotalAmount: 2_000_000, CreatedAt: time.Now()},
	}
	lineRepo := &mockLineItemRepoMulti{items: lineItems, callCount: &callCount}

	invRepo := &mockInvoiceRepo{created: inv, updated: inv, found: inv}
	engSource := &mockEngSource{eng: eng, costs: []*engdomain.DirectCost{approvedCost, draftCost}}
	tsSource := &mockTSSource{}

	uc := usecase.NewGenerateUseCase(invRepo, lineRepo, engSource, tsSource, nil)
	resp, err := uc.GenerateFromEngagement(context.Background(), usecase.GenerateFromEngagementRequest{
		EngagementID:  eng.ID,
		InvoiceNumber: "INV-2026-003",
		PeriodStart:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
	}, callerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 fixed fee + 1 approved direct cost (draft cost excluded)
	if len(resp.LineItems) != 2 {
		t.Errorf("want 2 line items (fixed fee + approved cost), got %d", len(resp.LineItems))
	}
}

func TestGenerateUseCase_EngagementNotFound(t *testing.T) {
	t.Parallel()
	engSource := &mockEngSource{engErr: engdomain.ErrEngagementNotFound}
	tsSource := &mockTSSource{}
	invRepo := &mockInvoiceRepo{}
	lineRepo := &mockLineItemRepo{}

	uc := usecase.NewGenerateUseCase(invRepo, lineRepo, engSource, tsSource, nil)
	_, err := uc.GenerateFromEngagement(context.Background(), usecase.GenerateFromEngagementRequest{
		EngagementID:  uuid.New(),
		InvoiceNumber: "INV-ERR-001",
		PeriodStart:   time.Now(),
		PeriodEnd:     time.Now(),
	}, uuid.New(), "")
	if err == nil {
		t.Fatal("expected error for engagement not found, got nil")
	}
}

// mockLineItemRepoMulti returns different items on successive calls.
type mockLineItemRepoMulti struct {
	items     []*domain.InvoiceLineItem
	callCount *int
}

func (m *mockLineItemRepoMulti) Add(_ context.Context, _ domain.AddLineItemParams) (*domain.InvoiceLineItem, error) {
	idx := *m.callCount
	*m.callCount++
	if idx < len(m.items) {
		return m.items[idx], nil
	}
	return m.items[len(m.items)-1], nil
}
func (m *mockLineItemRepoMulti) FindByID(_ context.Context, _ uuid.UUID) (*domain.InvoiceLineItem, error) {
	return nil, nil
}
func (m *mockLineItemRepoMulti) ListByInvoice(_ context.Context, _ uuid.UUID) ([]*domain.InvoiceLineItem, error) {
	return m.items, nil
}
func (m *mockLineItemRepoMulti) Delete(_ context.Context, _ uuid.UUID) error { return nil }
