package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
)

// EmployeeCreateRequest is the body for POST /api/v1/employees.
type EmployeeCreateRequest struct {
	FullName                string               `json:"full_name"                  binding:"required"`
	Email                   string               `json:"email"                      binding:"required,email"`
	Phone                   *string              `json:"phone"`
	DateOfBirth             *time.Time           `json:"date_of_birth"`
	Grade                   domain.EmployeeGrade `json:"grade"                      binding:"required"`
	Position                *string              `json:"position"`
	OfficeID                *uuid.UUID           `json:"office_id"`
	ManagerID               *uuid.UUID           `json:"manager_id"`
	HourlyRate              *float64             `json:"hourly_rate"`
	EmploymentDate          *time.Time           `json:"employment_date"`
	ContractEndDate         *time.Time           `json:"contract_end_date"`
	IsSalesperson           bool                 `json:"is_salesperson"`
	SalesCommissionEligible bool                 `json:"sales_commission_eligible"`
}

// EmployeeUpdateRequest is the body for PUT /api/v1/employees/:id.
type EmployeeUpdateRequest struct {
	FullName                string               `json:"full_name"                  binding:"required"`
	Phone                   *string              `json:"phone"`
	DateOfBirth             *time.Time           `json:"date_of_birth"`
	Grade                   domain.EmployeeGrade `json:"grade"                      binding:"required"`
	Position                *string              `json:"position"`
	OfficeID                *uuid.UUID           `json:"office_id"`
	ManagerID               *uuid.UUID           `json:"manager_id"`
	HourlyRate              *float64             `json:"hourly_rate"`
	EmploymentDate          *time.Time           `json:"employment_date"`
	ContractEndDate         *time.Time           `json:"contract_end_date"`
	IsSalesperson           bool                 `json:"is_salesperson"`
	SalesCommissionEligible bool                 `json:"sales_commission_eligible"`
}

// EmployeeBankDetailsRequest is the body for PUT /api/v1/employees/:id/bank-details.
// The plain-text account number is accepted here; encryption happens in the usecase.
type EmployeeBankDetailsRequest struct {
	BankAccountNumber string  `json:"bank_account_number" binding:"required"`
	BankAccountName   *string `json:"bank_account_name"`
}

// EmployeeListRequest holds validated query parameters.
type EmployeeListRequest struct {
	Page     int                   `form:"page,default=1"  binding:"min=1"`
	Size     int                   `form:"size,default=20" binding:"min=1,max=100"`
	Status   domain.EmployeeStatus `form:"status"`
	OfficeID *uuid.UUID            `form:"office_id"`
	Q        string                `form:"q"`
}

// EmployeeResponse is the JSON representation of an employee.
// Bank account fields are intentionally omitted (json:"-" in domain entity).
type EmployeeResponse struct {
	ID                      uuid.UUID             `json:"id"`
	FullName                string                `json:"full_name"`
	Email                   string                `json:"email"`
	Phone                   *string               `json:"phone"`
	DateOfBirth             *time.Time            `json:"date_of_birth"`
	Grade                   domain.EmployeeGrade  `json:"grade"`
	Position                *string               `json:"position"`
	OfficeID                *uuid.UUID            `json:"office_id"`
	ManagerID               *uuid.UUID            `json:"manager_id"`
	HourlyRate              *float64              `json:"hourly_rate"`
	Status                  domain.EmployeeStatus `json:"status"`
	EmploymentDate          *time.Time            `json:"employment_date"`
	ContractEndDate         *time.Time            `json:"contract_end_date"`
	IsSalesperson           bool                  `json:"is_salesperson"`
	SalesCommissionEligible bool                  `json:"sales_commission_eligible"`
	CreatedAt               time.Time             `json:"created_at"`
	UpdatedAt               time.Time             `json:"updated_at"`
	CreatedBy               *uuid.UUID            `json:"created_by"`
	UpdatedBy               *uuid.UUID            `json:"updated_by"`
}

// PaginatedResult is a generic paginated wrapper.
type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalPages int   `json:"total_pages"`
}

func newPaginatedResult[T any](data []T, total int64, page, size int) PaginatedResult[T] {
	tp := int(total) / size
	if int(total)%size != 0 {
		tp++
	}
	return PaginatedResult[T]{Data: data, Total: total, Page: page, Size: size, TotalPages: tp}
}

func toEmployeeResponse(e *domain.Employee) EmployeeResponse {
	return EmployeeResponse{
		ID: e.ID, FullName: e.FullName, Email: e.Email,
		Phone: e.Phone, DateOfBirth: e.DateOfBirth,
		Grade: e.Grade, Position: e.Position,
		OfficeID: e.OfficeID, ManagerID: e.ManagerID,
		HourlyRate: e.HourlyRate, Status: e.Status,
		EmploymentDate: e.EmploymentDate, ContractEndDate: e.ContractEndDate,
		IsSalesperson:           e.IsSalesperson,
		SalesCommissionEligible: e.SalesCommissionEligible,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
		CreatedBy: e.CreatedBy, UpdatedBy: e.UpdatedBy,
	}
}
