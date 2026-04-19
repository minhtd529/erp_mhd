package usecase_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/commission/domain"
	"github.com/mdh/erp-audit/api/internal/commission/usecase"
)

// ── mock plan repo ────────────────────────────────────────────────────────────

type mockPlanRepo struct {
	created  *domain.CommissionPlan
	found    *domain.CommissionPlan
	updated  *domain.CommissionPlan
	deact    *domain.CommissionPlan
	list     []*domain.CommissionPlan
	total    int64
	err      error
}

func (m *mockPlanRepo) Create(_ context.Context, _ domain.CreatePlanParams) (*domain.CommissionPlan, error) {
	return m.created, m.err
}
func (m *mockPlanRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.CommissionPlan, error) {
	return m.found, m.err
}
func (m *mockPlanRepo) FindByCode(_ context.Context, _ string) (*domain.CommissionPlan, error) {
	return m.found, m.err
}
func (m *mockPlanRepo) Update(_ context.Context, _ domain.UpdatePlanParams) (*domain.CommissionPlan, error) {
	return m.updated, m.err
}
func (m *mockPlanRepo) Deactivate(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*domain.CommissionPlan, error) {
	return m.deact, m.err
}
func (m *mockPlanRepo) List(_ context.Context, _ domain.ListPlansFilter, _, _ int) ([]*domain.CommissionPlan, int64, error) {
	return m.list, m.total, m.err
}

// ── mock eng commission repo ──────────────────────────────────────────────────

type mockECRepo struct {
	created  *domain.EngagementCommission
	found    *domain.EngagementCommission
	list     []*domain.EngagementCommission
	total    int64
	sumRate  float64
	err      error
}

func (m *mockECRepo) Create(_ context.Context, _ domain.CreateEngCommissionParams) (*domain.EngagementCommission, error) {
	return m.created, m.err
}
func (m *mockECRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.EngagementCommission, error) {
	return m.found, m.err
}
func (m *mockECRepo) List(_ context.Context, _ domain.ListEngCommissionsFilter, _, _ int) ([]*domain.EngagementCommission, int64, error) {
	return m.list, m.total, m.err
}
func (m *mockECRepo) SumRateByEngagement(_ context.Context, _ uuid.UUID) (float64, error) {
	return m.sumRate, m.err
}
func (m *mockECRepo) Cancel(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*domain.EngagementCommission, error) {
	return m.found, m.err
}
func (m *mockECRepo) Approve(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*domain.EngagementCommission, error) {
	return m.found, m.err
}
func (m *mockECRepo) ListActiveByTrigger(_ context.Context, _ uuid.UUID, _ domain.CommissionTrigger) ([]*domain.EngagementCommission, error) {
	return m.list, m.err
}
func (m *mockECRepo) SumHoldbackByEngagement(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, m.err
}

// ── Plan tests ────────────────────────────────────────────────────────────────

func TestPlanUseCase_Create_Success(t *testing.T) {
	t.Parallel()
	planID := uuid.New()
	callerID := uuid.New()
	repo := &mockPlanRepo{created: &domain.CommissionPlan{
		ID:          planID,
		Code:        "audit_std_5pct",
		Name:        "Standard 5%",
		Type:        domain.CommissionTypeFlat,
		DefaultRate: 0.05,
		Tiers:       []domain.CommissionTier{},
		ServiceTypes: []string{},
		IsActive:    true,
		CreatedBy:   callerID,
	}}
	uc := usecase.NewPlanUseCase(repo, nil)

	resp, err := uc.Create(context.Background(), usecase.PlanCreateRequest{
		Code:      "audit_std_5pct",
		Name:      "Standard 5%",
		Type:      domain.CommissionTypeFlat,
		ApplyBase: domain.CommBaseFeePaid,
		TriggerOn: domain.CommTriggerPaymentReceived,
	}, callerID, net.ParseIP("127.0.0.1"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != planID {
		t.Errorf("want %v, got %v", planID, resp.ID)
	}
	if resp.Code != "audit_std_5pct" {
		t.Errorf("wrong code: %s", resp.Code)
	}
}

func TestPlanUseCase_Create_CodeConflict(t *testing.T) {
	t.Parallel()
	repo := &mockPlanRepo{err: domain.ErrPlanCodeConflict}
	uc := usecase.NewPlanUseCase(repo, nil)

	_, err := uc.Create(context.Background(), usecase.PlanCreateRequest{
		Code:      "dup",
		Name:      "Dup",
		Type:      domain.CommissionTypeFlat,
		ApplyBase: domain.CommBaseFeePaid,
		TriggerOn: domain.CommTriggerPaymentReceived,
	}, uuid.New(), net.ParseIP("127.0.0.1"))

	if !errors.Is(err, domain.ErrPlanCodeConflict) {
		t.Errorf("want ErrPlanCodeConflict, got %v", err)
	}
}

func TestPlanUseCase_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := &mockPlanRepo{err: domain.ErrPlanNotFound}
	uc := usecase.NewPlanUseCase(repo, nil)

	_, err := uc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrPlanNotFound) {
		t.Errorf("want ErrPlanNotFound, got %v", err)
	}
}

