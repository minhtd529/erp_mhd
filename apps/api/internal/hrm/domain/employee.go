package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Employee errors
var (
	ErrEmployeeNotFound      = errors.New("EMPLOYEE_NOT_FOUND")
	ErrEmployeeEmailConflict = errors.New("EMPLOYEE_EMAIL_CONFLICT")
	ErrInvalidBranchDept     = errors.New("INVALID_BRANCH_DEPT")
	ErrDependentNotFound     = errors.New("DEPENDENT_NOT_FOUND")
	ErrContractNotFound      = errors.New("CONTRACT_NOT_FOUND")
	ErrContractActiveExists  = errors.New("CONTRACT_ACTIVE_EXISTS")
)

// Employee is the full HRM employee entity (columns from migrations 000004 + 000021).
type Employee struct {
	// Base columns (migration 000004)
	ID              uuid.UUID
	FullName        string
	Email           string
	Phone           *string
	DateOfBirth     *time.Time
	Grade           string
	Position        *string
	OfficeID        *uuid.UUID
	ManagerID       *uuid.UUID
	HourlyRate      *float64
	Status          string
	EmploymentDate  *time.Time
	ContractEndDate *time.Time
	IsDeleted       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CreatedBy       *uuid.UUID
	UpdatedBy       *uuid.UUID
	// Extended columns (migration 000021)
	EmployeeCode        *string
	UserID              *uuid.UUID
	BranchID            *uuid.UUID
	DepartmentID        *uuid.UUID
	PositionTitle       *string
	EmploymentType      string
	HiredDate           *time.Time
	ProbationEndDate    *time.Time
	TerminationDate     *time.Time
	TerminationReason   *string
	CurrentContractID   *uuid.UUID
	DisplayName         *string
	Gender              *string
	PlaceOfBirth        *string
	Nationality         *string
	Ethnicity           *string
	PersonalEmail       *string
	PersonalPhone       *string
	WorkPhone           *string
	CurrentAddress      *string
	PermanentAddress    *string
	CccdEncrypted       *string
	CccdIssuedDate      *time.Time
	CccdIssuedPlace     *string
	PassportNumber      *string
	PassportExpiry      *time.Time
	HiredSource         *string
	ReferrerEmployeeID  *uuid.UUID
	ProbationSalaryPct  *float64
	WorkLocation        string
	RemoteDaysPerWeek   *int16
	EducationLevel      *string
	EducationMajor      *string
	EducationSchool     *string
	EducationGraduationYear *int16
	VnCpaNumber         *string
	VnCpaIssuedDate     *time.Time
	VnCpaExpiryDate     *time.Time
	PracticingCertNumber *string
	PracticingCertExpiry *time.Time
	BaseSalary          *float64
	SalaryCurrency      *string
	SalaryEffectiveDate *time.Time
	BankAccountEncrypted *string
	BankName            *string
	BankBranch          *string
	MstCaNhanEncrypted  *string
	CommissionRate      *float64
	CommissionType      string
	SalesTargetYearly   *float64
	BizDevRegion        *string
	SoBhxhEncrypted             *string
	BhxhRegisteredDate          *time.Time
	BhxhProvinceCode            *string
	BhytCardNumber              *string
	BhytExpiryDate              *time.Time
	BhytRegisteredHospitalCode  *string
	BhytRegisteredHospitalName  *string
	TncnRegistered              bool
}

