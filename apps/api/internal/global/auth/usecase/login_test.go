package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/internal/global/auth/usecase"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// activeUser returns a UserForAuth with a valid bcrypt-hashed password for "secret123".
func activeUser() *domain.UserForAuth {
	hashed, _ := pkgauth.HashPassword("secret123")
	return &domain.UserForAuth{
		ID:             uuid.New(),
		Email:          "alice@example.com",
		HashedPassword: hashed,
		FullName:       "Alice",
		Status:         "active",
		Roles:          []string{"AUDIT_STAFF"},
		Permissions:    []string{"crm:client:read"},
	}
}

func TestLoginUseCase(t *testing.T) {
	t.Parallel()

	validJWT := &mockJWTSvc{
		accessToken: "access.token.here",
		rawRefresh:  uuid.New().String(),
		expiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}

	tests := []struct {
		name      string
		req       usecase.LoginRequest
		userRepo  *mockUserRepo
		tokenRepo *mockTokenRepo
		jwtSvc    *mockJWTSvc
		wantErr   error
		wantToken bool
	}{
		{
			name: "success",
			req:  usecase.LoginRequest{Email: "alice@example.com", Password: "secret123"},
			userRepo: &mockUserRepo{user: activeUser()},
			tokenRepo: &mockTokenRepo{},
			jwtSvc:    validJWT,
			wantToken: true,
		},
		{
			name:      "user not found → INVALID_CREDENTIALS",
			req:       usecase.LoginRequest{Email: "nobody@example.com", Password: "secret123"},
			userRepo:  &mockUserRepo{findErr: domain.ErrUserNotFound},
			tokenRepo: &mockTokenRepo{},
			jwtSvc:    validJWT,
			wantErr:   domain.ErrInvalidCredentials,
		},
		{
			name: "wrong password → INVALID_CREDENTIALS",
			req:  usecase.LoginRequest{Email: "alice@example.com", Password: "wrongpassword"},
			userRepo: &mockUserRepo{user: activeUser()},
			tokenRepo: &mockTokenRepo{},
			jwtSvc:    validJWT,
			wantErr:   domain.ErrInvalidCredentials,
		},
		{
			name: "locked account → USER_LOCKED",
			req:  usecase.LoginRequest{Email: "alice@example.com", Password: "secret123"},
			userRepo: func() *mockUserRepo {
				u := activeUser()
				u.Status = "locked"
				return &mockUserRepo{user: u}
			}(),
			tokenRepo: &mockTokenRepo{},
			jwtSvc:    validJWT,
			wantErr:   domain.ErrUserLocked,
		},
		{
			name: "inactive account → USER_INACTIVE",
			req:  usecase.LoginRequest{Email: "alice@example.com", Password: "secret123"},
			userRepo: func() *mockUserRepo {
				u := activeUser()
				u.Status = "inactive"
				return &mockUserRepo{user: u}
			}(),
			tokenRepo: &mockTokenRepo{},
			jwtSvc:    validJWT,
			wantErr:   domain.ErrUserInactive,
		},
		{
			name:      "JWT issuance failure → internal error",
			req:       usecase.LoginRequest{Email: "alice@example.com", Password: "secret123"},
			userRepo:  &mockUserRepo{user: activeUser()},
			tokenRepo: &mockTokenRepo{},
			jwtSvc:    &mockJWTSvc{issueErr: errors.New("signing key missing")},
			wantErr:   errors.New("login: issue access token"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewLoginUseCase(tt.userRepo, &mockRoleRepo{}, tt.tokenRepo, nil, tt.jwtSvc, nil)
			resp, err := uc.Execute(context.Background(), tt.req, "127.0.0.1")

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if tt.wantErr == domain.ErrInvalidCredentials ||
					tt.wantErr == domain.ErrUserLocked ||
					tt.wantErr == domain.ErrUserInactive {
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("want %v, got %v", tt.wantErr, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantToken && resp.AccessToken == "" {
				t.Error("expected non-empty access token")
			}
			if tt.wantToken && resp.RefreshToken == "" {
				t.Error("expected non-empty refresh token")
			}
			if tt.wantToken && resp.ExpiresIn != 900 {
				t.Errorf("expected expires_in=900, got %d", resp.ExpiresIn)
			}
		})
	}
}
