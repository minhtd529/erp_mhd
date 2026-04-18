package domain

import (
	"time"

	"github.com/google/uuid"
)

// EmployeeGrade represents seniority level.
type EmployeeGrade string

const (
	GradeIntern    EmployeeGrade = "INTERN"
	GradeJunior    EmployeeGrade = "JUNIOR"
	GradeSenior    EmployeeGrade = "SENIOR"
	GradeManager   EmployeeGrade = "MANAGER"
	GradeDirector  EmployeeGrade = "DIRECTOR"
	GradePartner   EmployeeGrade = "PARTNER"
)

// EmployeeStatus represents the employment state.
type EmployeeStatus string

const (
	StatusActive   EmployeeStatus = "ACTIVE"
	StatusOnLeave  EmployeeStatus = "ON_LEAVE"
	StatusResigned EmployeeStatus = "RESIGNED"
	StatusRetired  EmployeeStatus = "RETIRED"
)

// Employee is the HRM aggregate root.
type Employee struct {
	ID                       uuid.UUID      `json:"id"                         db:"id"`
	FullName                 string         `json:"full_name"                  db:"full_name"`
	Email                    string         `json:"email"                      db:"email"`
	Phone                    *string        `json:"phone"                      db:"phone"`
	DateOfBirth              *time.Time     `json:"date_of_birth"              db:"date_of_birth"`
	Grade                    EmployeeGrade  `json:"grade"                      db:"grade"`
	Position                 *string        `json:"position"                   db:"position"`
	OfficeID                 *uuid.UUID     `json:"office_id"                  db:"office_id"`
	ManagerID                *uuid.UUID     `json:"manager_id"                 db:"manager_id"`
	HourlyRate               *float64       `json:"hourly_rate"                db:"hourly_rate"`
	Status                   EmployeeStatus `json:"status"                     db:"status"`
	EmploymentDate           *time.Time     `json:"employment_date"            db:"employment_date"`
	ContractEndDate          *time.Time     `json:"contract_end_date"          db:"contract_end_date"`
	IsSalesperson            bool           `json:"is_salesperson"             db:"is_salesperson"`
	SalesCommissionEligible  bool           `json:"sales_commission_eligible"  db:"sales_commission_eligible"`
	// BankAccountNumberEnc is the AES-256-GCM ciphertext of the bank account number.
	// It is NEVER included in list/detail API responses; set via a dedicated endpoint.
	BankAccountNumberEnc     *string        `json:"-"                          db:"bank_account_number_enc"`
	BankAccountName          *string        `json:"-"                          db:"bank_account_name"`
	IsDeleted                bool           `json:"is_deleted"                 db:"is_deleted"`
	CreatedAt                time.Time      `json:"created_at"                 db:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"                 db:"updated_at"`
	CreatedBy                *uuid.UUID     `json:"created_by"                 db:"created_by"`
	UpdatedBy                *uuid.UUID     `json:"updated_by"                 db:"updated_by"`
}

// CreateEmployeeParams holds the parameters needed to create a new employee.
type CreateEmployeeParams struct {
	FullName                string
	Email                   string
	Phone                   *string
	DateOfBirth             *time.Time
	Grade                   EmployeeGrade
	Position                *string
	OfficeID                *uuid.UUID
	ManagerID               *uuid.UUID
	HourlyRate              *float64
	EmploymentDate          *time.Time
	ContractEndDate         *time.Time
	IsSalesperson           bool
	SalesCommissionEligible bool
	CreatedBy               *uuid.UUID
}

// UpdateEmployeeParams holds the fields that can be changed after creation.
type UpdateEmployeeParams struct {
	ID                      uuid.UUID
	FullName                string
	Phone                   *string
	DateOfBirth             *time.Time
	Grade                   EmployeeGrade
	Position                *string
	OfficeID                *uuid.UUID
	ManagerID               *uuid.UUID
	HourlyRate              *float64
	EmploymentDate          *time.Time
	ContractEndDate         *time.Time
	IsSalesperson           bool
	SalesCommissionEligible bool
	UpdatedBy               *uuid.UUID
}

// UpdateBankDetailsParams carries the encrypted bank fields for a dedicated
// bank-details update operation (never mixed with general profile update).
type UpdateBankDetailsParams struct {
	ID                   uuid.UUID
	BankAccountNumberEnc *string // AES-256-GCM ciphertext; nil clears the field
	BankAccountName      *string
	UpdatedBy            *uuid.UUID
}

// ListEmployeesFilter controls pagination and filtering.
type ListEmployeesFilter struct {
	Page          int
	Size          int
	Status        EmployeeStatus
	OfficeID      *uuid.UUID
	Q             string // full-text search via tsvector
	IsSalesperson *bool  // filter by salesperson flag
}
