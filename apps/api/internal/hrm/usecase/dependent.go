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

type DependentResponse struct {
	ID                     uuid.UUID `json:"id"`
	EmployeeID             uuid.UUID `json:"employee_id"`
	FullName               string    `json:"full_name"`
	Relationship           string    `json:"relationship"`
	DateOfBirth            *string   `json:"date_of_birth,omitempty"`
	CccdOrBirthCert        *string   `json:"cccd_or_birth_cert,omitempty"`
	TaxDeductionRegistered bool      `json:"tax_deduction_registered"`
	TaxDeductionFrom       *string   `json:"tax_deduction_from,omitempty"`
	TaxDeductionTo         *string   `json:"tax_deduction_to,omitempty"`
	Notes                  *string   `json:"notes,omitempty"`
	CreatedAt              string    `json:"created_at"`
	UpdatedAt              string    `json:"updated_at"`
}

type CreateDependentRequest struct {
	FullName               string  `json:"full_name"     binding:"required,max=200"`
	Relationship           string  `json:"relationship"  binding:"required,oneof=SPOUSE CHILD PARENT SIBLING OTHER"`
	DateOfBirth            *string `json:"date_of_birth"`
	CccdOrBirthCert        *string `json:"cccd_or_birth_cert"`
	TaxDeductionRegistered bool    `json:"tax_deduction_registered"`
	TaxDeductionFrom       *string `json:"tax_deduction_from"`
	TaxDeductionTo         *string `json:"tax_deduction_to"`
	Notes                  *string `json:"notes"`
}

type UpdateDependentRequest struct {
	FullName               *string `json:"full_name"     binding:"omitempty,max=200"`
	Relationship           *string `json:"relationship"  binding:"omitempty,oneof=SPOUSE CHILD PARENT SIBLING OTHER"`
	DateOfBirth            *string `json:"date_of_birth"`
	CccdOrBirthCert        *string `json:"cccd_or_birth_cert"`
	TaxDeductionRegistered *bool   `json:"tax_deduction_registered"`
	TaxDeductionFrom       *string `json:"tax_deduction_from"`
	TaxDeductionTo         *string `json:"tax_deduction_to"`
	Notes                  *string `json:"notes"`
}

// ─── UseCase ─────────────────────────────────────────────────────────────────

type DependentUseCase struct {
	repo     domain.DependentRepository
	auditLog *audit.Logger
}

func NewDependentUseCase(repo domain.DependentRepository, auditLog *audit.Logger) *DependentUseCase {
	return &DependentUseCase{repo: repo, auditLog: auditLog}
}

func (uc *DependentUseCase) ListDependents(ctx context.Context, employeeID uuid.UUID) ([]DependentResponse, error) {
	deps, err := uc.repo.ListByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("dependent.List: %w", err)
	}
	items := make([]DependentResponse, len(deps))
	for i, d := range deps {
		items[i] = toDependentResponse(d)
	}
	return items, nil
}

func (uc *DependentUseCase) CreateDependent(ctx context.Context, employeeID uuid.UUID, req CreateDependentRequest, callerID *uuid.UUID, ip string) (*DependentResponse, error) {
	p := domain.CreateDependentParams{
		EmployeeID:             employeeID,
		FullName:               req.FullName,
		Relationship:           req.Relationship,
		CccdOrBirthCert:        req.CccdOrBirthCert,
		TaxDeductionRegistered: req.TaxDeductionRegistered,
		Notes:                  req.Notes,
	}
	if req.DateOfBirth != nil {
		if t, err := time.Parse("2006-01-02", *req.DateOfBirth); err == nil {
			p.DateOfBirth = &t
		}
	}
	if req.TaxDeductionFrom != nil {
		if t, err := time.Parse("2006-01-02", *req.TaxDeductionFrom); err == nil {
			p.TaxDeductionFrom = &t
		}
	}
	if req.TaxDeductionTo != nil {
		if t, err := time.Parse("2006-01-02", *req.TaxDeductionTo); err == nil {
			p.TaxDeductionTo = &t
		}
	}

	d, err := uc.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	if uc.auditLog != nil {
		did := d.ID
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "employee_dependents",
			ResourceID: &did, Action: "CREATE", IPAddress: ip,
		})
	}
	resp := toDependentResponse(d)
	return &resp, nil
}

func (uc *DependentUseCase) UpdateDependent(ctx context.Context, id, employeeID uuid.UUID, req UpdateDependentRequest, callerID *uuid.UUID, ip string) (*DependentResponse, error) {
	p := domain.UpdateDependentParams{
		ID:                     id,
		EmployeeID:             employeeID,
		FullName:               req.FullName,
		Relationship:           req.Relationship,
		CccdOrBirthCert:        req.CccdOrBirthCert,
		TaxDeductionRegistered: req.TaxDeductionRegistered,
		Notes:                  req.Notes,
	}
	parseDate := func(s *string) *time.Time {
		if s == nil {
			return nil
		}
		if t, err := time.Parse("2006-01-02", *s); err == nil {
			return &t
		}
		return nil
	}
	p.DateOfBirth = parseDate(req.DateOfBirth)
	p.TaxDeductionFrom = parseDate(req.TaxDeductionFrom)
	p.TaxDeductionTo = parseDate(req.TaxDeductionTo)

	d, err := uc.repo.Update(ctx, p)
	if err != nil {
		return nil, err
	}

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "employee_dependents",
			ResourceID: &id, Action: "UPDATE", IPAddress: ip,
		})
	}
	resp := toDependentResponse(d)
	return &resp, nil
}

func (uc *DependentUseCase) DeleteDependent(ctx context.Context, id, employeeID uuid.UUID, callerID *uuid.UUID, ip string) error {
	if err := uc.repo.Delete(ctx, id, employeeID); err != nil {
		return err
	}
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "employee_dependents",
			ResourceID: &id, Action: "DELETE", IPAddress: ip,
		})
	}
	return nil
}

func toDependentResponse(d *domain.EmployeeDependent) DependentResponse {
	return DependentResponse{
		ID:                     d.ID,
		EmployeeID:             d.EmployeeID,
		FullName:               d.FullName,
		Relationship:           d.Relationship,
		DateOfBirth:            dateStr(d.DateOfBirth),
		CccdOrBirthCert:        d.CccdOrBirthCert,
		TaxDeductionRegistered: d.TaxDeductionRegistered,
		TaxDeductionFrom:       dateStr(d.TaxDeductionFrom),
		TaxDeductionTo:         dateStr(d.TaxDeductionTo),
		Notes:                  d.Notes,
		CreatedAt:              d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              d.UpdatedAt.Format(time.RFC3339),
	}
}
