package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/global/auth/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// AssignRoleUseCase assigns a role to an existing user.
type AssignRoleUseCase struct {
	users    domain.UserRepository
	roles    domain.RoleRepository
	auditLog *audit.Logger
}

// NewAssignRoleUseCase constructs an AssignRoleUseCase.
func NewAssignRoleUseCase(
	users domain.UserRepository,
	roles domain.RoleRepository,
	auditLog *audit.Logger,
) *AssignRoleUseCase {
	return &AssignRoleUseCase{users: users, roles: roles, auditLog: auditLog}
}

// Execute assigns roleCode to the user identified by userID.
func (uc *AssignRoleUseCase) Execute(
	ctx context.Context,
	userID uuid.UUID,
	roleCode string,
	assignedBy *uuid.UUID,
	ipAddress string,
) error {
	// Verify user exists
	user, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify role exists
	role, err := uc.roles.FindByCode(ctx, roleCode)
	if err != nil {
		return err
	}

	// Upsert assignment
	if err := uc.roles.AssignToUser(ctx, user.ID, role.ID); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}

	// Emit audit log (best-effort)
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID:     assignedBy,
			Module:     "global",
			Resource:   "user_roles",
			ResourceID: &userID,
			Action:     "ASSIGN_ROLE",
			IPAddress:  ipAddress,
			NewValue:   map[string]string{"role_code": roleCode},
		})
	}

	return nil
}
