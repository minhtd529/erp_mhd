package domain

import (
	"time"

	"github.com/google/uuid"
)

type DeadlineType string

const (
	DeadlineVATFiling          DeadlineType = "VAT_FILING"
	DeadlineCorporateTax       DeadlineType = "CORPORATE_TAX"
	DeadlinePersonalTax        DeadlineType = "PERSONAL_TAX"
	DeadlineComplianceReporting DeadlineType = "COMPLIANCE_REPORTING"
	DeadlineCustom             DeadlineType = "CUSTOM"
)

type DeadlineStatus string

const (
	DeadlineStatusNotDue    DeadlineStatus = "NOT_DUE"
	DeadlineStatusDueSoon   DeadlineStatus = "DUE_SOON"
	DeadlineStatusOverdue   DeadlineStatus = "OVERDUE"
	DeadlineStatusCompleted DeadlineStatus = "COMPLETED"
)

type SubmissionStatus string

const (
	SubmissionPending   SubmissionStatus = "PENDING"
	SubmissionSubmitted SubmissionStatus = "SUBMITTED"
	SubmissionLate      SubmissionStatus = "LATE"
	SubmissionWaived    SubmissionStatus = "WAIVED"
)

type AdvisoryType string

const (
	AdvisoryTaxConsultation AdvisoryType = "TAX_CONSULTATION"
	AdvisoryBusinessAdvisory AdvisoryType = "BUSINESS_ADVISORY"
	AdvisoryComplianceReview AdvisoryType = "COMPLIANCE_REVIEW"
)

type AdvisoryStatus string

const (
	AdvisoryDrafted   AdvisoryStatus = "DRAFTED"
	AdvisoryDelivered AdvisoryStatus = "DELIVERED"
	AdvisoryActedOn   AdvisoryStatus = "ACTED_ON"
)

type RuleSeverity string

const (
	SeverityLow    RuleSeverity = "LOW"
	SeverityMedium RuleSeverity = "MEDIUM"
	SeverityHigh   RuleSeverity = "HIGH"
)

type TaxDeadline struct {
	ID                     uuid.UUID        `json:"id"`
	ClientID               uuid.UUID        `json:"client_id"`
	DeadlineType           DeadlineType     `json:"deadline_type"`
	DeadlineName           string           `json:"deadline_name"`
	DueDate                time.Time        `json:"due_date"`
	Status                 DeadlineStatus   `json:"status"`
	ExpectedSubmissionDate *time.Time       `json:"expected_submission_date"`
	ActualSubmissionDate   *time.Time       `json:"actual_submission_date"`
	SubmissionStatus       *SubmissionStatus `json:"submission_status"`
	Notes                  string           `json:"notes"`
	CreatedBy              uuid.UUID        `json:"created_by"`
	UpdatedBy              *uuid.UUID       `json:"updated_by"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
}

type AdvisoryRecord struct {
	ID             uuid.UUID      `json:"id"`
	ClientID       uuid.UUID      `json:"client_id"`
	EngagementID   *uuid.UUID     `json:"engagement_id"`
	AdvisoryType   AdvisoryType   `json:"advisory_type"`
	Recommendation string         `json:"recommendation"`
	Findings       string         `json:"findings"`
	Status         AdvisoryStatus `json:"status"`
	DeliveredDate  *time.Time     `json:"delivered_date"`
	ClientFeedback string         `json:"client_feedback"`
	CreatedBy      uuid.UUID      `json:"created_by"`
	UpdatedBy      *uuid.UUID     `json:"updated_by"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type AdvisoryFile struct {
	ID         uuid.UUID `json:"id"`
	AdvisoryID uuid.UUID `json:"advisory_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	CreatedBy  uuid.UUID `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type TaxComplianceRule struct {
	ID          uuid.UUID    `json:"id"`
	RuleType    string       `json:"rule_type"`
	Description string       `json:"description"`
	CheckQuery  string       `json:"check_query"`
	Severity    RuleSeverity `json:"severity"`
	IsActive    bool         `json:"is_active"`
	CreatedAt   time.Time    `json:"created_at"`
}

// ComplianceStatus is the materialized view result for one client.
type ComplianceStatus struct {
	ClientID        uuid.UUID `json:"client_id"`
	TotalDeadlines  int64     `json:"total_deadlines"`
	Completed       int64     `json:"completed"`
	Overdue         int64     `json:"overdue"`
	DueSoon         int64     `json:"due_soon"`
	ComplianceScore int64     `json:"compliance_score"`
}
