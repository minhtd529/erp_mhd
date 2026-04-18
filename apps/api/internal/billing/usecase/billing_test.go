package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/billing/domain"
	"github.com/mdh/erp-audit/api/internal/billing/usecase"
)

// ── mock invoice repo ──────────────────────────────────────────────────────────

type mockInvoiceRepo struct {
	created    *domain.Invoice
	found      *domain.Invoice
	updated    *domain.Invoice
	statusUpd  *domain.Invoice
	snapUpd    *domain.Invoice
	createErr  error
	findErr    error
	updateErr  error
	statusErr  error
	deleteErr  error
	listItems  []*domain.Invoice
	listTotal  int64
}

func (m *mockInvoiceRepo) Create(_ context.Context, _ domain.CreateInvoiceParams) (*domain.Invoice, error) {
	return m.created, m.createErr
}
func (m *mockInvoiceRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Invoice, error) {
	return m.found, m.findErr
}
func (m *mockInvoiceRepo) Update(_ context.Context, _ domain.UpdateInvoiceParams) (*domain.Invoice, error) {
	return m.updated, m.updateErr
}
func (m *mockInvoiceRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.InvoiceStatus, _ uuid.UUID) (*domain.Invoice, error) {
	return m.statusUpd, m.statusErr
}
func (m *mockInvoiceRepo) UpdateSnapshot(_ context.Context, _ uuid.UUID, _ []byte, _ uuid.UUID) (*domain.Invoice, error) {
	return m.snapUpd, m.statusErr
}
func (m *mockInvoiceRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	return m.deleteErr
}
func (m *mockInvoiceRepo) List(_ context.Context, _ domain.ListInvoicesFilter) ([]*domain.Invoice, int64, error) {
	return m.listItems, m.listTotal, m.findErr
}

// ── mock line item repo ────────────────────────────────────────────────────────

type mockLineItemRepo struct {
	added    *domain.InvoiceLineItem
	addErr   error
	delErr   error
	items    []*domain.InvoiceLineItem
}

func (m *mockLineItemRepo) Add(_ context.Context, _ domain.AddLineItemParams) (*domain.InvoiceLineItem, error) {
	return m.added, m.addErr
}
func (m *mockLineItemRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.InvoiceLineItem, error) {
	return nil, nil
}
func (m *mockLineItemRepo) ListByInvoice(_ context.Context, _ uuid.UUID) ([]*domain.InvoiceLineItem, error) {
	return m.items, nil
}
func (m *mockLineItemRepo) Delete(_ context.Context, _ uuid.UUID) error {
	return m.delErr
}

// ── mock payment repo ──────────────────────────────────────────────────────────

type mockPaymentRepo struct {
	recorded  *domain.Payment
	found     *domain.Payment
	updated   *domain.Payment
	statusUpd *domain.Payment
	recordErr error
	findErr   error
	updateErr error
	statusErr error
	sumPaid   float64
	sumErr    error
	items     []*domain.Payment
}

func (m *mockPaymentRepo) Record(_ context.Context, _ domain.RecordPaymentParams) (*domain.Payment, error) {
	return m.recorded, m.recordErr
}
func (m *mockPaymentRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Payment, error) {
	return m.found, m.findErr
}
func (m *mockPaymentRepo) Update(_ context.Context, _ domain.UpdatePaymentParams) (*domain.Payment, error) {
	return m.updated, m.updateErr
}
func (m *mockPaymentRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.PaymentStatus) (*domain.Payment, error) {
	return m.statusUpd, m.statusErr
}
func (m *mockPaymentRepo) Clear(_ context.Context, _ uuid.UUID) (*domain.Payment, error) {
	if m.found != nil {
		m.found.Status = domain.PaymentCleared
		return m.found, nil
	}
	return m.statusUpd, m.statusErr
}
func (m *mockPaymentRepo) ListByInvoice(_ context.Context, _ uuid.UUID) ([]*domain.Payment, error) {
	return m.items, nil
}
func (m *mockPaymentRepo) SumPaidByInvoice(_ context.Context, _ uuid.UUID) (float64, error) {
	return m.sumPaid, m.sumErr
}

