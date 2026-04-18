// Package usecase implements the HRM application layer.
package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/crypto"
)

// EmployeeUseCase bundles all employee CRUD operations.
type EmployeeUseCase struct {
	repo       domain.EmployeeRepository
	auditLog   *audit.Logger
	encryptKey string // hex-encoded 32-byte AES key for bank fields
}

// NewEmployeeUseCase constructs an EmployeeUseCase.
func NewEmployeeUseCase(repo domain.EmployeeRepository, auditLog *audit.Logger, encryptKey string) *EmployeeUseCase {
	return &EmployeeUseCase{repo: repo, auditLog: auditLog, encryptKey: encryptKey}
}

// Create creates a new employee with ACTIVE status.
func (uc *EmployeeUseCase) Create(ctx context.Context, req EmployeeCreateRequest, callerID *uuid.UUID, ip string) (*EmployeeResponse, error) {
	e, err := uc.repo.Create(ctx, domain.CreateEmployeeParams{
		FullName:                req.FullName,
		Email:                   req.Email,
		Phone:                   req.Phone,
		DateOfBirth:             req.DateOfBirth,
		Grade:                   req.Grade,
		Position:                req.Position,
		OfficeID:                req.OfficeID,
		ManagerID:               req.ManagerID,
		HourlyRate:              req.HourlyRate,
		EmploymentDate:          req.EmploymentDate,
		ContractEndDate:         req.ContractEndDate,
		IsSalesperson:           req.IsSalesperson,
		SalesCommissionEligible: req.SalesCommissionEligible,
		CreatedBy:               callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     callerID,
		Module:     "hrm",
		Resource:   "employees",
		ResourceID: &e.ID,
		Action:     "CREATE",
		IPAddress:  ip,
	})

	resp := toEmployeeResponse(e)
	return &resp, nil
}

// GetByID retrieves a single employee.
func (uc *EmployeeUseCase) GetByID(ctx context.Context, id uuid.UUID) (*EmployeeResponse, error) {
	e, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toEmployeeResponse(e)
	return &resp, nil
}

// Update mutates allowed fields on an existing employee.
func (uc *EmployeeUseCase) Update(ctx context.Context, id uuid.UUID, req EmployeeUpdateRequest, callerID *uuid.UUID, ip string) (*EmployeeResponse, error) {
	e, err := uc.repo.Update(ctx, domain.UpdateEmployeeParams{
		ID:                      id,
		FullName:                req.FullName,
		Phone:                   req.Phone,
		DateOfBirth:             req.DateOfBirth,
		Grade:                   req.Grade,
		Position:                req.Position,
		OfficeID:                req.OfficeID,
		ManagerID:               req.ManagerID,
		HourlyRate:              req.HourlyRate,
		EmploymentDate:          req.EmploymentDate,
		ContractEndDate:         req.ContractEndDate,
		IsSalesperson:           req.IsSalesperson,
		SalesCommissionEligible: req.SalesCommissionEligible,
		UpdatedBy:               callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     callerID,
		Module:     "hrm",
		Resource:   "employees",
		ResourceID: &id,
		Action:     "UPDATE",
		IPAddress:  ip,
	})

	resp := toEmployeeResponse(e)
	return &resp, nil
}

// Delete soft-deletes an employee (marks as RESIGNED).
func (uc *EmployeeUseCase) Delete(ctx context.Context, id uuid.UUID, callerID *uuid.UUID, ip string) error {
	if err := uc.repo.SoftDelete(ctx, id, callerID); err != nil {
		return err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     callerID,
		Module:     "hrm",
		Resource:   "employees",
		ResourceID: &id,
		Action:     "DELETE",
		IPAddress:  ip,
	})

	return nil
}

// UpdateBankDetails encrypts and stores bank account information for an employee.
// The plaintext account number is encrypted with AES-256-GCM before persistence.
func (uc *EmployeeUseCase) UpdateBankDetails(ctx context.Context, id uuid.UUID, req EmployeeBankDetailsRequest, callerID *uuid.UUID, ip string) error {
	enc, err := crypto.Encrypt(uc.encryptKey, req.BankAccountNumber)
	if err != nil {
		return fmt.Errorf("hrm.UpdateBankDetails encrypt: %w", err)
	}

	if _, err := uc.repo.UpdateBankDetails(ctx, domain.UpdateBankDetailsParams{
		ID:                   id,
		BankAccountNumberEnc: &enc,
		BankAccountName:      req.BankAccountName,
		UpdatedBy:            callerID,
	}); err != nil {
		return err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID:     callerID,
		Module:     "hrm",
		Resource:   "employees",
		ResourceID: &id,
		Action:     "UPDATE_BANK_DETAILS",
		IPAddress:  ip,
	})

	return nil
}

// List returns a paginated list of employees.
func (uc *EmployeeUseCase) List(ctx context.Context, req EmployeeListRequest) (PaginatedResult[EmployeeResponse], error) {
	employees, total, err := uc.repo.List(ctx, domain.ListEmployeesFilter{
		Page:          req.Page,
		Size:          req.Size,
		Status:        req.Status,
		OfficeID:      req.OfficeID,
		Q:             req.Q,
		IsSalesperson: req.IsSalesperson,
	})
	if err != nil {
		return PaginatedResult[EmployeeResponse]{}, fmt.Errorf("hrm.List: %w", err)
	}

	items := make([]EmployeeResponse, len(employees))
	for i, e := range employees {
		items[i] = toEmployeeResponse(e)
	}
	return newPaginatedResult(items, total, req.Page, req.Size), nil
}
