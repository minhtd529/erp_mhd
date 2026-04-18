package usecase_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/tax/domain"
	"github.com/mdh/erp-audit/api/internal/tax/usecase"
)

// ── mock deadline repo ────────────────────────────────────────────────────────

type mockDeadlineRepo struct {
	created  *domain.TaxDeadline
	found    *domain.TaxDeadline
	list     []*domain.TaxDeadline
	total    int64
	err      error
}

func (m *mockDeadlineRepo) Create(_ context.Context, p domain.CreateDeadlineParams) (*domain.TaxDeadline, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.created != nil {
		return m.created, nil
	}
	return &domain.TaxDeadline{
		ID:           uuid.New(),
		ClientID:     p.ClientID,
		DeadlineType: p.DeadlineType,
		DeadlineName: p.DeadlineName,
		DueDate:      p.DueDate,
		Status:       domain.DeadlineStatusNotDue,
		CreatedBy:    p.CreatedBy,
	}, nil
}
func (m *mockDeadlineRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.TaxDeadline, error) {
	return m.found, m.err
}
func (m *mockDeadlineRepo) List(_ context.Context, _ domain.ListDeadlinesFilter, _, _ int) ([]*domain.TaxDeadline, int64, error) {
	return m.list, m.total, m.err
}
func (m *mockDeadlineRepo) Update(_ context.Context, p domain.UpdateDeadlineParams) (*domain.TaxDeadline, error) {
	if m.err != nil {
		return nil, m.err
	}
	d := m.found
	if d == nil {
		d = &domain.TaxDeadline{ID: p.ID}
	}
	d.DeadlineName = p.DeadlineName
	return d, nil
}
func (m *mockDeadlineRepo) MarkCompleted(_ context.Context, id uuid.UUID, _ time.Time, _ uuid.UUID) (*domain.TaxDeadline, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.TaxDeadline{ID: id, Status: domain.DeadlineStatusCompleted}, nil
}
func (m *mockDeadlineRepo) UpdateStatus(_ context.Context, _ uuid.UUID, _ domain.DeadlineStatus) error {
	return m.err
}
func (m *mockDeadlineRepo) ListDueSoon(_ context.Context, _ time.Time) ([]*domain.TaxDeadline, error) {
	return m.list, m.err
}
func (m *mockDeadlineRepo) ListOverdue(_ context.Context) ([]*domain.TaxDeadline, error) {
	return m.list, m.err
}

// ── mock advisory repo ────────────────────────────────────────────────────────

type mockAdvisoryRepo struct {
	created *domain.AdvisoryRecord
	found   *domain.AdvisoryRecord
	list    []*domain.AdvisoryRecord
	total   int64
	file    *domain.AdvisoryFile
	files   []*domain.AdvisoryFile
	err     error
}

func (m *mockAdvisoryRepo) Create(_ context.Context, p domain.CreateAdvisoryParams) (*domain.AdvisoryRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.created != nil {
		return m.created, nil
	}
	return &domain.AdvisoryRecord{
		ID:             uuid.New(),
		ClientID:       p.ClientID,
		AdvisoryType:   p.AdvisoryType,
		Recommendation: p.Recommendation,
		Status:         domain.AdvisoryDrafted,
		CreatedBy:      p.CreatedBy,
	}, nil
}
func (m *mockAdvisoryRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.AdvisoryRecord, error) {
	return m.found, m.err
}
func (m *mockAdvisoryRepo) List(_ context.Context, _ domain.ListAdvisoryFilter, _, _ int) ([]*domain.AdvisoryRecord, int64, error) {
	return m.list, m.total, m.err
}
func (m *mockAdvisoryRepo) Update(_ context.Context, p domain.UpdateAdvisoryParams) (*domain.AdvisoryRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.AdvisoryRecord{ID: p.ID, Recommendation: p.Recommendation, Status: domain.AdvisoryDrafted}, nil
}
func (m *mockAdvisoryRepo) Deliver(_ context.Context, id uuid.UUID, _ uuid.UUID) (*domain.AdvisoryRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &domain.AdvisoryRecord{ID: id, Status: domain.AdvisoryDelivered}, nil
}
func (m *mockAdvisoryRepo) AttachFile(_ context.Context, p domain.AttachFileParams) (*domain.AdvisoryFile, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.file != nil {
		return m.file, nil
	}
	return &domain.AdvisoryFile{ID: uuid.New(), AdvisoryID: p.AdvisoryID, FileName: p.FileName}, nil
}
func (m *mockAdvisoryRepo) ListFiles(_ context.Context, _ uuid.UUID) ([]*domain.AdvisoryFile, error) {
	return m.files, m.err
}

// ── mock compliance repo ──────────────────────────────────────────────────────

type mockComplianceRepo struct {
	status    *domain.ComplianceStatus
	deadlines []*domain.TaxDeadline
	err       error
}

