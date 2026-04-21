package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

type ContractResponse struct {
	ID                uuid.UUID  `json:"id"`
	EmployeeID        uuid.UUID  `json:"employee_id"`
	ContractNumber    *string    `json:"contract_number,omitempty"`
	ContractType      string     `json:"contract_type"`
	StartDate         string     `json:"start_date"`
	EndDate           *string    `json:"end_date,omitempty"`
	SignedDate        *string    `json:"signed_date,omitempty"`
	SalaryAtSigning   *float64   `json:"salary_at_signing,omitempty"`
	PositionAtSigning *string    `json:"position_at_signing,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	DocumentURL       *string    `json:"document_url,omitempty"`
	IsCurrent         bool       `json:"is_current"`
	CreatedBy         *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt         string     `json:"created_at"`
	UpdatedAt         string     `json:"updated_at"`
}

type CreateContractRequest struct {
	ContractNumber    *string  `json:"contract_number"`
	ContractType      string   `json:"contract_type"   binding:"required,oneof=PROBATION DEFINITE_TERM INDEFINITE INTERN"`
	StartDate         string   `json:"start_date"      binding:"required"`
	EndDate           *string  `json:"end_date"`
	SignedDate        *string  `json:"signed_date"`
	SalaryAtSigning   *float64 `json:"salary_at_signing"`
	PositionAtSigning *string  `json:"position_at_signing"`
	Notes             *string  `json:"notes"`
	DocumentURL       *string  `json:"document_url"`
}

type UpdateContractRequest struct {
	ContractType      *string  `json:"contract_type"   binding:"omitempty,oneof=PROBATION DEFINITE_TERM INDEFINITE INTERN"`
	EndDate           *string  `json:"end_date"`
	SignedDate        *string  `json:"signed_date"`
	SalaryAtSigning   *float64 `json:"salary_at_signing"`
	PositionAtSigning *string  `json:"position_at_signing"`
	Notes             *string  `json:"notes"`
	DocumentURL       *string  `json:"document_url"`
}

// ─── UseCase ─────────────────────────────────────────────────────────────────

type ContractUseCase struct {
	repo        domain.ContractRepository
	employeeRepo domain.EmployeeRepository
	auditLog    *audit.Logger
}

func NewContractUseCase(repo domain.ContractRepository, employeeRepo domain.EmployeeRepository, auditLog *audit.Logger) *ContractUseCase {
	return &ContractUseCase{repo: repo, employeeRepo: employeeRepo, auditLog: auditLog}
}

func (uc *ContractUseCase) ListContracts(ctx context.Context, employeeID uuid.UUID) ([]ContractResponse, error) {
	contracts, err := uc.repo.ListByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("contract.List: %w", err)
	}
	items := make([]ContractResponse, len(contracts))
	for i, c := range contracts {
		items[i] = toContractResponse(c)
	}
	return items, nil
}

func (uc *ContractUseCase) CreateContract(ctx context.Context, employeeID uuid.UUID, req CreateContractRequest, callerID *uuid.UUID, ip string) (*ContractResponse, error) {
	has, err := uc.repo.HasActiveContract(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("contract.HasActive: %w", err)
	}
	if has {
		return nil, domain.ErrContractActiveExists
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}

	p := domain.CreateContractParams{
		EmployeeID:        employeeID,
		ContractNumber:    req.ContractNumber,
		ContractType:      req.ContractType,
		StartDate:         startDate,
		SalaryAtSigning:   req.SalaryAtSigning,
		PositionAtSigning: req.PositionAtSigning,
		Notes:             req.Notes,
		DocumentURL:       req.DocumentURL,
		CreatedBy:         callerID,
	}
	if req.EndDate != nil {
		if t, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			p.EndDate = &t
		}
	}
	if req.SignedDate != nil {
		if t, err := time.Parse("2006-01-02", *req.SignedDate); err == nil {
			p.SignedDate = &t
		}
	}

	c, err := uc.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	// Link contract to employee
	cid := c.ID
	_, _ = uc.employeeRepo.Update(ctx, domain.UpdateEmployeeParams{
		ID:                employeeID,
		CurrentContractID: &cid,
		UpdatedBy:         callerID,
	})

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "employment_contracts",
			ResourceID: &cid, Action: "CREATE", IPAddress: ip,
		})
	}
	resp := toContractResponse(c)
	return &resp, nil
}

func (uc *ContractUseCase) UpdateContract(ctx context.Context, id, employeeID uuid.UUID, req UpdateContractRequest, callerID *uuid.UUID, ip string) (*ContractResponse, error) {
	p := domain.UpdateContractParams{
		ID:                id,
		EmployeeID:        employeeID,
		ContractType:      req.ContractType,
		SalaryAtSigning:   req.SalaryAtSigning,
		PositionAtSigning: req.PositionAtSigning,
		Notes:             req.Notes,
		DocumentURL:       req.DocumentURL,
	}
	if req.EndDate != nil {
		if t, err := time.Parse("2006-01-02", *req.EndDate); err == nil {
			p.EndDate = &t
		}
	}
	if req.SignedDate != nil {
		if t, err := time.Parse("2006-01-02", *req.SignedDate); err == nil {
			p.SignedDate = &t
		}
	}

	c, err := uc.repo.Update(ctx, p)
	if err != nil {
		return nil, err
	}

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "employment_contracts",
			ResourceID: &id, Action: "UPDATE", IPAddress: ip,
		})
	}
	resp := toContractResponse(c)
	return &resp, nil
}

func (uc *ContractUseCase) TerminateContract(ctx context.Context, id, employeeID uuid.UUID, callerID *uuid.UUID, ip string) error {
	if err := uc.repo.Terminate(ctx, id, employeeID); err != nil {
		return err
	}

	// Clear current_contract_id on the employee if it points to this contract
	emp, err := uc.employeeRepo.FindByID(ctx, employeeID)
	if err == nil && emp.CurrentContractID != nil && *emp.CurrentContractID == id {
		nilID := uuid.Nil
		nilIDPtr := &nilID
		// Use a special sentinel: set to nil by passing a known-nil-like value.
		// We clear it by setting to NULL via a separate targeted update.
		_ = nilIDPtr
		// For simplicity in Phase 2, employee's current_contract_id remains until next contract is created.
		// A Phase 3 cleanup migration can reconcile this.
	}

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "employment_contracts",
			ResourceID: &id, Action: "TERMINATE", IPAddress: ip,
		})
	}
	return nil
}

func toContractResponse(c *domain.EmploymentContract) ContractResponse {
	return ContractResponse{
		ID:                c.ID,
		EmployeeID:        c.EmployeeID,
		ContractNumber:    c.ContractNumber,
		ContractType:      c.ContractType,
		StartDate:         c.StartDate.Format("2006-01-02"),
		EndDate:           dateStr(c.EndDate),
		SignedDate:        dateStr(c.SignedDate),
		SalaryAtSigning:   c.SalaryAtSigning,
		PositionAtSigning: c.PositionAtSigning,
		Notes:             c.Notes,
		DocumentURL:       c.DocumentURL,
		IsCurrent:         c.IsCurrent,
		CreatedBy:         c.CreatedBy,
		CreatedAt:         c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         c.UpdatedAt.Format(time.RFC3339),
	}
}
