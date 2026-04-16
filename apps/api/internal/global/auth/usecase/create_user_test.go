package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
)

func TestCreateUserUseCase(t *testing.T) {
	t.Parallel()

	newID := uuid.New()
	adminRole := &domain.Role{ID: uuid.New(), Code: "AUDIT_STAFF", Name: "Audit Staff"}

	tests := []struct {
		name      string
		req       usecase.UserCreateRequest
		userRepo  *mockUserRepo
		roleRepo  *mockRoleRepo
		wantErr   error
		wantEmail string
	}{
		{
			name: "success",
			req: usecase.UserCreateRequest{
				Email:    "bob@example.com",
				Password: "Password123!",
				FullName: "Bob",
				RoleCode: "AUDIT_STAFF",
			},
			userRepo:  &mockUserRepo{createID: newID},
			roleRepo:  &mockRoleRepo{role: adminRole},
			wantEmail: "bob@example.com",
		},
		{
			name: "duplicate email → USER_ALREADY_EXISTS",
			req: usecase.UserCreateRequest{
				Email:    "existing@example.com",
				Password: "Password123!",
				FullName: "Dup",
				RoleCode: "AUDIT_STAFF",
			},
			userRepo: &mockUserRepo{createErr: domain.ErrUserAlreadyExists},
			roleRepo: &mockRoleRepo{role: adminRole},
			wantErr:  domain.ErrUserAlreadyExists,
		},
		{
			name: "unknown role → ROLE_NOT_FOUND",
			req: usecase.UserCreateRequest{
				Email:    "carol@example.com",
				Password: "Password123!",
				FullName: "Carol",
				RoleCode: "UNKNOWN",
			},
			userRepo: &mockUserRepo{createID: uuid.New()},
			roleRepo: &mockRoleRepo{findErr: domain.ErrRoleNotFound},
			wantErr:  domain.ErrRoleNotFound,
		},
		{
			name: "role assign fails → internal error",
			req: usecase.UserCreateRequest{
				Email:    "dave@example.com",
				Password: "Password123!",
				FullName: "Dave",
				RoleCode: "AUDIT_STAFF",
			},
			userRepo: &mockUserRepo{createID: uuid.New()},
			roleRepo: &mockRoleRepo{role: adminRole, assignErr: errors.New("db error")},
			wantErr:  errors.New("create user: assign role"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewCreateUserUseCase(tt.userRepo, tt.roleRepo, nil)
			resp, err := uc.Execute(context.Background(), tt.req, nil, "127.0.0.1")

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				// Check sentinel errors with errors.Is
				if errors.Is(tt.wantErr, domain.ErrUserAlreadyExists) ||
					errors.Is(tt.wantErr, domain.ErrRoleNotFound) {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("want %v, got %v", tt.wantErr, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Email != tt.wantEmail {
				t.Errorf("want email %q, got %q", tt.wantEmail, resp.Email)
			}
			if resp.Status != "active" {
				t.Errorf("want status active, got %q", resp.Status)
			}
		})
	}
}
