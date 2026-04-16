package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/internal/hrm/usecase"
)

// ── mock repo ─────────────────────────────────────────────────────────────────

type mockEmployeeRepo struct {
	created   *domain.Employee
	found     *domain.Employee
	updated   *domain.Employee
	createErr error
	findErr   error
	updateErr error
	deleteErr error
	listItems []*domain.Employee
	listTotal int64
}

func (m *mockEmployeeRepo) Create(_ context.Context, _ domain.CreateEmployeeParams) (*domain.Employee, error) {
	return m.created, m.createErr
}
func (m *mockEmployeeRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Employee, error) {
	return m.found, m.findErr
}
func (m *mockEmployeeRepo) Update(_ context.Context, _ domain.UpdateEmployeeParams) (*domain.Employee, error) {
	return m.updated, m.updateErr
}
func (m *mockEmployeeRepo) UpdateBankDetails(_ context.Context, _ domain.UpdateBankDetailsParams) (*domain.Employee, error) {
	return m.updated, m.updateErr
}
func (m *mockEmployeeRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return m.deleteErr
}
func (m *mockEmployeeRepo) List(_ context.Context, _ domain.ListEmployeesFilter) ([]*domain.Employee, int64, error) {
	return m.listItems, m.listTotal, m.findErr
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestEmployeeUseCase_Create(t *testing.T) {
	t.Parallel()

	empID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockEmployeeRepo
		req     usecase.EmployeeCreateRequest
		wantErr error
	}{
		{
			name: "success",
			repo: &mockEmployeeRepo{
				created: &domain.Employee{
					ID:       empID,
					FullName: "Alice Smith",
					Email:    "alice@firm.com",
					Grade:    domain.GradeJunior,
					Status:   domain.StatusActive,
				},
			},
			req: usecase.EmployeeCreateRequest{
				FullName: "Alice Smith",
				Email:    "alice@firm.com",
				Grade:    domain.GradeJunior,
			},
		},
		{
			name:    "duplicate email → DUPLICATE_EMAIL",
			repo:    &mockEmployeeRepo{createErr: domain.ErrDuplicateEmail},
			req:     usecase.EmployeeCreateRequest{FullName: "Bob", Email: "dup@firm.com", Grade: domain.GradeJunior},
			wantErr: domain.ErrDuplicateEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewEmployeeUseCase(tt.repo, nil, "")
			resp, err := uc.Create(context.Background(), tt.req, nil, "")
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want err %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Email != tt.req.Email {
				t.Errorf("want email %q, got %q", tt.req.Email, resp.Email)
			}
		})
	}
}

func TestEmployeeUseCase_GetByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name    string
		repo    *mockEmployeeRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockEmployeeRepo{
				found: &domain.Employee{ID: id, FullName: "Alice", Email: "a@b.com", Grade: domain.GradeSenior, Status: domain.StatusActive},
			},
		},
		{
			name:    "not found → EMPLOYEE_NOT_FOUND",
			repo:    &mockEmployeeRepo{findErr: domain.ErrEmployeeNotFound},
			wantErr: domain.ErrEmployeeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewEmployeeUseCase(tt.repo, nil, "")
			_, err := uc.GetByID(context.Background(), id)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want err %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEmployeeUseCase_Update(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name    string
		repo    *mockEmployeeRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockEmployeeRepo{
				updated: &domain.Employee{ID: id, FullName: "Updated", Email: "a@b.com", Grade: domain.GradeManager, Status: domain.StatusActive},
			},
		},
		{
			name:    "not found → EMPLOYEE_NOT_FOUND",
			repo:    &mockEmployeeRepo{updateErr: domain.ErrEmployeeNotFound},
			wantErr: domain.ErrEmployeeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewEmployeeUseCase(tt.repo, nil, "")
			_, err := uc.Update(context.Background(), id, usecase.EmployeeUpdateRequest{FullName: "Updated", Grade: domain.GradeManager}, nil, "")
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want err %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEmployeeUseCase_Delete(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name    string
		repo    *mockEmployeeRepo
		wantErr error
	}{
		{name: "success", repo: &mockEmployeeRepo{}},
		{
			name:    "not found → EMPLOYEE_NOT_FOUND",
			repo:    &mockEmployeeRepo{deleteErr: domain.ErrEmployeeNotFound},
			wantErr: domain.ErrEmployeeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewEmployeeUseCase(tt.repo, nil, "")
			err := uc.Delete(context.Background(), id, nil, "")
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want err %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEmployeeUseCase_UpdateBankDetails(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	// A valid 64-char hex key (32 zero bytes) — safe for test use only.
	const testKey = "0000000000000000000000000000000000000000000000000000000000000000"

	name := "Nguyen Van A"
	tests := []struct {
		name    string
		repo    *mockEmployeeRepo
		wantErr bool
	}{
		{
			name: "success",
			repo: &mockEmployeeRepo{
				updated: &domain.Employee{ID: id, FullName: "Alice", Email: "a@b.com", Grade: domain.GradeSenior, Status: domain.StatusActive},
			},
		},
		{
			name:    "employee not found",
			repo:    &mockEmployeeRepo{updateErr: domain.ErrEmployeeNotFound},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewEmployeeUseCase(tt.repo, nil, testKey)
			err := uc.UpdateBankDetails(context.Background(), id, usecase.EmployeeBankDetailsRequest{
				BankAccountNumber: "1234567890",
				BankAccountName:   &name,
			}, nil, "")
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestEmployeeUseCase_List(t *testing.T) {
	t.Parallel()

	items := []*domain.Employee{
		{ID: uuid.New(), FullName: "Alice", Email: "alice@firm.com", Grade: domain.GradeJunior, Status: domain.StatusActive},
		{ID: uuid.New(), FullName: "Bob", Email: "bob@firm.com", Grade: domain.GradeSenior, Status: domain.StatusActive},
	}

	uc := usecase.NewEmployeeUseCase(&mockEmployeeRepo{listItems: items, listTotal: 2}, nil, "")
	result, err := uc.List(context.Background(), usecase.EmployeeListRequest{Page: 1, Size: 20})
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