func TestPlanUseCase_List(t *testing.T) {
	t.Parallel()
	repo := &mockPlanRepo{
		list: []*domain.CommissionPlan{
			{ID: uuid.New(), Code: "p1", Tiers: []domain.CommissionTier{}, ServiceTypes: []string{}},
			{ID: uuid.New(), Code: "p2", Tiers: []domain.CommissionTier{}, ServiceTypes: []string{}},
		},
		total: 2,
	}
	uc := usecase.NewPlanUseCase(repo, nil)

	result, err := uc.List(context.Background(), domain.ListPlansFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("want total 2, got %d", result.Total)
	}
	if len(result.Data) != 2 {
		t.Errorf("want 2 items, got %d", len(result.Data))
	}
}

func TestPlanUseCase_Deactivate_Success(t *testing.T) {
	t.Parallel()
	planID := uuid.New()
	repo := &mockPlanRepo{deact: &domain.CommissionPlan{
		ID:           planID,
		Code:         "old",
		IsActive:     false,
		Tiers:        []domain.CommissionTier{},
		ServiceTypes: []string{},
	}}
	uc := usecase.NewPlanUseCase(repo, nil)

	resp, err := uc.Deactivate(context.Background(), planID, uuid.New(), net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.IsActive {
		t.Error("want is_active=false")
	}
}

// ── EngCommission tests ───────────────────────────────────────────────────────

func TestEngCommissionUseCase_Create_Success(t *testing.T) {
	t.Parallel()
	ecID := uuid.New()
	callerID := uuid.New()
	repo := &mockECRepo{
		sumRate: 0.0, // no existing commissions
		created: &domain.EngagementCommission{
			ID:           ecID,
			EngagementID: uuid.New(),
			Role:         domain.SalesRolePrimary,
			Rate:         0.05,
			Status:       "active",
			Tiers:        []domain.CommissionTier{},
		},
	}
	uc := usecase.NewEngCommissionUseCase(repo, nil)

	resp, err := uc.Create(context.Background(), usecase.EngCommissionCreateRequest{
		EngagementID:  uuid.New(),
		SalespersonID: uuid.New(),
		Role:          domain.SalesRolePrimary,
		RateType:      domain.CommissionTypeFlat,
		Rate:          0.05,
		ApplyBase:     domain.CommBaseFeePaid,
		TriggerOn:     domain.CommTriggerPaymentReceived,
	}, callerID, net.ParseIP("127.0.0.1"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != ecID {
		t.Errorf("wrong ID")
	}
}

func TestEngCommissionUseCase_Create_RateExceeds(t *testing.T) {
	t.Parallel()
	repo := &mockECRepo{sumRate: 0.96} // existing 96%, adding 5% would exceed 100%
	uc := usecase.NewEngCommissionUseCase(repo, nil)

	_, err := uc.Create(context.Background(), usecase.EngCommissionCreateRequest{
		EngagementID:  uuid.New(),
		SalespersonID: uuid.New(),
		Role:          domain.SalesRolePrimary,
		RateType:      domain.CommissionTypeFlat,
		Rate:          0.05,
		ApplyBase:     domain.CommBaseFeePaid,
		TriggerOn:     domain.CommTriggerPaymentReceived,
	}, uuid.New(), net.ParseIP("127.0.0.1"))

	if !errors.Is(err, domain.ErrEngCommissionRateExceeds) {
		t.Errorf("want ErrEngCommissionRateExceeds, got %v", err)
	}
}

func TestEngCommissionUseCase_Cancel_NotFound(t *testing.T) {
	t.Parallel()
	repo := &mockECRepo{err: domain.ErrEngCommissionNotFound}
	uc := usecase.NewEngCommissionUseCase(repo, nil)

	_, err := uc.Cancel(context.Background(), uuid.New(), uuid.New(), net.ParseIP("127.0.0.1"))
	if !errors.Is(err, domain.ErrEngCommissionNotFound) {
		t.Errorf("want ErrEngCommissionNotFound, got %v", err)
	}
}

func TestEngCommissionUseCase_Approve_Success(t *testing.T) {
	t.Parallel()
	ecID := uuid.New()
	approverID := uuid.New()
	repo := &mockECRepo{found: &domain.EngagementCommission{
		ID:         ecID,
		Status:     "active",
		ApprovedBy: &approverID,
		Tiers:      []domain.CommissionTier{},
	}}
	uc := usecase.NewEngCommissionUseCase(repo, nil)

	resp, err := uc.Approve(context.Background(), ecID, approverID, net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ApprovedBy == nil || *resp.ApprovedBy != approverID {
		t.Errorf("wrong approved_by")
	}
}

// ── Accrual tests ─────────────────────────────────────────────────────────────

type mockRecordRepo struct {
	created   *domain.CommissionRecord
	found     *domain.CommissionRecord
	list      []*domain.CommissionRecord
	total     int64
	summary   *domain.SalespersonSummary
	err       error
	createErr error // per-Create error (overrides err for Create only)
}

func (m *mockRecordRepo) Create(_ context.Context, r *domain.CommissionRecord) (*domain.CommissionRecord, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.created != nil {
		return m.created, nil
	}
	return r, nil
}
func (m *mockRecordRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.CommissionRecord, error) {
	return m.found, m.err
}
func (m *mockRecordRepo) List(_ context.Context, _ domain.ListRecordsFilter, _, _ int) ([]*domain.CommissionRecord, int64, error) {
	return m.list, m.total, m.err
}
func (m *mockRecordRepo) ListByInvoice(_ context.Context, _ uuid.UUID) ([]*domain.CommissionRecord, error) {
	return m.list, m.err
}
func (m *mockRecordRepo) ListForStatement(_ context.Context, _ domain.StatementFilter) ([]*domain.CommissionRecord, error) {
	return m.list, m.err
}
func (m *mockRecordRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.CommissionStatus, _ *uuid.UUID, updatedBy uuid.UUID, _ string) (*domain.CommissionRecord, error) {
	if m.found != nil {
		cp := *m.found
		cp.UpdatedBy = &updatedBy
		return &cp, m.err
	}
	return m.found, m.err
}
func (m *mockRecordRepo) BulkUpdateStatus(_ context.Context, _ []uuid.UUID, _ domain.CommissionStatus, _ *uuid.UUID, _ uuid.UUID, _ string) (int64, error) {
	return m.total, m.err
}
func (m *mockRecordRepo) SummarySalesperson(_ context.Context, _ uuid.UUID) (*domain.SalespersonSummary, error) {
	return m.summary, m.err
}
func (m *mockRecordRepo) ListByTeam(_ context.Context, _ uuid.UUID, _, _ int) ([]*domain.CommissionRecord, int64, error) {
	return m.list, m.total, m.err
}

type mockBillingReader struct {
	invoice *domain.InvoiceAccrualData
	payment *domain.PaymentAccrualData
	err     error
}

func (m *mockBillingReader) GetInvoiceForAccrual(_ context.Context, _ uuid.UUID) (*domain.InvoiceAccrualData, error) {
	return m.invoice, m.err
}
func (m *mockBillingReader) GetPaymentForAccrual(_ context.Context, _ uuid.UUID) (*domain.PaymentAccrualData, error) {
	return m.payment, m.err
}

func TestAccrualUseCase_AccrueOnInvoiceIssued_FlatRate(t *testing.T) {
	t.Parallel()
	engID := uuid.New()
	invoiceID := uuid.New()
	spID := uuid.New()
	ecID := uuid.New()

	billingReader := &mockBillingReader{
		invoice: &domain.InvoiceAccrualData{
			InvoiceID:    invoiceID,
			EngagementID: engID,
			TotalAmount:  200_000_000, // 200M VND
		},
	}
	ecRepo := &mockECRepo{
		list: []*domain.EngagementCommission{
			{
				ID:            ecID,
				EngagementID:  engID,
				SalespersonID: spID,
				RateType:      domain.CommissionTypeFlat,
				Rate:          0.05,
				HoldbackPct:   0.20,
				Tiers:         []domain.CommissionTier{},
			},
		},
	}
	recordRepo := &mockRecordRepo{}
	uc := usecase.NewAccrualUseCase(ecRepo, recordRepo, billingReader, nil)

	if err := uc.AccrueOnInvoiceIssued(context.Background(), invoiceID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAccrualUseCase_AccrueOnInvoiceIssued_Idempotent(t *testing.T) {
	t.Parallel()
	invoiceID := uuid.New()
	billingReader := &mockBillingReader{
		invoice: &domain.InvoiceAccrualData{
			InvoiceID:    invoiceID,
			EngagementID: uuid.New(),
			TotalAmount:  100_000_000,
		},
	}
	ecRepo := &mockECRepo{
		list: []*domain.EngagementCommission{
			{ID: uuid.New(), RateType: domain.CommissionTypeFlat, Rate: 0.05, Tiers: []domain.CommissionTier{}},
		},
	}
	// Simulate duplicate accrual error — should be silently skipped.
	recordRepo := &mockRecordRepo{err: domain.ErrDuplicateAccrual}
	uc := usecase.NewAccrualUseCase(ecRepo, recordRepo, billingReader, nil)

	if err := uc.AccrueOnInvoiceIssued(context.Background(), invoiceID); err != nil {
		t.Fatalf("duplicate accrual should be silently skipped, got: %v", err)
	}
}

func TestAccrualUseCase_AccrueOnPaymentReceived_FlatRate(t *testing.T) {
	t.Parallel()
	paymentID := uuid.New()
	engID := uuid.New()
	spID := uuid.New()

	billingReader := &mockBillingReader{
		payment: &domain.PaymentAccrualData{
			PaymentID:    paymentID,
			InvoiceID:    uuid.New(),
			EngagementID: engID,
			Amount:       100_000_000,
		},
	}
	ecRepo := &mockECRepo{
		list: []*domain.EngagementCommission{
			{
				ID:            uuid.New(),
				EngagementID:  engID,
				SalespersonID: spID,
				RateType:      domain.CommissionTypeFlat,
				Rate:          0.05,
				HoldbackPct:   0.0,
				Tiers:         []domain.CommissionTier{},
			},
		},
	}
	recordRepo := &mockRecordRepo{}
	uc := usecase.NewAccrualUseCase(ecRepo, recordRepo, billingReader, nil)

	if err := uc.AccrueOnPaymentReceived(context.Background(), paymentID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAccrualUseCase_AccrueOnPaymentReceived_BillingError(t *testing.T) {
	t.Parallel()
	billingReader := &mockBillingReader{err: errors.New("payment not found")}
	ecRepo := &mockECRepo{}
	recordRepo := &mockRecordRepo{}
	uc := usecase.NewAccrualUseCase(ecRepo, recordRepo, billingReader, nil)

	err := uc.AccrueOnPaymentReceived(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error when billing reader fails")
	}
}

func TestAccrualUseCase_CalculateFixed(t *testing.T) {
	t.Parallel()
	paymentID := uuid.New()
	fixed := int64(5_000_000)
	billingReader := &mockBillingReader{
		payment: &domain.PaymentAccrualData{
			PaymentID:    paymentID,
			EngagementID: uuid.New(),
			Amount:       200_000_000,
		},
	}
	ecRepo := &mockECRepo{
		list: []*domain.EngagementCommission{
			{
				ID:          uuid.New(),
				RateType:    domain.CommissionTypeFixed,
				FixedAmount: &fixed,
				Tiers:       []domain.CommissionTier{},
			},
		},
	}
	var capturedRecord *domain.CommissionRecord
	recordRepo := &mockRecordRepo{}
	_ = capturedRecord

	uc := usecase.NewAccrualUseCase(ecRepo, recordRepo, billingReader, nil)
	if err := uc.AccrueOnPaymentReceived(context.Background(), paymentID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAccrualUseCase_ClawbackOnInvoiceCancelled_CreatesNegativeRecord(t *testing.T) {
	t.Parallel()
	invoiceID := uuid.New()
	origID := uuid.New()
	spID := uuid.New()
	ecID := uuid.New()
	engID := uuid.New()

	// Existing accrued record for the cancelled invoice
	original := &domain.CommissionRecord{
		ID:                     origID,
		EngagementCommissionID: ecID,
		EngagementID:           engID,
		SalespersonID:          spID,
		InvoiceID:              &invoiceID,
		BaseAmount:             10_000_000,
		Rate:                   0.05,
		CalculatedAmount:       500_000,
		HoldbackAmount:         100_000,
		PayableAmount:          400_000,
		Status:                 domain.CommStatusAccrued,
		IsClawback:             false,
	}

	recordRepo := &mockRecordRepo{list: []*domain.CommissionRecord{original}}
	ecRepo := &mockECRepo{}
	billingReader := &mockBillingReader{}
	uc := usecase.NewAccrualUseCase(ecRepo, recordRepo, billingReader, nil)

	if err := uc.ClawbackOnInvoiceCancelled(context.Background(), invoiceID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAccrualUseCase_ClawbackOnInvoiceCancelled_SkipsExistingClawbacks(t *testing.T) {
	t.Parallel()
	invoiceID := uuid.New()

	// Record is already a clawback — should not create another
	existing := &domain.CommissionRecord{
		ID:         uuid.New(),
		InvoiceID:  &invoiceID,
		IsClawback: true,
		Status:     domain.CommStatusClawback,
	}

	recordRepo := &mockRecordRepo{list: []*domain.CommissionRecord{existing}}
	uc := usecase.NewAccrualUseCase(&mockECRepo{}, recordRepo, &mockBillingReader{}, nil)

	if err := uc.ClawbackOnInvoiceCancelled(context.Background(), invoiceID); err != nil {
		t.Fatalf("expected no error when all records are already clawbacks: %v", err)
	}
}

func TestAccrualUseCase_ClawbackOnInvoiceCancelled_SkipsCancelledRecords(t *testing.T) {
	t.Parallel()
	invoiceID := uuid.New()

	already := &domain.CommissionRecord{
		ID:         uuid.New(),
		InvoiceID:  &invoiceID,
		IsClawback: false,
		Status:     domain.CommStatusCancelled,
	}

	recordRepo := &mockRecordRepo{list: []*domain.CommissionRecord{already}}
	uc := usecase.NewAccrualUseCase(&mockECRepo{}, recordRepo, &mockBillingReader{}, nil)

	if err := uc.ClawbackOnInvoiceCancelled(context.Background(), invoiceID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAccrualUseCase_ClawbackOnInvoiceCancelled_IdempotentDuplicate(t *testing.T) {
	t.Parallel()
	invoiceID := uuid.New()

	orig := &domain.CommissionRecord{
		ID:               uuid.New(),
		InvoiceID:        &invoiceID,
		BaseAmount:       5_000_000,
		CalculatedAmount: 250_000,
		HoldbackAmount:   50_000,
		PayableAmount:    200_000,
		Status:           domain.CommStatusAccrued,
		IsClawback:       false,
	}

	// Create returns ErrDuplicateAccrual — idempotent clawback should not error
	recordRepo := &mockRecordRepo{
		list:      []*domain.CommissionRecord{orig},
		createErr: domain.ErrDuplicateAccrual,
	}
	uc := usecase.NewAccrualUseCase(&mockECRepo{}, recordRepo, &mockBillingReader{}, nil)

	if err := uc.ClawbackOnInvoiceCancelled(context.Background(), invoiceID); err != nil {
		t.Fatalf("ErrDuplicateAccrual should be silently ignored; got: %v", err)
	}
}

// ── RecordUseCase tests ───────────────────────────────────────────────────────

func TestRecordUseCase_Approve_Success(t *testing.T) {
	t.Parallel()
	recID := uuid.New()
	callerID := uuid.New()
	repo := &mockRecordRepo{
		found: &domain.CommissionRecord{ID: recID, Status: domain.CommStatusAccrued},
	}
	uc := usecase.NewRecordUseCase(repo, nil)

	resp, err := uc.Approve(context.Background(), recID, callerID, net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != recID {
		t.Errorf("wrong ID returned")
	}
}

func TestRecordUseCase_Approve_NotApprovable(t *testing.T) {
	t.Parallel()
	recID := uuid.New()
	repo := &mockRecordRepo{
		found: &domain.CommissionRecord{ID: recID, Status: domain.CommStatusApproved},
	}
	uc := usecase.NewRecordUseCase(repo, nil)

	_, err := uc.Approve(context.Background(), recID, uuid.New(), net.ParseIP("127.0.0.1"))
	if !errors.Is(err, domain.ErrRecordNotApprovable) {
		t.Errorf("want ErrRecordNotApprovable, got %v", err)
	}
}

func TestRecordUseCase_Approve_NotFound(t *testing.T) {
	t.Parallel()
	repo := &mockRecordRepo{err: domain.ErrCommissionRecordNotFound}
	uc := usecase.NewRecordUseCase(repo, nil)

	_, err := uc.Approve(context.Background(), uuid.New(), uuid.New(), net.ParseIP("127.0.0.1"))
	if !errors.Is(err, domain.ErrCommissionRecordNotFound) {
		t.Errorf("want ErrCommissionRecordNotFound, got %v", err)
	}
}

func TestRecordUseCase_MarkPaid_Success(t *testing.T) {
	t.Parallel()
	recID := uuid.New()
	callerID := uuid.New()
	repo := &mockRecordRepo{
		found: &domain.CommissionRecord{ID: recID, Status: domain.CommStatusApproved},
	}
	uc := usecase.NewRecordUseCase(repo, nil)

	resp, err := uc.MarkPaid(context.Background(), recID, "PAY-001", callerID, net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != recID {
		t.Errorf("wrong ID returned")
	}
}

func TestRecordUseCase_MarkPaid_NotPayable(t *testing.T) {
	t.Parallel()
	recID := uuid.New()
	repo := &mockRecordRepo{
		found: &domain.CommissionRecord{ID: recID, Status: domain.CommStatusAccrued},
	}
	uc := usecase.NewRecordUseCase(repo, nil)

	_, err := uc.MarkPaid(context.Background(), recID, "PAY-001", uuid.New(), net.ParseIP("127.0.0.1"))
	if !errors.Is(err, domain.ErrRecordNotPayable) {
		t.Errorf("want ErrRecordNotPayable, got %v", err)
	}
}

func TestRecordUseCase_BulkApprove(t *testing.T) {
	t.Parallel()
	repo := &mockRecordRepo{total: 3}
	uc := usecase.NewRecordUseCase(repo, nil)

	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	result, err := uc.BulkApprove(context.Background(), usecase.BulkApproveRequest{IDs: ids}, uuid.New(), net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AffectedCount != 3 {
		t.Errorf("want 3, got %d", result.AffectedCount)
	}
}

func TestRecordUseCase_Clawback_Success(t *testing.T) {
	t.Parallel()
	origID := uuid.New()
	repo := &mockRecordRepo{
		found: &domain.CommissionRecord{
			ID:               origID,
			Status:           domain.CommStatusAccrued,
			BaseAmount:       1_000_000,
			Rate:             0.05,
			CalculatedAmount: 50_000,
			HoldbackAmount:   10_000,
			PayableAmount:    40_000,
		},
	}
	uc := usecase.NewRecordUseCase(repo, nil)

	resp, err := uc.Clawback(context.Background(), origID, "duplicate billing", uuid.New(), net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PayableAmount != -40_000 {
		t.Errorf("want -40000, got %d", resp.PayableAmount)
	}
	if !resp.IsClawback {
		t.Error("want is_clawback=true")
	}
}

func TestRecordUseCase_MyCommissionSummary(t *testing.T) {
	t.Parallel()
	spID := uuid.New()
	repo := &mockRecordRepo{
		summary: &domain.SalespersonSummary{
			TotalYTD:      10_000_000,
			TotalMonth:    2_000_000,
			TotalPending:  500_000,
			TotalApproved: 1_500_000,
			TotalPaid:     8_000_000,
		},
	}
	uc := usecase.NewRecordUseCase(repo, nil)

	s, err := uc.MyCommissionSummary(context.Background(), spID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TotalYTD != 10_000_000 {
		t.Errorf("want 10000000, got %d", s.TotalYTD)
	}
}

// ── Statement tests ───────────────────────────────────────────────────────────

func TestRecordUseCase_GetStatement_Quarter(t *testing.T) {
	t.Parallel()
	spID := uuid.New()
	repo := &mockRecordRepo{
		list: []*domain.CommissionRecord{
			{
				ID:               uuid.New(),
				SalespersonID:    spID,
				EngagementID:     uuid.New(),
				Status:           domain.CommStatusApproved,
				CalculatedAmount: 5_000_000,
				PayableAmount:    4_000_000,
			},
			{
				ID:               uuid.New(),
				SalespersonID:    spID,
				EngagementID:     uuid.New(),
				Status:           domain.CommStatusPaid,
				CalculatedAmount: 3_000_000,
				PayableAmount:    2_400_000,
			},
		},
	}
	uc := usecase.NewRecordUseCase(repo, nil)

	stmt, err := uc.GetStatement(context.Background(), spID, "2026-Q1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stmt.Records) != 2 {
		t.Errorf("want 2 records, got %d", len(stmt.Records))
	}
	if stmt.TotalAccrued != 8_000_000 {
		t.Errorf("want total_accrued 8000000, got %d", stmt.TotalAccrued)
	}
	if stmt.TotalPayable != 6_400_000 {
		t.Errorf("want total_payable 6400000, got %d", stmt.TotalPayable)
	}
	if stmt.TotalPaid != 2_400_000 {
		t.Errorf("want total_paid 2400000, got %d", stmt.TotalPaid)
	}
	if stmt.Period != "2026-Q1" {
		t.Errorf("wrong period: %s", stmt.Period)
	}
}

func TestRecordUseCase_GetStatement_Month(t *testing.T) {
	t.Parallel()
	repo := &mockRecordRepo{list: []*domain.CommissionRecord{}}
	uc := usecase.NewRecordUseCase(repo, nil)

	stmt, err := uc.GetStatement(context.Background(), uuid.New(), "2026-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stmt.Records) != 0 {
		t.Errorf("want 0 records, got %d", len(stmt.Records))
	}
}

func TestRecordUseCase_GetStatement_Year(t *testing.T) {
	t.Parallel()
	repo := &mockRecordRepo{list: []*domain.CommissionRecord{}}
	uc := usecase.NewRecordUseCase(repo, nil)

	_, err := uc.GetStatement(context.Background(), uuid.New(), "2026")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordUseCase_GetStatement_InvalidPeriod(t *testing.T) {
	t.Parallel()
	repo := &mockRecordRepo{}
	uc := usecase.NewRecordUseCase(repo, nil)

	_, err := uc.GetStatement(context.Background(), uuid.New(), "invalid")
	if err == nil {
		t.Fatal("expected error for invalid period")
	}
}

func TestPlanUseCase_Update_SetsUpdatedBy(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	planID := uuid.New()
	repo := &mockPlanRepo{updated: &domain.CommissionPlan{
		ID:           planID,
		Code:         "p1",
		Name:         "Updated Plan",
		IsActive:     true,
		Tiers:        []domain.CommissionTier{},
		ServiceTypes: []string{},
		UpdatedBy:    &callerID,
	}}
	uc := usecase.NewPlanUseCase(repo, nil)

	resp, err := uc.Update(context.Background(), planID, usecase.PlanUpdateRequest{
		Name:      "Updated Plan",
		ApplyBase: domain.CommBaseFeePaid,
		TriggerOn: domain.CommTriggerPaymentReceived,
	}, callerID, net.ParseIP("127.0.0.1"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UpdatedBy == nil || *resp.UpdatedBy != callerID {
		t.Errorf("want updated_by=%v, got %v", callerID, resp.UpdatedBy)
	}
}

func TestRecordUseCase_Approve_SetsUpdatedBy(t *testing.T) {
	t.Parallel()
	recID := uuid.New()
	callerID := uuid.New()
	repo := &mockRecordRepo{
		found: &domain.CommissionRecord{ID: recID, Status: domain.CommStatusAccrued},
	}
	uc := usecase.NewRecordUseCase(repo, nil)

	resp, err := uc.Approve(context.Background(), recID, callerID, net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UpdatedBy == nil || *resp.UpdatedBy != callerID {
		t.Errorf("want updated_by=%v, got %v", callerID, resp.UpdatedBy)
	}
}

func TestRecordUseCase_ExportStatementCSV(t *testing.T) {
	t.Parallel()
	spID := uuid.New()
	repo := &mockRecordRepo{
		list: []*domain.CommissionRecord{
			{
				ID:               uuid.New(),
				SalespersonID:    spID,
				EngagementID:     uuid.New(),
				Status:           domain.CommStatusApproved,
				BaseAmount:       10_000_000,
				Rate:             0.05,
				CalculatedAmount: 500_000,
				PayableAmount:    400_000,
			},
		},
	}
	uc := usecase.NewRecordUseCase(repo, nil)

	data, err := uc.ExportStatementCSV(context.Background(), spID, "2026-Q1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty CSV")
	}
	// CSV should have header + 1 data row
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines < 2 {
		t.Errorf("want at least 2 lines (header + data), got %d", lines)
	}
}
