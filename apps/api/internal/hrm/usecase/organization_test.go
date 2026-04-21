package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/internal/hrm/usecase"
)

// ─── mock repo ────────────────────────────────────────────────────────────────

type mockOrgRepo struct {
	branch    *domain.HRMBranch
	branches  []*domain.HRMBranch
	dept      *domain.HRMDepartment
	depts     []*domain.HRMDepartment
	bd        *domain.BranchDepartment
	bds       []*domain.BranchDepartment
	chartData []*domain.OrgChartBranch

	branchErr    error
	deptErr      error
	bdErr        error
	softDeleteErr error
}

func (m *mockOrgRepo) FindBranchByID(_ context.Context, _ uuid.UUID) (*domain.HRMBranch, error) {
	return m.branch, m.branchErr
}
func (m *mockOrgRepo) ListBranches(_ context.Context, _ domain.ListHRMBranchesFilter) ([]*domain.HRMBranch, int64, error) {
	return m.branches, int64(len(m.branches)), m.branchErr
}
func (m *mockOrgRepo) UpdateBranch(_ context.Context, _ domain.UpdateHRMBranchParams) (*domain.HRMBranch, error) {
	return m.branch, m.branchErr
}
func (m *mockOrgRepo) FindDeptByID(_ context.Context, _ uuid.UUID) (*domain.HRMDepartment, error) {
	return m.dept, m.deptErr
}
func (m *mockOrgRepo) ListDepts(_ context.Context, _ domain.ListHRMDeptsFilter) ([]*domain.HRMDepartment, int64, error) {
	return m.depts, int64(len(m.depts)), m.deptErr
}
func (m *mockOrgRepo) UpdateDept(_ context.Context, _ domain.UpdateHRMDeptParams) (*domain.HRMDepartment, error) {
	return m.dept, m.deptErr
}
func (m *mockOrgRepo) ListBranchDepts(_ context.Context, _ domain.ListBranchDeptsFilter) ([]*domain.BranchDepartment, int64, error) {
	return m.bds, int64(len(m.bds)), m.bdErr
}
func (m *mockOrgRepo) CreateBranchDept(_ context.Context, _ domain.CreateBranchDeptParams) (*domain.BranchDepartment, error) {
	return m.bd, m.bdErr
}
func (m *mockOrgRepo) SoftDeleteBranchDept(_ context.Context, _, _ uuid.UUID) error {
	return m.softDeleteErr
}
func (m *mockOrgRepo) AssignBranchHead(_ context.Context, _ domain.AssignBranchHeadParams) (*domain.HRMBranch, error) {
	return m.branch, m.branchErr
}
func (m *mockOrgRepo) DeactivateBranch(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return m.branchErr
}
func (m *mockOrgRepo) AssignDeptHead(_ context.Context, _ domain.AssignDeptHeadParams) (*domain.HRMDepartment, error) {
	return m.dept, m.deptErr
}
func (m *mockOrgRepo) DeactivateDept(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return m.deptErr
}
func (m *mockOrgRepo) ListBranchesWithDepts(_ context.Context) ([]*domain.OrgChartBranch, error) {
	return m.chartData, m.branchErr
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func newTestOrgUC(repo domain.OrgRepository) *usecase.OrgUseCase {
	return usecase.NewOrgUseCase(repo, nil)
}

func sampleBranch() *domain.HRMBranch {
	city := "Hà Nội"
	return &domain.HRMBranch{
		ID:           uuid.New(),
		Code:         "HO",
		Name:         "Trụ sở chính",
		IsActive:     true,
		IsHeadOffice: true,
		City:         &city,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func sampleDept() *domain.HRMDepartment {
	return &domain.HRMDepartment{
		ID:        uuid.New(),
		Code:      "AUDIT",
		Name:      "Phòng Kiểm toán",
		IsActive:  true,
		IsDeleted: false,
		DeptType:  "CORE",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func sampleBD() *domain.BranchDepartment {
	return &domain.BranchDepartment{
		BranchID:     uuid.New(),
		DepartmentID: uuid.New(),
		IsActive:     true,
		CreatedAt:    time.Now(),
	}
}

// ─── Branch tests ─────────────────────────────────────────────────────────────

func TestOrgUseCase_GetBranch_Success(t *testing.T) {
	t.Parallel()
	b := sampleBranch()
	uc := newTestOrgUC(&mockOrgRepo{branch: b})

	resp, err := uc.GetBranch(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != b.ID {
		t.Errorf("got id %v, want %v", resp.ID, b.ID)
	}
	if resp.IsHeadOffice != true {
		t.Error("expected IsHeadOffice = true")
	}
}

func TestOrgUseCase_GetBranch_NotFound(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{branchErr: domain.ErrBranchNotFound})

	_, err := uc.GetBranch(context.Background(), uuid.New())
	if err != domain.ErrBranchNotFound {
		t.Errorf("expected ErrBranchNotFound, got %v", err)
	}
}

func TestOrgUseCase_ListBranches_ReturnsAll(t *testing.T) {
	t.Parallel()
	branches := []*domain.HRMBranch{sampleBranch(), sampleBranch()}
	uc := newTestOrgUC(&mockOrgRepo{branches: branches})

	result, err := uc.ListBranches(context.Background(), usecase.ListBranchHRMRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if int(result.Total) != len(branches) {
		t.Errorf("got total %d, want %d", result.Total, len(branches))
	}
}

func TestOrgUseCase_UpdateBranch_Success(t *testing.T) {
	t.Parallel()
	b := sampleBranch()
	newName := "Trụ sở chính (updated)"
	b.Name = newName
	uc := newTestOrgUC(&mockOrgRepo{branch: b})

	callerID := uuid.New()
	resp, err := uc.UpdateBranch(context.Background(), b.ID,
		usecase.UpdateBranchHRMRequest{Name: &newName}, &callerID, "127.0.0.1",
		[]string{"SUPER_ADMIN"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != newName {
		t.Errorf("got name %q, want %q", resp.Name, newName)
	}
}

func TestOrgUseCase_UpdateBranch_NotFound(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{branchErr: domain.ErrBranchNotFound})

	name := "x"
	callerID := uuid.New()
	_, err := uc.UpdateBranch(context.Background(), uuid.New(),
		usecase.UpdateBranchHRMRequest{Name: &name}, &callerID, "127.0.0.1",
		[]string{"SUPER_ADMIN"}, nil)
	if err != domain.ErrBranchNotFound {
		t.Errorf("expected ErrBranchNotFound, got %v", err)
	}
}

// ─── Department tests ─────────────────────────────────────────────────────────

func TestOrgUseCase_GetDept_Success(t *testing.T) {
	t.Parallel()
	d := sampleDept()
	uc := newTestOrgUC(&mockOrgRepo{dept: d})

	resp, err := uc.GetDept(context.Background(), d.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Code != "AUDIT" {
		t.Errorf("got code %q, want AUDIT", resp.Code)
	}
	if resp.DeptType != "CORE" {
		t.Errorf("got dept_type %q, want CORE", resp.DeptType)
	}
}

func TestOrgUseCase_GetDept_NotFound(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{deptErr: domain.ErrDeptNotFound})

	_, err := uc.GetDept(context.Background(), uuid.New())
	if err != domain.ErrDeptNotFound {
		t.Errorf("expected ErrDeptNotFound, got %v", err)
	}
}

func TestOrgUseCase_UpdateDept_InvalidDeptType_Passthrough(t *testing.T) {
	t.Parallel()
	d := sampleDept()
	invalid := "INVALID"
	uc := newTestOrgUC(&mockOrgRepo{dept: d})

	callerID := uuid.New()
	// UseCase does not validate dept_type itself — CHECK constraint enforced at DB layer.
	// We test that the usecase passes the value through without error on its own.
	_, err := uc.UpdateDept(context.Background(), d.ID,
		usecase.UpdateDeptHRMRequest{DeptType: &invalid}, &callerID, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── BranchDepartment tests ───────────────────────────────────────────────────

func TestOrgUseCase_LinkBranchDept_Success(t *testing.T) {
	t.Parallel()
	bd := sampleBD()
	uc := newTestOrgUC(&mockOrgRepo{bd: bd})

	callerID := uuid.New()
	resp, err := uc.LinkBranchDept(context.Background(),
		usecase.LinkBranchDeptRequest{BranchID: bd.BranchID, DepartmentID: bd.DepartmentID},
		&callerID, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.BranchID != bd.BranchID {
		t.Errorf("got branch_id %v, want %v", resp.BranchID, bd.BranchID)
	}
}

func TestOrgUseCase_LinkBranchDept_Duplicate(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{bdErr: domain.ErrDuplicateBranchDept})

	callerID := uuid.New()
	_, err := uc.LinkBranchDept(context.Background(),
		usecase.LinkBranchDeptRequest{BranchID: uuid.New(), DepartmentID: uuid.New()},
		&callerID, "127.0.0.1")
	if err != domain.ErrDuplicateBranchDept {
		t.Errorf("expected ErrDuplicateBranchDept, got %v", err)
	}
}

func TestOrgUseCase_UnlinkBranchDept_Success(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{softDeleteErr: nil})

	callerID := uuid.New()
	err := uc.UnlinkBranchDept(context.Background(), uuid.New(), uuid.New(), &callerID, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgUseCase_UnlinkBranchDept_NotFound(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{softDeleteErr: domain.ErrBranchDeptNotFound})

	callerID := uuid.New()
	err := uc.UnlinkBranchDept(context.Background(), uuid.New(), uuid.New(), &callerID, "127.0.0.1")
	if err != domain.ErrBranchDeptNotFound {
		t.Errorf("expected ErrBranchDeptNotFound, got %v", err)
	}
}

// ─── OrgChart tests ───────────────────────────────────────────────────────────

func TestOrgUseCase_GetOrgChart_ReturnsBranchesWithDepts(t *testing.T) {
	t.Parallel()
	branches := []*domain.OrgChartBranch{
		{
			ID:           uuid.New(),
			Code:         "HO",
			Name:         "Trụ sở chính",
			IsHeadOffice: true,
			Departments: []domain.OrgChartDept{
				{ID: uuid.New(), Code: "AUDIT", Name: "Kiểm toán", DeptType: "CORE"},
			},
		},
	}
	uc := newTestOrgUC(&mockOrgRepo{chartData: branches})

	resp, err := uc.GetOrgChart(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(resp.Branches))
	}
	if len(resp.Branches[0].Departments) != 1 {
		t.Errorf("expected 1 dept, got %d", len(resp.Branches[0].Departments))
	}
}

// ─── HoB scope tests ──────────────────────────────────────────────────────────

func TestOrgUseCase_UpdateBranch_HoB_OwnBranch_NonCritical_OK(t *testing.T) {
	t.Parallel()
	b := sampleBranch()
	uc := newTestOrgUC(&mockOrgRepo{branch: b})

	addr := "123 Street"
	callerID := uuid.New()
	callerBranchID := b.ID
	_, err := uc.UpdateBranch(context.Background(), b.ID,
		usecase.UpdateBranchHRMRequest{Address: &addr}, &callerID, "127.0.0.1",
		[]string{"HEAD_OF_BRANCH"}, &callerBranchID)
	if err != nil {
		t.Fatalf("HoB own-branch non-critical update should succeed, got: %v", err)
	}
}

func TestOrgUseCase_UpdateBranch_HoB_OtherBranch_Forbidden(t *testing.T) {
	t.Parallel()
	b := sampleBranch()
	uc := newTestOrgUC(&mockOrgRepo{branch: b})

	addr := "123 Street"
	callerID := uuid.New()
	otherBranchID := uuid.New()
	_, err := uc.UpdateBranch(context.Background(), b.ID,
		usecase.UpdateBranchHRMRequest{Address: &addr}, &callerID, "127.0.0.1",
		[]string{"HEAD_OF_BRANCH"}, &otherBranchID)
	if err != domain.ErrInsufficientPermission {
		t.Errorf("expected ErrInsufficientPermission, got %v", err)
	}
}

func TestOrgUseCase_UpdateBranch_HoB_CriticalField_Forbidden(t *testing.T) {
	t.Parallel()
	b := sampleBranch()
	uc := newTestOrgUC(&mockOrgRepo{branch: b})

	name := "New Name"
	callerID := uuid.New()
	callerBranchID := b.ID
	_, err := uc.UpdateBranch(context.Background(), b.ID,
		usecase.UpdateBranchHRMRequest{Name: &name}, &callerID, "127.0.0.1",
		[]string{"HEAD_OF_BRANCH"}, &callerBranchID)
	if err != domain.ErrInsufficientPermission {
		t.Errorf("expected ErrInsufficientPermission for critical field, got %v", err)
	}
}

// ─── AssignBranchHead / DeactivateBranch / AssignDeptHead / DeactivateDept ───

func TestOrgUseCase_AssignBranchHead_Success(t *testing.T) {
	t.Parallel()
	b := sampleBranch()
	uc := newTestOrgUC(&mockOrgRepo{branch: b})

	callerID := uuid.New()
	resp, err := uc.AssignBranchHead(context.Background(), b.ID,
		usecase.AssignBranchHeadRequest{UserID: uuid.New().String()}, &callerID, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != b.ID {
		t.Errorf("got id %v, want %v", resp.ID, b.ID)
	}
}

func TestOrgUseCase_DeactivateBranch_Success(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{branchErr: nil})

	callerID := uuid.New()
	err := uc.DeactivateBranch(context.Background(), uuid.New(), &callerID, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgUseCase_DeactivateBranch_NotFound(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{branchErr: domain.ErrBranchNotFound})

	callerID := uuid.New()
	err := uc.DeactivateBranch(context.Background(), uuid.New(), &callerID, "127.0.0.1")
	if err != domain.ErrBranchNotFound {
		t.Errorf("expected ErrBranchNotFound, got %v", err)
	}
}

func TestOrgUseCase_AssignDeptHead_Success(t *testing.T) {
	t.Parallel()
	d := sampleDept()
	uc := newTestOrgUC(&mockOrgRepo{dept: d})

	callerID := uuid.New()
	resp, err := uc.AssignDeptHead(context.Background(), d.ID,
		usecase.AssignDeptHeadRequest{EmployeeID: uuid.New().String()}, &callerID, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != d.ID {
		t.Errorf("got id %v, want %v", resp.ID, d.ID)
	}
}

func TestOrgUseCase_DeactivateDept_Success(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{deptErr: nil})

	callerID := uuid.New()
	err := uc.DeactivateDept(context.Background(), uuid.New(), &callerID, "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOrgUseCase_DeactivateDept_NotFound(t *testing.T) {
	t.Parallel()
	uc := newTestOrgUC(&mockOrgRepo{deptErr: domain.ErrDeptNotFound})

	callerID := uuid.New()
	err := uc.DeactivateDept(context.Background(), uuid.New(), &callerID, "127.0.0.1")
	if err != domain.ErrDeptNotFound {
		t.Errorf("expected ErrDeptNotFound, got %v", err)
	}
}
