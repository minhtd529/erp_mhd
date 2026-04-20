package domain

import (
	"context"

	"github.com/google/uuid"
)

type BranchRepository interface {
	Create(ctx context.Context, p CreateBranchParams) (*Branch, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Branch, error)
	Update(ctx context.Context, p UpdateBranchParams) (*Branch, error)
	List(ctx context.Context, f ListBranchesFilter) ([]*Branch, int64, error)
}

type DepartmentRepository interface {
	Create(ctx context.Context, p CreateDepartmentParams) (*Department, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Department, error)
	Update(ctx context.Context, p UpdateDepartmentParams) (*Department, error)
	List(ctx context.Context, f ListDepartmentsFilter) ([]*Department, int64, error)
}
