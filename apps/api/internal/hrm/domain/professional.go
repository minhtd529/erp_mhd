package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Professional development errors
var (
	ErrCertificationNotFound   = errors.New("CERTIFICATION_NOT_FOUND")
	ErrTrainingCourseNotFound  = errors.New("TRAINING_COURSE_NOT_FOUND")
	ErrTrainingRecordNotFound  = errors.New("TRAINING_RECORD_NOT_FOUND")
	ErrCPERequirementNotFound  = errors.New("CPE_REQUIREMENT_NOT_FOUND")
	ErrDuplicateCourseCode     = errors.New("DUPLICATE_COURSE_CODE")
	ErrDuplicateCPERequirement = errors.New("DUPLICATE_CPE_REQUIREMENT")
)

// ─── Entities ─────────────────────────────────────────────────────────────────

// Certification maps to the certifications table (SPEC §11.9).
type Certification struct {
	ID               uuid.UUID
	EmployeeID       uuid.UUID
	CertType         string
	CertName         string
	CertNumber       *string
	IssuedDate       *time.Time
	ExpiryDate       *time.Time
	IssuingAuthority *string
	Status           string
	DocumentURL      *string
	Notes            *string
	IsDeleted        bool
	CreatedBy        *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// TrainingCourse maps to the training_courses table (SPEC §11.10).
type TrainingCourse struct {
	ID          uuid.UUID
	Code        string
	Name        string
	Provider    *string
	Description *string
	CpeHours    float64
	CourseType  string
	IsActive    bool
	Notes       *string
	CreatedBy   *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TrainingRecord maps to the training_records table (SPEC §11.11).
type TrainingRecord struct {
	ID             uuid.UUID
	EmployeeID     uuid.UUID
	CourseID       uuid.UUID
	CompletionDate *time.Time
	CpeHoursEarned float64
	CertificateURL *string
	Status         string
	Notes          *string
	IsDeleted      bool
	CreatedBy      *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CPERequirement maps to the cpe_requirements_by_role table (SPEC §11.12).
type CPERequirement struct {
	ID                uuid.UUID
	RoleCode          string
	Year              int16
	RequiredHours     float64
	CategoryBreakdown []byte // JSONB
	Notes             *string
	CreatedBy         *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CPESummary aggregates training_records for one employee/year.
type CPESummary struct {
	EmployeeID     uuid.UUID
	Year           int
	TotalHours     float64
	RequiredHours  float64  // from cpe_requirements_by_role for the employee's role
	ByCategory     []byte   // JSONB: map[category]hours
}

// ─── Request params ────────────────────────────────────────────────────────────

type CreateCertificationParams struct {
	EmployeeID       uuid.UUID
	CertType         string
	CertName         string
	CertNumber       *string
	IssuedDate       *time.Time
	ExpiryDate       *time.Time
	IssuingAuthority *string
	Status           string
	DocumentURL      *string
	Notes            *string
	CreatedBy        *uuid.UUID
}

type UpdateCertificationParams struct {
	ID               uuid.UUID
	CertType         string
	CertName         string
	CertNumber       *string
	IssuedDate       *time.Time
	ExpiryDate       *time.Time
	IssuingAuthority *string
	Status           string
	DocumentURL      *string
	Notes            *string
}

type CreateTrainingCourseParams struct {
	Code        string
	Name        string
	Provider    *string
	Description *string
	CpeHours    float64
	CourseType  string
	IsActive    bool
	Notes       *string
	CreatedBy   *uuid.UUID
}

type UpdateTrainingCourseParams struct {
	ID          uuid.UUID
	Name        string
	Provider    *string
	Description *string
	CpeHours    float64
	CourseType  string
	IsActive    bool
	Notes       *string
}

type CreateTrainingRecordParams struct {
	EmployeeID     uuid.UUID
	CourseID       uuid.UUID
	CompletionDate *time.Time
	CpeHoursEarned float64
	CertificateURL *string
	Status         string
	Notes          *string
	CreatedBy      *uuid.UUID
}

type UpdateTrainingRecordParams struct {
	ID             uuid.UUID
	CompletionDate *time.Time
	CpeHoursEarned float64
	CertificateURL *string
	Status         string
	Notes          *string
}

type ListTrainingCoursesFilter struct {
	CourseType string
	IsActive   *bool
	Q          string
	Page       int
	Size       int
}

type CreateCPERequirementParams struct {
	RoleCode          string
	Year              int16
	RequiredHours     float64
	CategoryBreakdown []byte
	Notes             *string
	CreatedBy         *uuid.UUID
}

type UpdateCPERequirementParams struct {
	ID                uuid.UUID
	RequiredHours     float64
	CategoryBreakdown []byte
	Notes             *string
}

// ─── Repository interfaces ─────────────────────────────────────────────────────

type CertificationRepository interface {
	Create(ctx context.Context, p CreateCertificationParams) (*Certification, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Certification, error)
	ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*Certification, error)
	Update(ctx context.Context, p UpdateCertificationParams) (*Certification, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ListExpiring(ctx context.Context, withinDays int) ([]*Certification, error)
}

type TrainingCourseRepository interface {
	Create(ctx context.Context, p CreateTrainingCourseParams) (*TrainingCourse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*TrainingCourse, error)
	List(ctx context.Context, f ListTrainingCoursesFilter) ([]*TrainingCourse, int64, error)
	Update(ctx context.Context, p UpdateTrainingCourseParams) (*TrainingCourse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type TrainingRecordRepository interface {
	Create(ctx context.Context, p CreateTrainingRecordParams) (*TrainingRecord, error)
	FindByID(ctx context.Context, id uuid.UUID) (*TrainingRecord, error)
	ListByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*TrainingRecord, error)
	Update(ctx context.Context, p UpdateTrainingRecordParams) (*TrainingRecord, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetCPESummary(ctx context.Context, employeeID uuid.UUID, year int) (*CPESummary, error)
}

type CPERequirementRepository interface {
	Create(ctx context.Context, p CreateCPERequirementParams) (*CPERequirement, error)
	FindByID(ctx context.Context, id uuid.UUID) (*CPERequirement, error)
	List(ctx context.Context, roleCode string, year int) ([]*CPERequirement, error)
	Update(ctx context.Context, p UpdateCPERequirementParams) (*CPERequirement, error)
}
