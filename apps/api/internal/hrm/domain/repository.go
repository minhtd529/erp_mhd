package domain

import (
	"context"

	"github.com/google/uuid"
)

// EmployeeRepository defines the data-access contract for Employee.
type EmployeeRepository interface {
	Create(ctx context.Context, p CreateEmployeeParams) (*Employee, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Employee, error)
	Update(ctx context.Context, p UpdateEmployeeParams) (*Employee, error)
	UpdateBankDetails(ctx context.Context, p UpdateBankDetailsParams) (*Employee, error)
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	List(ctx context.Context, f ListEmployeesFilter) ([]*Employee, int64, error)
}