func (m *mockComplianceRepo) GetComplianceStatus(_ context.Context, _ uuid.UUID) (*domain.ComplianceStatus, error) {
	return m.status, m.err
}
func (m *mockComplianceRepo) ListAllOverdue(_ context.Context) ([]*domain.TaxDeadline, error) {
	return m.deadlines, m.err
}
func (m *mockComplianceRepo) DashboardDeadlines(_ context.Context, _, _ time.Time) ([]*domain.TaxDeadline, error) {
	return m.deadlines, m.err
}

// ── Deadline tests ────────────────────────────────────────────────────────────

func TestTaxDeadlineUseCase_Create_Success(t *testing.T) {
	t.Parallel()
	clientID := uuid.New()
	callerID := uuid.New()
	repo := &mockDeadlineRepo{}
	uc := usecase.NewTaxDeadlineUseCase(repo, nil)

	d, err := uc.Create(context.Background(), clientID, usecase.CreateDeadlineRequest{
		DeadlineType: domain.DeadlineVATFiling,
		DeadlineName: "VAT Q1",
		DueDate:      time.Now().Add(30 * 24 * time.Hour),
	}, callerID, net.ParseIP("127.0.0.1"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.DeadlineType != domain.DeadlineVATFiling {
		t.Errorf("wrong deadline type: %s", d.DeadlineType)
	}
	if d.Status != domain.DeadlineStatusNotDue {
		t.Errorf("want NOT_DUE, got %s", d.Status)
	}
}

func TestTaxDeadlineUseCase_Create_Error(t *testing.T) {
	t.Parallel()
	repo := &mockDeadlineRepo{err: errors.New("db error")}
	uc := usecase.NewTaxDeadlineUseCase(repo, nil)

	_, err := uc.Create(context.Background(), uuid.New(), usecase.CreateDeadlineRequest{
		DeadlineType: domain.DeadlineVATFiling,
		DeadlineName: "VAT Q1",
		DueDate:      time.Now(),
	}, uuid.New(), net.ParseIP("127.0.0.1"))

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTaxDeadlineUseCase_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := &mockDeadlineRepo{err: domain.ErrTaxDeadlineNotFound}
	uc := usecase.NewTaxDeadlineUseCase(repo, nil)

	_, err := uc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrTaxDeadlineNotFound) {
		t.Errorf("want ErrTaxDeadlineNotFound, got %v", err)
	}
}

func TestTaxDeadlineUseCase_List(t *testing.T) {
	t.Parallel()
	repo := &mockDeadlineRepo{
		list: []*domain.TaxDeadline{
			{ID: uuid.New(), DeadlineName: "D1", Status: domain.DeadlineStatusNotDue},
			{ID: uuid.New(), DeadlineName: "D2", Status: domain.DeadlineStatusDueSoon},
		},
		total: 2,
	}
	uc := usecase.NewTaxDeadlineUseCase(repo, nil)

	result, err := uc.List(context.Background(), uuid.New(), "", 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("want total 2, got %d", result.Total)
	}
}

func TestTaxDeadlineUseCase_MarkCompleted_AlreadyCompleted(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	repo := &mockDeadlineRepo{
		found: &domain.TaxDeadline{ID: id, Status: domain.DeadlineStatusCompleted},
	}
	uc := usecase.NewTaxDeadlineUseCase(repo, nil)

	_, err := uc.MarkCompleted(context.Background(), id, usecase.MarkCompletedRequest{
		ActualDate: time.Now(),
	}, uuid.New(), net.ParseIP("127.0.0.1"))

	if !errors.Is(err, domain.ErrInvalidStateTransition) {
		t.Errorf("want ErrInvalidStateTransition, got %v", err)
	}
}

func TestTaxDeadlineUseCase_MarkCompleted_Success(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	repo := &mockDeadlineRepo{
		found: &domain.TaxDeadline{ID: id, Status: domain.DeadlineStatusNotDue},
	}
	uc := usecase.NewTaxDeadlineUseCase(repo, nil)

	d, err := uc.MarkCompleted(context.Background(), id, usecase.MarkCompletedRequest{
		ActualDate: time.Now(),
	}, uuid.New(), net.ParseIP("127.0.0.1"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Status != domain.DeadlineStatusCompleted {
		t.Errorf("want COMPLETED, got %s", d.Status)
	}
}

func TestTaxDeadlineUseCase_AutoGenerate_NoFiscalYear(t *testing.T) {
	t.Parallel()
	repo := &mockDeadlineRepo{}
	uc := usecase.NewTaxDeadlineUseCase(repo, nil)

	_, err := uc.AutoGenerate(context.Background(), uuid.New(), "", 2026, uuid.New(), net.ParseIP("127.0.0.1"))
	if !errors.Is(err, domain.ErrFiscalYearNotConfigured) {
		t.Errorf("want ErrFiscalYearNotConfigured, got %v", err)
	}
}

func TestTaxDeadlineUseCase_AutoGenerate_Success(t *testing.T) {
	t.Parallel()
	repo := &mockDeadlineRepo{}
	uc := usecase.NewTaxDeadlineUseCase(repo, nil)

	deadlines, err := uc.AutoGenerate(context.Background(), uuid.New(), "12-31", 2026, uuid.New(), net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deadlines) != 6 {
		t.Errorf("want 6 standard deadlines, got %d", len(deadlines))
	}
}

func TestTaxDeadlineUseCase_RefreshOverdue(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	repo := &mockDeadlineRepo{
		list: []*domain.TaxDeadline{{ID: id, Status: domain.DeadlineStatusNotDue}},
	}
	uc := usecase.NewTaxDeadlineUseCase(repo, nil)

	if err := uc.RefreshOverdue(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── Advisory tests ────────────────────────────────────────────────────────────

func TestAdvisoryUseCase_Create_Success(t *testing.T) {
	t.Parallel()
	clientID := uuid.New()
	callerID := uuid.New()
	repo := &mockAdvisoryRepo{}
	uc := usecase.NewAdvisoryUseCase(repo, nil)

	a, err := uc.Create(context.Background(), clientID, usecase.CreateAdvisoryRequest{
		AdvisoryType:   domain.AdvisoryTaxConsultation,
		Recommendation: "Optimize VAT deductions",
	}, callerID, net.ParseIP("127.0.0.1"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.AdvisoryType != domain.AdvisoryTaxConsultation {
		t.Errorf("wrong type: %s", a.AdvisoryType)
	}
	if a.Status != domain.AdvisoryDrafted {
		t.Errorf("want DRAFTED, got %s", a.Status)
	}
}

func TestAdvisoryUseCase_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := &mockAdvisoryRepo{err: domain.ErrAdvisoryRecordNotFound}
	uc := usecase.NewAdvisoryUseCase(repo, nil)

	_, err := uc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrAdvisoryRecordNotFound) {
		t.Errorf("want ErrAdvisoryRecordNotFound, got %v", err)
	}
}

func TestAdvisoryUseCase_Deliver_Success(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	repo := &mockAdvisoryRepo{}
	uc := usecase.NewAdvisoryUseCase(repo, nil)

	a, err := uc.Deliver(context.Background(), id, uuid.New(), net.ParseIP("127.0.0.1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Status != domain.AdvisoryDelivered {
		t.Errorf("want DELIVERED, got %s", a.Status)
	}
}

func TestAdvisoryUseCase_Deliver_NotDeliverable(t *testing.T) {
	t.Parallel()
	repo := &mockAdvisoryRepo{err: domain.ErrAdvisoryNotDeliverable}
	uc := usecase.NewAdvisoryUseCase(repo, nil)

	_, err := uc.Deliver(context.Background(), uuid.New(), uuid.New(), net.ParseIP("127.0.0.1"))
	if !errors.Is(err, domain.ErrAdvisoryNotDeliverable) {
		t.Errorf("want ErrAdvisoryNotDeliverable, got %v", err)
	}
}

func TestAdvisoryUseCase_AttachFile(t *testing.T) {
	t.Parallel()
	advisoryID := uuid.New()
	repo := &mockAdvisoryRepo{}
	uc := usecase.NewAdvisoryUseCase(repo, nil)

	f, err := uc.AttachFile(context.Background(), advisoryID, usecase.AttachFileRequest{
		FileName: "report.pdf",
		FilePath: "/files/tax/report.pdf",
	}, uuid.New(), net.ParseIP("127.0.0.1"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.FileName != "report.pdf" {
		t.Errorf("wrong file name: %s", f.FileName)
	}
}

// ── Compliance tests ──────────────────────────────────────────────────────────

func TestComplianceUseCase_GetStatus(t *testing.T) {
	t.Parallel()
	clientID := uuid.New()
	repo := &mockComplianceRepo{
		status: &domain.ComplianceStatus{
			ClientID:        clientID,
			TotalDeadlines:  10,
			Completed:       8,
			Overdue:         1,
			ComplianceScore: 90,
		},
	}
	uc := usecase.NewComplianceUseCase(repo)

	s, err := uc.GetClientComplianceStatus(context.Background(), clientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ComplianceScore != 90 {
		t.Errorf("want score 90, got %d", s.ComplianceScore)
	}
}

func TestComplianceUseCase_ListOverdueAlerts(t *testing.T) {
	t.Parallel()
	repo := &mockComplianceRepo{
		deadlines: []*domain.TaxDeadline{
			{ID: uuid.New(), Status: domain.DeadlineStatusOverdue},
			{ID: uuid.New(), Status: domain.DeadlineStatusOverdue},
		},
	}
	uc := usecase.NewComplianceUseCase(repo)

	deadlines, err := uc.ListOverdueAlerts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deadlines) != 2 {
		t.Errorf("want 2 overdue, got %d", len(deadlines))
	}
}

func TestComplianceUseCase_Dashboard(t *testing.T) {
	t.Parallel()
	repo := &mockComplianceRepo{deadlines: []*domain.TaxDeadline{}}
	uc := usecase.NewComplianceUseCase(repo)

	deadlines, err := uc.DashboardDeadlines(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deadlines == nil {
		t.Error("want non-nil deadlines slice")
	}
}