// ── mock memo repo ────────────────────────────────────────────────────────────

type mockMemoRepo struct {
	created   *domain.BillingMemo
	createErr error
	items     []*domain.BillingMemo
	total     int64
}

func (m *mockMemoRepo) Create(_ context.Context, _ domain.CreateMemoParams) (*domain.BillingMemo, error) {
	return m.created, m.createErr
}
func (m *mockMemoRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.BillingMemo, error) {
	return nil, nil
}
func (m *mockMemoRepo) List(_ context.Context, _, _ int) ([]*domain.BillingMemo, int64, error) {
	return m.items, m.total, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newInvoice(status domain.InvoiceStatus) *domain.Invoice {
	return &domain.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: "INV-001",
		ClientID:      uuid.New(),
		InvoiceType:   domain.InvoiceTypeFixedFee,
		Status:        status,
		TotalAmount:   10_000_000,
		TaxAmount:     1_000_000,
		SnapshotData:  []byte("{}"),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		CreatedBy:     uuid.New(),
		UpdatedBy:     uuid.New(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestInvoiceUseCase_Create(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	inv := newInvoice(domain.InvoiceStatusDraft)

	tests := []struct {
		name    string
		repo    *mockInvoiceRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockInvoiceRepo{created: inv},
		},
		{
			name:    "duplicate invoice number",
			repo:    &mockInvoiceRepo{createErr: domain.ErrInvoiceNumberDuplicate},
			wantErr: domain.ErrInvoiceNumberDuplicate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewInvoiceUseCase(tt.repo, &mockLineItemRepo{}, nil, nil)
			resp, err := uc.Create(context.Background(), usecase.InvoiceCreateRequest{
				InvoiceNumber: "INV-001",
				ClientID:      uuid.New(),
				InvoiceType:   domain.InvoiceTypeFixedFee,
				TotalAmount:   10_000_000,
			}, callerID, "")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Status != domain.InvoiceStatusDraft {
				t.Errorf("want DRAFT, got %s", resp.Status)
			}
		})
	}
}

func TestInvoiceUseCase_GetByID(t *testing.T) {
	t.Parallel()
	id := uuid.New()

	tests := []struct {
		name    string
		repo    *mockInvoiceRepo
		wantErr error
	}{
		{name: "success", repo: &mockInvoiceRepo{found: newInvoice(domain.InvoiceStatusDraft)}},
		{name: "not found", repo: &mockInvoiceRepo{findErr: domain.ErrInvoiceNotFound}, wantErr: domain.ErrInvoiceNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewInvoiceUseCase(tt.repo, &mockLineItemRepo{}, nil, nil)
			_, err := uc.GetByID(context.Background(), id)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInvoiceUseCase_Send_StateTransition(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()

	tests := []struct {
		name    string
		found   *domain.Invoice
		wantErr error
	}{
		{
			name:  "DRAFT → SENT succeeds",
			found: newInvoice(domain.InvoiceStatusDraft),
		},
		{
			name:    "ISSUED → SENT invalid",
			found:   newInvoice(domain.InvoiceStatusIssued),
			wantErr: domain.ErrInvalidStateTransition,
		},
		{
			name:    "PAID → SENT invalid",
			found:   newInvoice(domain.InvoiceStatusPaid),
			wantErr: domain.ErrInvalidStateTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sentInvoice := newInvoice(domain.InvoiceStatusSent)
			repo := &mockInvoiceRepo{
				found:     tt.found,
				statusUpd: sentInvoice,
			}
			uc := usecase.NewInvoiceUseCase(repo, &mockLineItemRepo{}, nil, nil)
			_, err := uc.Send(context.Background(), uuid.New(), callerID, "")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInvoiceUseCase_Issue_SnapshotFrozen(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	confirmedInv := newInvoice(domain.InvoiceStatusConfirmed)
	issuedInv := newInvoice(domain.InvoiceStatusIssued)
	repo := &mockInvoiceRepo{
		found:     confirmedInv,
		snapUpd:   confirmedInv,
		statusUpd: issuedInv,
	}
	uc := usecase.NewInvoiceUseCase(repo, &mockLineItemRepo{}, nil, nil)
	resp, err := uc.Issue(context.Background(), uuid.New(), callerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != domain.InvoiceStatusIssued {
		t.Errorf("want ISSUED, got %s", resp.Status)
	}
}

func TestInvoiceUseCase_Issue_WrongStatus(t *testing.T) {
	t.Parallel()
	repo := &mockInvoiceRepo{found: newInvoice(domain.InvoiceStatusDraft)}
	uc := usecase.NewInvoiceUseCase(repo, &mockLineItemRepo{}, nil, nil)
	_, err := uc.Issue(context.Background(), uuid.New(), uuid.New(), "")
	if !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("want INVALID_STATE_TRANSITION, got %v", err)
	}
}

func TestInvoiceUseCase_List(t *testing.T) {
	t.Parallel()
	items := []*domain.Invoice{
		newInvoice(domain.InvoiceStatusDraft),
		newInvoice(domain.InvoiceStatusIssued),
	}
	uc := usecase.NewInvoiceUseCase(&mockInvoiceRepo{listItems: items, listTotal: 2}, &mockLineItemRepo{}, nil, nil)
	result, err := uc.List(context.Background(), usecase.InvoiceListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("want total 2, got %d", result.Total)
	}
}

func TestInvoiceUseCase_AddLineItem_DraftOnly(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	lineItem := &domain.InvoiceLineItem{
		ID:          uuid.New(),
		Description: "Audit services",
		Quantity:    1, UnitPrice: 5_000_000, TotalAmount: 5_000_000,
		SourceType: domain.SourceEngagementFee,
		CreatedAt:  time.Now(),
	}

	tests := []struct {
		name    string
		inv     *domain.Invoice
		wantErr error
	}{
		{name: "success on DRAFT", inv: newInvoice(domain.InvoiceStatusDraft)},
		{name: "fail on ISSUED", inv: newInvoice(domain.InvoiceStatusIssued), wantErr: domain.ErrInvoiceLocked},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			invRepo := &mockInvoiceRepo{found: tt.inv}
			lineRepo := &mockLineItemRepo{added: lineItem}
			uc := usecase.NewInvoiceUseCase(invRepo, lineRepo, nil, nil)
			_, err := uc.AddLineItem(context.Background(), uuid.New(), usecase.LineItemAddRequest{
				Description: "Audit services", Quantity: 1, UnitPrice: 5_000_000,
			}, callerID, "")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("want %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPaymentUseCase_Record_ExceedsBalance(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	inv := newInvoice(domain.InvoiceStatusIssued)
	payRepo := &mockPaymentRepo{sumPaid: 9_000_000} // 9M already paid, total is 10M
	invRepo := &mockInvoiceRepo{found: inv}

	uc := usecase.NewPaymentUseCase(payRepo, invRepo, nil, nil)
	_, err := uc.Record(context.Background(), uuid.New(), usecase.PaymentRecordRequest{
		PaymentMethod: domain.PaymentBankTransfer,
		Amount:        2_000_000, // Would exceed 10M total
		PaymentDate:   time.Now(),
	}, callerID, "")
	if !errors.Is(err, domain.ErrPaymentExceedsBalance) {
		t.Fatalf("want PAYMENT_EXCEEDS_BALANCE, got %v", err)
	}
}

func TestPaymentUseCase_Record_Success(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	inv := newInvoice(domain.InvoiceStatusIssued)
	payment := &domain.Payment{
		ID:            uuid.New(),
		InvoiceID:     inv.ID,
		PaymentMethod: domain.PaymentBankTransfer,
		Amount:        5_000_000,
		Status:        domain.PaymentRecorded,
		RecordedBy:    callerID,
		RecordedAt:    time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	payRepo := &mockPaymentRepo{sumPaid: 0, recorded: payment, statusUpd: payment}
	invRepo := &mockInvoiceRepo{found: inv, statusUpd: inv}

	uc := usecase.NewPaymentUseCase(payRepo, invRepo, nil, nil)
	resp, err := uc.Record(context.Background(), uuid.New(), usecase.PaymentRecordRequest{
		PaymentMethod: domain.PaymentBankTransfer,
		Amount:        5_000_000,
		PaymentDate:   time.Now(),
	}, callerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Amount != 5_000_000 {
		t.Errorf("want amount 5000000, got %v", resp.Amount)
	}
}

func TestPaymentUseCase_Record_NotIssuedInvoice(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	inv := newInvoice(domain.InvoiceStatusDraft)
	payRepo := &mockPaymentRepo{}
	invRepo := &mockInvoiceRepo{found: inv}

	uc := usecase.NewPaymentUseCase(payRepo, invRepo, nil, nil)
	_, err := uc.Record(context.Background(), uuid.New(), usecase.PaymentRecordRequest{
		PaymentMethod: domain.PaymentBankTransfer,
		Amount:        1_000_000,
		PaymentDate:   time.Now(),
	}, callerID, "")
	if !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Fatalf("want INVALID_STATE_TRANSITION, got %v", err)
	}
}

func TestMemoUseCase_Create(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	invoiceID := uuid.New()
	memo := &domain.BillingMemo{
		ID:         uuid.New(),
		MemoType:   domain.MemoCreditNote,
		MemoNumber: "CN-001",
		Amount:     500_000,
		Reason:     "Service adjustment",
		Status:     domain.MemoStatusDraft,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		CreatedBy:  callerID,
		UpdatedBy:  callerID,
	}
	memoRepo := &mockMemoRepo{created: memo}
	invRepo := &mockInvoiceRepo{found: newInvoice(domain.InvoiceStatusIssued)}
	uc := usecase.NewMemoUseCase(memoRepo, invRepo, nil, nil)

	resp, err := uc.Create(context.Background(), invoiceID, usecase.MemoCreateRequest{
		MemoType:   domain.MemoCreditNote,
		MemoNumber: "CN-001",
		Amount:     500_000,
		Reason:     "Service adjustment",
	}, callerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.MemoNumber != "CN-001" {
		t.Errorf("want memo number CN-001, got %s", resp.MemoNumber)
	}
}

func TestMemoUseCase_List(t *testing.T) {
	t.Parallel()
	memos := []*domain.BillingMemo{
		{ID: uuid.New(), MemoType: domain.MemoCreditNote, Status: domain.MemoStatusDraft, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	memoRepo := &mockMemoRepo{items: memos, total: 1}
	invRepo := &mockInvoiceRepo{}
	uc := usecase.NewMemoUseCase(memoRepo, invRepo, nil, nil)

	result, err := uc.List(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("want total 1, got %d", result.Total)
	}
}

// ── Payment Clear/Dispute Tests ───────────────────────────────────────────────

func TestPaymentUseCase_ClearPayment_Success(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	recorded := &domain.Payment{
		ID: uuid.New(), Status: domain.PaymentRecorded,
		RecordedBy: callerID, RecordedAt: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	payRepo := &mockPaymentRepo{found: recorded}
	invRepo := &mockInvoiceRepo{}
	uc := usecase.NewPaymentUseCase(payRepo, invRepo, nil, nil)

	resp, err := uc.ClearPayment(context.Background(), recorded.ID, callerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != domain.PaymentCleared {
		t.Errorf("want CLEARED, got %s", resp.Status)
	}
}

func TestPaymentUseCase_ClearPayment_NotFound(t *testing.T) {
	t.Parallel()
	payRepo := &mockPaymentRepo{statusErr: domain.ErrPaymentNotFound}
	uc := usecase.NewPaymentUseCase(payRepo, &mockInvoiceRepo{}, nil, nil)

	_, err := uc.ClearPayment(context.Background(), uuid.New(), uuid.New(), "")
	if !errors.Is(err, domain.ErrPaymentNotFound) {
		t.Fatalf("want PAYMENT_NOT_FOUND, got %v", err)
	}
}

func TestPaymentUseCase_DisputePayment_Success(t *testing.T) {
	t.Parallel()
	callerID := uuid.New()
	id := uuid.New()
	cleared := &domain.Payment{
		ID: id, Status: domain.PaymentCleared,
		RecordedBy: callerID, RecordedAt: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	disputed := &domain.Payment{
		ID: id, Status: domain.PaymentDisputed,
		RecordedBy: callerID, RecordedAt: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	payRepo := &mockPaymentRepo{found: cleared, statusUpd: disputed}
	uc := usecase.NewPaymentUseCase(payRepo, &mockInvoiceRepo{}, nil, nil)

	resp, err := uc.DisputePayment(context.Background(), id, callerID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != domain.PaymentDisputed {
		t.Errorf("want DISPUTED, got %s", resp.Status)
	}
}

func TestPaymentUseCase_DisputePayment_NotCleared(t *testing.T) {
	t.Parallel()
	recorded := &domain.Payment{ID: uuid.New(), Status: domain.PaymentRecorded}
	payRepo := &mockPaymentRepo{found: recorded}
	uc := usecase.NewPaymentUseCase(payRepo, &mockInvoiceRepo{}, nil, nil)

	_, err := uc.DisputePayment(context.Background(), uuid.New(), uuid.New(), "")
	if !errors.Is(err, domain.ErrPaymentNotCleared) {
		t.Fatalf("want PAYMENT_NOT_CLEARED, got %v", err)
	}
}

// ── AR Tests ──────────────────────────────────────────────────────────────────

type mockARRepo struct {
	aging       []*domain.ARAgingRow
	outstanding []*domain.AROutstandingRow
	err         error
}

func (m *mockARRepo) GetAging(_ context.Context) ([]*domain.ARAgingRow, error) {
	return m.aging, m.err
}
func (m *mockARRepo) GetOutstanding(_ context.Context) ([]*domain.AROutstandingRow, error) {
	return m.outstanding, m.err
}

func TestARUseCase_GetAging(t *testing.T) {
	t.Parallel()
	clientID := uuid.New()
	arRepo := &mockARRepo{aging: []*domain.ARAgingRow{
		{ClientID: clientID, Current: 5_000_000, TotalOutstanding: 5_000_000},
	}}
	uc := usecase.NewARUseCase(arRepo)

	result, err := uc.GetAging(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("want 1 row, got %d", len(result))
	}
	if result[0].ClientID != clientID {
		t.Errorf("wrong client ID")
	}
	if result[0].TotalOutstanding != 5_000_000 {
		t.Errorf("wrong total outstanding: %v", result[0].TotalOutstanding)
	}
}

func TestARUseCase_GetOutstanding(t *testing.T) {
	t.Parallel()
	clientID := uuid.New()
	arRepo := &mockARRepo{outstanding: []*domain.AROutstandingRow{
		{ClientID: clientID, InvoiceCount: 3, TotalBilled: 30_000_000, TotalPaid: 10_000_000, TotalOutstanding: 20_000_000},
	}}
	uc := usecase.NewARUseCase(arRepo)

	result, err := uc.GetOutstanding(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("want 1 row, got %d", len(result))
	}
	if result[0].TotalOutstanding != 20_000_000 {
		t.Errorf("wrong outstanding: %v", result[0].TotalOutstanding)
	}
}

// ── Report Tests ───────────────────────────────────────────────────────────────

type mockReportRepo struct {
	periodSummary  *domain.BillingPeriodSummary
	paymentSummary *domain.PaymentSummary
	invoices       []*domain.Invoice
	err            error
}

func (m *mockReportRepo) GetPeriodSummary(_ context.Context, _, _ time.Time) (*domain.BillingPeriodSummary, error) {
	return m.periodSummary, m.err
}
func (m *mockReportRepo) GetPaymentSummary(_ context.Context, _, _ time.Time) (*domain.PaymentSummary, error) {
	return m.paymentSummary, m.err
}
func (m *mockReportRepo) ListInvoicesForExport(_ context.Context, _ domain.ListInvoicesFilter) ([]*domain.Invoice, error) {
	return m.invoices, m.err
}

func TestReportUseCase_GetPeriodSummary(t *testing.T) {
	t.Parallel()
	now := time.Now()
	repo := &mockReportRepo{periodSummary: &domain.BillingPeriodSummary{
		PeriodStart:   now,
		PeriodEnd:     now.Add(24 * time.Hour),
		TotalInvoiced: 100_000,
		TotalPaid:     60_000,
		InvoiceCount:  5,
		PaidCount:     3,
		ByStatus:      []domain.StatusCount{},
	}}
	uc := usecase.NewReportUseCase(repo)

	req := usecase.PeriodSummaryRequest{Start: now, End: now.Add(24 * time.Hour)}
	result, err := uc.GetPeriodSummary(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalInvoiced != 100_000 {
		t.Errorf("want 100000, got %v", result.TotalInvoiced)
	}
	if result.InvoiceCount != 5 {
		t.Errorf("want 5, got %d", result.InvoiceCount)
	}
}

func TestReportUseCase_GetPeriodSummary_Error(t *testing.T) {
	t.Parallel()
	repo := &mockReportRepo{err: errors.New("db error")}
	uc := usecase.NewReportUseCase(repo)

	_, err := uc.GetPeriodSummary(context.Background(), usecase.PeriodSummaryRequest{
		Start: time.Now(), End: time.Now().Add(time.Hour),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReportUseCase_ExportInvoicesCSV(t *testing.T) {
	t.Parallel()
	invID := uuid.New()
	clientID := uuid.New()
	callerID := uuid.New()
	now := time.Now()
	inv := &domain.Invoice{
		ID:            invID,
		InvoiceNumber: "INV-001",
		ClientID:      clientID,
		InvoiceType:   domain.InvoiceTypeFixedFee,
		Status:        domain.InvoiceStatusIssued,
		TotalAmount:   5_000_000,
		TaxAmount:     500_000,
		CreatedAt:     now,
		CreatedBy:     callerID,
	}
	repo := &mockReportRepo{invoices: []*domain.Invoice{inv}}
	uc := usecase.NewReportUseCase(repo)

	data, err := uc.ExportInvoicesCSV(context.Background(), domain.ListInvoicesFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected CSV data, got empty")
	}
	csv := string(data)
	if !contains(csv, "INV-001") {
		t.Errorf("CSV missing invoice number: %s", csv)
	}
	if !contains(csv, "5000000.00") {
		t.Errorf("CSV missing total amount: %s", csv)
	}
}

func TestReportUseCase_ExportInvoicesCSV_Empty(t *testing.T) {
	t.Parallel()
	repo := &mockReportRepo{invoices: []*domain.Invoice{}}
	uc := usecase.NewReportUseCase(repo)

	data, err := uc.ExportInvoicesCSV(context.Background(), domain.ListInvoicesFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should still have header row
	if !contains(string(data), "invoice_number") {
		t.Errorf("CSV missing header: %s", string(data))
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
