package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	pkgauth "github.com/mdh/erp-audit/api/pkg/auth"
)

// CreateUserUseCase creates a new user account and assigns an initial role.
type CreateUserUseCase struct {
	users    domain.UserRepository
	roles    domain.RoleRepository
	auditLog *audit.Logger
}

// NewCreateUserUseCase constructs a CreateUserUseCase.
func NewCreateUserUseCase(
	users domain.UserRepository,
	roles domain.RoleRepository,
	auditLog *audit.Logger,
) *CreateUserUseCase {
	return &CreateUserUseCase{users: users, roles: roles, auditLog: auditLog}
}

// Execute validates, creates, and returns the new user.
func (uc *CreateUserUseCase) Execute(
	ctx context.Context,
	req UserCreateRequest,
	createdBy *uuid.UUID,
	ipAddress string,
) (*UserDetailResponse, error) {
	// 1. Hash password
	hashed, err := pkgauth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("create user: hash password: %w", err)
	}

	// 2. Insert user (repo returns ErrUserAlreadyExists on duplicate email)
	userID, err := uc.users.CreateUser(ctx, domain.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashed,
		FullName:       req.FullName,
		BranchID:       req.BranchID,
		DepartmentID:   req.DepartmentID,
		CreatedBy:      createdBy,
	})
	if err != nil {
		return nil, err // domain sentinel already set by repo
	}

	// 3. Assign initial role
	role, err := uc.roles.FindByCode(ctx, req.RoleCode)
	if err != nil {
		return nil, err
	}
	if err := uc.roles.AssignToUser(ctx, userID, role.ID); err != nil {
		return nil, fmt.Errorf("create user: assign role: %w", err)
	}

	// 4. Emit audit log (best-effort)
	if uc.auditLog != nil {
		_ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     createdBy,
			Module:     "global",
			Resource:   "users",
			ResourceID: &userID,
			Action:     "CREATE",
			IPAddress:  ipAddress,
		})
	}

	return &UserDetailResponse{
		ID:           userID,
		Email:        req.Email,
		FullName:     req.FullName,
		Status:       "active",
		Roles:        []string{role.Code},
		BranchID:     req.BranchID,
		DepartmentID: req.DepartmentID,
	}, nil
}