// EmployeeDependent tracks dependents for TNCN deduction.
type EmployeeDependent struct {
	ID                     uuid.UUID
	EmployeeID             uuid.UUID
	FullName               string
	Relationship           string
	DateOfBirth            *time.Time
	CccdOrBirthCert        *string
	TaxDeductionRegistered bool
	TaxDeductionFrom       *time.Time
	TaxDeductionTo         *time.Time
	Notes                  *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// EmploymentContract represents an employment_contracts row.
type EmploymentContract struct {
	ID                uuid.UUID
	EmployeeID        uuid.UUID
	ContractNumber    *string
	ContractType      string
	StartDate         time.Time
	EndDate           *time.Time
	SignedDate        *time.Time
	SalaryAtSigning   *float64
	PositionAtSigning *string
	Notes             *string
	DocumentURL       *string
	IsCurrent         bool
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ─── Filters / Params ─────────────────────────────────────────────────────────

type ListEmployeesFilter struct {
	Page         int
	Size         int
	BranchID     *uuid.UUID
	DepartmentID *uuid.UUID
	Status       *string
	Grade        *string
	Q            string
	UserID       *uuid.UUID // scope: only this user's employee record
	BranchScope  *uuid.UUID // scope: only employees in this branch
}

type CreateEmployeeParams struct {
	FullName       string
	Email          string
	Phone          *string
	DateOfBirth    *time.Time
	Grade          string
	ManagerID      *uuid.UUID
	Status         string
	BranchID       *uuid.UUID
	DepartmentID   *uuid.UUID
	PositionTitle  *string
	EmploymentType string
	HiredDate      *time.Time
	DisplayName    *string
	Gender         *string
	PersonalEmail  *string
	PersonalPhone  *string
	WorkLocation   string
	HiredSource    *string
	EducationLevel *string
	CommissionType string
	CreatedBy      *uuid.UUID
}

type UpdateEmployeeParams struct {
	ID                      uuid.UUID
	FullName                *string
	Phone                   *string
	Grade                   *string
	ManagerID               *uuid.UUID
	Status                  *string
	BranchID                *uuid.UUID
	DepartmentID            *uuid.UUID
	PositionTitle           *string
	EmploymentType          *string
	HiredDate               *time.Time
	ProbationEndDate        *time.Time
	TerminationDate         *time.Time
	TerminationReason       *string
	CurrentContractID       *uuid.UUID
	DisplayName             *string
	Gender                  *string
	PersonalEmail           *string
	PersonalPhone           *string
	WorkPhone               *string
	CurrentAddress          *string
	PermanentAddress        *string
	WorkLocation            *string
	RemoteDaysPerWeek       *int16
	HiredSource             *string
	EducationLevel          *string
	EducationMajor          *string
	EducationSchool         *string
	EducationGraduationYear *int16
	VnCpaNumber             *string
	PracticingCertNumber    *string
	CommissionType          *string
	CommissionRate          *float64
	BizDevRegion            *string
	Nationality             *string
	Ethnicity               *string
	UpdatedBy               *uuid.UUID
}

type UpdateProfileParams struct {
	UserID           uuid.UUID
	DisplayName      *string
	PersonalPhone    *string
	PersonalEmail    *string
	CurrentAddress   *string
	PermanentAddress *string
	UpdatedBy        *uuid.UUID
}

type CreateDependentParams struct {
	EmployeeID             uuid.UUID
	FullName               string
	Relationship           string
	DateOfBirth            *time.Time
	CccdOrBirthCert        *string
	TaxDeductionRegistered bool
	TaxDeductionFrom       *time.Time
	TaxDeductionTo         *time.Time
	Notes                  *string
}

type UpdateDependentParams struct {
	ID                     uuid.UUID
	EmployeeID             uuid.UUID
	FullName               *string
	Relationship           *string
	DateOfBirth            *time.Time
	CccdOrBirthCert        *string
	TaxDeductionRegistered *bool
	TaxDeductionFrom       *time.Time
	TaxDeductionTo         *time.Time
	Notes                  *string
}

type CreateContractParams struct {
	EmployeeID        uuid.UUID
	ContractNumber    *string
	ContractType      string
	StartDate         time.Time
	EndDate           *time.Time
	SignedDate        *time.Time
	SalaryAtSigning   *float64
	PositionAtSigning *string
	Notes             *string
	DocumentURL       *string
	CreatedBy         *uuid.UUID
}

type UpdateContractParams struct {
	ID                uuid.UUID
	EmployeeID        uuid.UUID
	ContractType      *string
	EndDate           *time.Time
	SignedDate        *time.Time
	SalaryAtSigning   *float64
	PositionAtSigning *string
	Notes             *string
	DocumentURL       *string
}

// UpdateSensitiveParams carries the encrypted field values to persist.
type UpdateSensitiveParams struct {
	ID                   uuid.UUID
	CccdEncrypted        *string
	CccdIssuedDate       *time.Time
	CccdIssuedPlace      *string
	PassportNumber       *string
	PassportExpiry       *time.Time
	MstCaNhanEncrypted   *string
	SoBhxhEncrypted      *string
	BankAccountEncrypted *string
	BankName             *string
	BankBranch           *string
	UpdatedBy            *uuid.UUID
}

// SalaryHistory is one row from employee_salary_history.
type SalaryHistory struct {
	ID              uuid.UUID
	EmployeeID      uuid.UUID
	EffectiveDate   time.Time
	BaseSalary      float64
	AllowancesTotal float64
	SalaryNote      *string
	ChangeType      string
	ApprovedBy      *uuid.UUID
	CreatedBy       *uuid.UUID
	CreatedByName   *string // populated via JOIN with users
	CreatedAt       time.Time
}

// CreateSalaryHistoryParams carries validated inputs for a new salary record.
type CreateSalaryHistoryParams struct {
	EmployeeID      uuid.UUID
	EffectiveDate   time.Time
	BaseSalary      float64
	AllowancesTotal float64
	SalaryNote      *string
	ChangeType      string
	ApprovedBy      *uuid.UUID
	CreatedBy       *uuid.UUID
}

// ─── Repository interfaces ────────────────────────────────────────────────────

type EmployeeRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Employee, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*Employee, error)
	List(ctx context.Context, f ListEmployeesFilter) ([]*Employee, int64, error)
	Create(ctx context.Context, p CreateEmployeeParams) (*Employee, error)
	Update(ctx context.Context, p UpdateEmployeeParams) (*Employee, error)
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	UpdateProfile(ctx context.Context, p UpdateProfileParams) (*Employee, error)
	UpdateSensitiveFields(ctx context.Context, p UpdateSensitiveParams) error
}

type SalaryHistoryRepository interface {
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*SalaryHistory, error)
	Create(ctx context.Context, p CreateSalaryHistoryParams) (*SalaryHistory, error)
}

type DependentRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*EmployeeDependent, error)
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeDependent, error)
	Create(ctx context.Context, p CreateDependentParams) (*EmployeeDependent, error)
	Update(ctx context.Context, p UpdateDependentParams) (*EmployeeDependent, error)
	Delete(ctx context.Context, id, employeeID uuid.UUID) error
}

// ContractExpiryAlert is returned by ListExpiringContracts for the HRM reminder job.
type ContractExpiryAlert struct {
	ContractID uuid.UUID
	EmployeeID uuid.UUID
	UserID     uuid.UUID
	EndDate    time.Time
}

type ContractRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*EmploymentContract, error)
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]*EmploymentContract, error)
	HasActiveContract(ctx context.Context, employeeID uuid.UUID) (bool, error)
	Create(ctx context.Context, p CreateContractParams) (*EmploymentContract, error)
	Update(ctx context.Context, p UpdateContractParams) (*EmploymentContract, error)
	Terminate(ctx context.Context, id, employeeID uuid.UUID) error
	// ListExpiringContracts returns current contracts ending within withinDays days,
	// enriched with the employee's user_id. Used by the HRM daily reminder job.
	ListExpiringContracts(ctx context.Context, withinDays int) ([]ContractExpiryAlert, error)
}
