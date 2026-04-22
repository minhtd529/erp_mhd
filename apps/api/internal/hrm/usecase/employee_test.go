package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/internal/hrm/usecase"
)

// ─── mock EmployeeRepository ─────────────────────────────────────────────────

type mockEmpRepo struct {
	emp    *domain.Employee
	emps   []*domain.Employee
	total  int64
	err    error
}

func (m *mockEmpRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Employee, error) {
	return m.emp, m.err
}
func (m *mockEmpRepo) FindByUserID(_ context.Context, _ uuid.UUID) (*domain.Employee, error) {
	return m.emp, m.err
}
func (m *mockEmpRepo) List(_ context.Context, _ domain.ListEmployeesFilter) ([]*domain.Employee, int64, error) {
	return m.emps, m.total, m.err
}
func (m *mockEmpRepo) Create(_ context.Context, _ domain.CreateEmployeeParams) (*domain.Employee, error) {
	return m.emp, m.err
}
func (m *mockEmpRepo) Update(_ context.Context, _ domain.UpdateEmployeeParams) (*domain.Employee, error) {
	return m.emp, m.err
}
func (m *mockEmpRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return m.err
}
func (m *mockEmpRepo) UpdateProfile(_ context.Context, _ domain.UpdateProfileParams) (*domain.Employee, error) {
	return m.emp, m.err
}
func (m *mockEmpRepo) UpdateSensitiveFields(_ context.Context, _ domain.UpdateSensitiveParams) error {
	return m.err
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func newEmpUC(repo domain.EmployeeRepository) *usecase.EmployeeUseCase {
	return usecase.NewEmployeeUseCase(repo, nil)
}

func sampleEmployee() *domain.Employee {
	uid := uuid.New()
	bid := uuid.New()
	return &domain.Employee{
		ID:             uuid.New(),
		FullName:       "Nguyễn Văn A",
		Email:          "vana@example.com",
		Grade:          "SENIOR",
		Status:         "ACTIVE",
		EmploymentType: "FULL_TIME",
		WorkLocation:   "OFFICE",
		CommissionType: "NONE",
		UserID:         &uid,
		BranchID:       &bid,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestEmployeeUseCase_List_ScopeAll(t *testing.T) {
	t.Parallel()
	e := sampleEmployee()
	uc := newEmpUC(&mockEmpRepo{emps: []*domain.Employee{e}, total: 1})

	callerID := uuid.New()
	result, err := uc.ListEmployees(context.Background(),
		usecase.ListEmployeeRequest{Page: 1, Size: 20},
		callerID, []string{"HR_MANAGER"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("got total %d, want 1", result.Total)
	}
}

func TestEmployeeUseCase_List_ScopeSelf(t *testing.T) {
	t.Parallel()
	e := sampleEmployee()
	uc := newEmpUC(&mockEmpRepo{emps: []*domain.Employee{e}, total: 1})

	callerID := uuid.New()
	// JUNIOR_AUDITOR gets self-only scope
	result, err := uc.ListEmployees(context.Background(),
		usecase.ListEmployeeRequest{Page: 1, Size: 20},
		callerID, []string{"JUNIOR_AUDITOR"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Result depends on mock returning all; scope filter is passed to repo which is mocked
	if result.Total != 1 {
		t.Errorf("got total %d, want 1", result.Total)
	}
}

func TestEmployeeUseCase_Get_Success(t *testing.T) {
	t.Parallel()
	e := sampleEmployee()
	uc := newEmpUC(&mockEmpRepo{emp: e})

	callerID := uuid.New()
	resp, err := uc.GetEmployee(context.Background(), e.ID, callerID, []string{"HR_MANAGER"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != e.ID {
		t.Errorf("got id %v, want %v", resp.ID, e.ID)
	}
}

func TestEmployeeUseCase_Get_NotFound(t *testing.T) {
	t.Parallel()
	uc := newEmpUC(&mockEmpRepo{err: domain.ErrEmployeeNotFound})

	callerID := uuid.New()
	_, err := uc.GetEmployee(context.Background(), uuid.New(), callerID, []string{"HR_MANAGER"}, nil)
	if err != domain.ErrEmployeeNotFound {
		t.Errorf("expected ErrEmployeeNotFound, got %v", err)
	}
}

func TestEmployeeUseCase_Get_ScopeBranch_OwnBranch_OK(t *testing.T) {
	t.Parallel()
	e := sampleEmployee()
	uc := newEmpUC(&mockEmpRepo{emp: e})

	callerID := uuid.New()
	branchID := *e.BranchID // same branch
	resp, err := uc.GetEmployee(context.Background(), e.ID, callerID, []string{"HEAD_OF_BRANCH"}, &branchID)
	if err != nil {
		t.Fatalf("HoB own-branch should succeed, got: %v", err)
	}
	if resp.ID != e.ID {
		t.Errorf("got id %v, want %v", resp.ID, e.ID)
	}
}

func TestEmployeeUseCase_Get_ScopeBranch_OtherBranch_Forbidden(t *testing.T) {
	t.Parallel()
	e := sampleEmployee()
	uc := newEmpUC(&mockEmpRepo{emp: e})

	callerID := uuid.New()
	otherBranch := uuid.New()
	_, err := uc.GetEmployee(context.Background(), e.ID, callerID, []string{"HEAD_OF_BRANCH"}, &otherBranch)
	if err != domain.ErrInsufficientPermission {
		t.Errorf("expected ErrInsufficientPermission, got %v", err)
	}
}

func TestEmployeeUseCase_Get_ScopeSelf_OwnRecord_OK(t *testing.T) {
	t.Parallel()
	e := sampleEmployee()
	uc := newEmpUC(&mockEmpRepo{emp: e})

	callerID := *e.UserID // same user
	resp, err := uc.GetEmployee(context.Background(), e.ID, callerID, []string{"JUNIOR_AUDITOR"}, nil)
	if err != nil {
		t.Fatalf("self-only own-record should succeed, got: %v", err)
	}
	if resp.ID != e.ID {
		t.Errorf("got id %v, want %v", resp.ID, e.ID)
	}
}

func TestEmployeeUseCase_Get_ScopeSelf_OtherRecord_Forbidden(t *testing.T) {
	t.Parallel()
	e := sampleEmployee()
	uc := newEmpUC(&mockEmpRepo{emp: e})

	callerID := uuid.New() // different user
	_, err := uc.GetEmployee(context.Background(), e.ID, callerID, []string{"JUNIOR_AUDITOR"}, nil)
	if err != domain.ErrInsufficientPermission {
		t.Errorf("expected ErrInsufficientPermission, got %v", err)
	}
}

func TestEmployeeUseCase_Create_Success(t *testing.T) {
	t.Parallel()
	e := sampleEmployee()
	uc := newEmpUC(&mockEmpRepo{emp: e})

	callerID := uuid.New()
	resp, err := uc.CreateEmployee(context.Background(), usecase.CreateEmployeeRequest{
		FullName:       "Nguyễn Văn A",
		Email:          "vana@example.com",
		Grade:          "SENIOR",
		Status:         "ACTIVE",
		EmploymentType: "FULL_TIME",
		WorkLocation:   "OFFICE",
		CommissionType: "NONE",
	}, &callerID, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Email != e.Email {
		t.Errorf("got email %q, want %q", resp.Email, e.Email)
	}
}

func TestEmployeeUseCase_Create_EmailConflict(t *testing.T) {
	t.Parallel()
	uc := newEmpUC(&mockEmpRepo{err: domain.ErrEmployeeEmailConflict})

	callerID := uuid.New()
	_, err := uc.CreateEmployee(context.Background(), usecase.CreateEmployeeRequest{
		FullName:       "Test",
		Email:          "dup@example.com",
		Grade:          "JUNIOR",
		Status:         "ACTIVE",
		EmploymentType: "FULL_TIME",
		WorkLocation:   "OFFICE",
		CommissionType: "NONE",
	}, &callerID, "127.0.0.1")
	if err != domain.ErrEmployeeEmailConflict {
		t.Errorf("expected ErrEmployeeEmailConflict, got %v", err)
	}
}

func TestEmployeeUseCase_Delete_Success(t *testing.T) {
	t.Parallel()
	uc := newEmpUC(&mockEmpRepo{err: nil})

	callerID := uuid.New()
	err := uc.DeleteEmployee(context.Background(), uuid.New(), &callerID, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmployeeUseCase_SensitiveFieldsMasked(t *testing.T) {
	t.Parallel()
	e := sampleEmployee()
	cc := "1234567890"
	e.CccdEncrypted = &cc
	bank := "secret-bank-acc"
	e.BankAccountEncrypted = &bank
	uc := newEmpUC(&mockEmpRepo{emp: e})

	callerID := uuid.New()
	resp, err := uc.GetEmployee(context.Background(), e.ID, callerID, []string{"HR_MANAGER"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CccdEncrypted != "***" {
		t.Errorf("expected CccdEncrypted = ***, got %q", resp.CccdEncrypted)
	}
	if resp.BankAccountEncrypted != "***" {
		t.Errorf("expected BankAccountEncrypted = ***, got %q", resp.BankAccountEncrypted)
	}
	if resp.MstCaNhanEncrypted != "***" {
		t.Errorf("expected MstCaNhanEncrypted = ***, got %q", resp.MstCaNhanEncrypted)
	}
	if resp.SoBhxhEncrypted != "***" {
		t.Errorf("expected SoBhxhEncrypted = ***, got %q", resp.SoBhxhEncrypted)
	}
}
