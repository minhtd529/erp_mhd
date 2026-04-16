package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
)

func TestListUsersUseCase(t *testing.T) {
	t.Parallel()

	uid := uuid.New()
	user := &domain.UserForAuth{
		ID:       uid,
		Email:    "alice@firm.com",
		FullName: "Alice",
		Status:   "active",
	}

	tests := []struct {
		name       string
		repo       *mockUserRepo
		wantTotal  int64
		wantLen    int
	}{
		{
			name:      "returns one user",
			repo:      &mockUserRepo{user: user},
			wantTotal: 1,
			wantLen:   1,
		},
		{
			name:      "empty list",
			repo:      &mockUserRepo{},
			wantTotal: 0,
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewListUsersUseCase(tt.repo, &mockRoleRepo{})
			result, err := uc.Execute(context.Background(), usecase.UserListRequest{Page: 1, Size: 20})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Total != tt.wantTotal {
				t.Errorf("want total %d, got %d", tt.wantTotal, result.Total)
			}
			if len(result.Data) != tt.wantLen {
				t.Errorf("want %d items, got %d", tt.wantLen, len(result.Data))
			}
		})
	}
}

func TestUpdateUserUseCase(t *testing.T) {
	t.Parallel()

	uid := uuid.New()

	tests := []struct {
		name    string
		repo    *mockUserRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockUserRepo{},
		},
		{
			name:    "not found → USER_NOT_FOUND",
			repo:    &mockUserRepo{findErr: domain.ErrUserNotFound},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewUpdateUserUseCase(tt.repo, nil)
			err := uc.Execute(context.Background(), uid, usecase.UserUpdateRequest{
				FullName: "Updated Name",
				Status:   "active",
			}, nil, "")
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

func TestDeleteUserUseCase(t *testing.T) {
	t.Parallel()

	uid := uuid.New()

	tests := []struct {
		name    string
		repo    *mockUserRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockUserRepo{},
		},
		{
			name:    "not found → USER_NOT_FOUND",
			repo:    &mockUserRepo{findErr: domain.ErrUserNotFound},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewDeleteUserUseCase(tt.repo, nil)
			err := uc.Execute(context.Background(), uid, nil, "")
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
