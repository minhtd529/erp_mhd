package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// UserListRequest holds validated query parameters for GET /api/v1/users.
type UserListRequest struct {
	Page   int    `form:"page,default=1"  binding:"min=1"`
	Size   int    `form:"size,default=20" binding:"min=1,max=100"`
	Status string `form:"status"`
	Q      string `form:"q"`
}

// UserSummaryResponse is a single user row in a list response.
type UserSummaryResponse struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	FullName         string     `json:"full_name"`
	Status           string     `json:"status"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	BranchID         *uuid.UUID `json:"branch_id"`
	DepartmentID     *uuid.UUID `json:"department_id"`
}

// UserUpdateRequest is the body for PUT /api/v1/users/:id.
type UserUpdateRequest struct {
	FullName     string     `json:"full_name"     binding:"required"`
	Status       string     `json:"status"        binding:"required,oneof=active inactive locked"`
	BranchID     *uuid.UUID `json:"branch_id"`
	DepartmentID *uuid.UUID `json:"department_id"`
}

// MeResponse is the current user profile returned by GET /api/v1/me.
type MeResponse struct {
	ID               uuid.UUID  `json:"id"`
	Email            string     `json:"email"`
	FullName         string     `json:"full_name"`
	Status           string     `json:"status"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	TwoFactorMethod  string     `json:"two_factor_method,omitempty"`
	BranchID         *uuid.UUID `json:"branch_id"`
	DepartmentID     *uuid.UUID `json:"department_id"`
	Roles            []string   `json:"roles"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
}

// ─── ListUsersUseCase ────────────────────────────────────────────────────────

// ListUsersUseCase returns a paginated list of users.
type ListUsersUseCase struct {
	users domain.UserRepository
	roles domain.RoleRepository
}

func NewListUsersUseCase(users domain.UserRepository, roles domain.RoleRepository) *ListUsersUseCase {
	return &ListUsersUseCase{users: users, roles: roles}
}

// GetByID retrieves a single user by ID.
func (uc *ListUsersUseCase) GetByID(ctx context.Context, id uuid.UUID) (*UserSummaryResponse, error) {
	u, err := uc.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	roles, _ := uc.roles.GetUserRoles(ctx, u.ID)
	resp := &UserSummaryResponse{
		ID:               u.ID,
		Email:            u.Email,
		FullName:         u.FullName,
		Status:           u.Status,
		TwoFactorEnabled: u.TwoFactorEnabled,
		BranchID:         u.BranchID,
		DepartmentID:     u.DepartmentID,
	}
	_ = roles // roles available if needed by caller
	return resp, nil
}

func (uc *ListUsersUseCase) Execute(ctx context.Context, req UserListRequest) (PaginatedResult[UserSummaryResponse], error) {
	list, total, err := uc.users.ListUsers(ctx, domain.ListUsersFilter{
		Page:   req.Page,
		Size:   req.Size,
		Status: req.Status,
		Q:      req.Q,
	})
	if err != nil {
		return PaginatedResult[UserSummaryResponse]{}, fmt.Errorf("list users: %w", err)
	}

	items := make([]UserSummaryResponse, len(list))
	for i, u := range list {
		items[i] = UserSummaryResponse{
			ID:               u.ID,
			Email:            u.Email,
			FullName:         u.FullName,
			Status:           u.Status,
			TwoFactorEnabled: u.TwoFactorEnabled,
			BranchID:         u.BranchID,
			DepartmentID:     u.DepartmentID,
		}
	}

	return pagination.NewOffsetResult(items, total, req.Page, req.Size), nil
}

// ─── UpdateUserUseCase ───────────────────────────────────────────────────────

// UpdateUserUseCase updates mutable fields on a user.
type UpdateUserUseCase struct {
	users    domain.UserRepository
	auditLog *audit.Logger
}

func NewUpdateUserUseCase(users domain.UserRepository, auditLog *audit.Logger) *UpdateUserUseCase {
	return &UpdateUserUseCase{users: users, auditLog: auditLog}
}

func (uc *UpdateUserUseCase) Execute(ctx context.Context, targetID uuid.UUID, req UserUpdateRequest, callerID *uuid.UUID, ip string) error {
	if err := uc.users.UpdateUser(ctx, domain.UpdateUserParams{
		ID:           targetID,
		FullName:     req.FullName,
		BranchID:     req.BranchID,
		DepartmentID: req.DepartmentID,
		Status:       req.Status,
		UpdatedBy:    callerID,
	}); err != nil {
		return err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     callerID,
		Module:     "global",
		Resource:   "users",
		ResourceID: &targetID,
		Action:     "UPDATE",
		IPAddress:  ip,
	})
	return nil
}

// ─── DeleteUserUseCase ───────────────────────────────────────────────────────

// DeleteUserUseCase soft-deletes a user.
type DeleteUserUseCase struct {
	users    domain.UserRepository
	auditLog *audit.Logger
}

func NewDeleteUserUseCase(users domain.UserRepository, auditLog *audit.Logger) *DeleteUserUseCase {
	return &DeleteUserUseCase{users: users, auditLog: auditLog}
}

func (uc *DeleteUserUseCase) Execute(ctx context.Context, targetID uuid.UUID, callerID *uuid.UUID, ip string) error {
	if err := uc.users.SoftDeleteUser(ctx, targetID, callerID); err != nil {
		return err
	}

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     callerID,
		Module:     "global",
		Resource:   "users",
		ResourceID: &targetID,
		Action:     "DELETE",
		IPAddress:  ip,
	})
	return nil
}
