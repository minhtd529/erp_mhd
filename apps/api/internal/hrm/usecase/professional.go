package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// hrManagerRoles are the roles that can manage all employees' professional data.
var hrManagerRoles = map[string]struct{}{
	"SUPER_ADMIN": {},
	"HR_MANAGER":  {},
	"CHAIRMAN":    {},
	"CEO":         {},
}

func isHRManager(roles []string) bool {
	for _, r := range roles {
		if _, ok := hrManagerRoles[r]; ok {
			return true
		}
	}
	return false
}

// ─── Response DTOs ─────────────────────────────────────────────────────────────

type CertificationResponse struct {
	ID               uuid.UUID  `json:"id"`
	EmployeeID       uuid.UUID  `json:"employee_id"`
	CertType         string     `json:"cert_type"`
	CertName         string     `json:"cert_name"`
	CertNumber       *string    `json:"cert_number,omitempty"`
	IssuedDate       *string    `json:"issued_date,omitempty"`
	ExpiryDate       *string    `json:"expiry_date,omitempty"`
	IssuingAuthority *string    `json:"issuing_authority,omitempty"`
	Status           string     `json:"status"`
	DocumentURL      *string    `json:"document_url,omitempty"`
	Notes            *string    `json:"notes,omitempty"`
	CreatedBy        *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt        string     `json:"created_at"`
	UpdatedAt        string     `json:"updated_at"`
}

type TrainingCourseResponse struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Provider    *string    `json:"provider,omitempty"`
	Description *string    `json:"description,omitempty"`
	CpeHours    float64    `json:"cpe_hours"`
	CourseType  string     `json:"course_type"`
	IsActive    bool       `json:"is_active"`
	Notes       *string    `json:"notes,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

type TrainingRecordResponse struct {
	ID             uuid.UUID  `json:"id"`
	EmployeeID     uuid.UUID  `json:"employee_id"`
	CourseID       uuid.UUID  `json:"course_id"`
	CompletionDate *string    `json:"completion_date,omitempty"`
	CpeHoursEarned float64    `json:"cpe_hours_earned"`
	CertificateURL *string    `json:"certificate_url,omitempty"`
	Status         string     `json:"status"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt      string     `json:"created_at"`
	UpdatedAt      string     `json:"updated_at"`
}

type CPERequirementResponse struct {
	ID                uuid.UUID       `json:"id"`
	RoleCode          string          `json:"role_code"`
	Year              int16           `json:"year"`
	RequiredHours     float64         `json:"required_hours"`
	CategoryBreakdown json.RawMessage `json:"category_breakdown,omitempty"`
	Notes             *string         `json:"notes,omitempty"`
	CreatedBy         *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt         string          `json:"created_at"`
	UpdatedAt         string          `json:"updated_at"`
}

type CPESummaryResponse struct {
	EmployeeID    uuid.UUID       `json:"employee_id"`
	Year          int             `json:"year"`
	TotalHours    float64         `json:"total_hours"`
	RequiredHours float64         `json:"required_hours"`
	ByCategory    json.RawMessage `json:"by_category"`
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type CreateCertificationRequest struct {
	CertType         string  `json:"cert_type"  binding:"required"`
	CertName         string  `json:"cert_name"  binding:"required"`
	CertNumber       *string `json:"cert_number"`
	IssuedDate       *string `json:"issued_date"`
	ExpiryDate       *string `json:"expiry_date"`
	IssuingAuthority *string `json:"issuing_authority"`
	Status           string  `json:"status"`
	DocumentURL      *string `json:"document_url"`
	Notes            *string `json:"notes"`
}

type UpdateCertificationRequest struct {
	CertType         string  `json:"cert_type"  binding:"required"`
	CertName         string  `json:"cert_name"  binding:"required"`
	CertNumber       *string `json:"cert_number"`
	IssuedDate       *string `json:"issued_date"`
	ExpiryDate       *string `json:"expiry_date"`
	IssuingAuthority *string `json:"issuing_authority"`
	Status           string  `json:"status"     binding:"required"`
	DocumentURL      *string `json:"document_url"`
	Notes            *string `json:"notes"`
}

type CreateTrainingCourseRequest struct {
	Code        string  `json:"code"        binding:"required"`
	Name        string  `json:"name"        binding:"required"`
	Provider    *string `json:"provider"`
	Description *string `json:"description"`
	CpeHours    float64 `json:"cpe_hours"`
	CourseType  string  `json:"course_type" binding:"required"`
	IsActive    *bool   `json:"is_active"`
	Notes       *string `json:"notes"`
}

type UpdateTrainingCourseRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Provider    *string `json:"provider"`
	Description *string `json:"description"`
	CpeHours    float64 `json:"cpe_hours"`
	CourseType  string  `json:"course_type" binding:"required"`
	IsActive    *bool   `json:"is_active"`
	Notes       *string `json:"notes"`
}

type ListTrainingCoursesRequest struct {
	CourseType string `form:"course_type"`
	IsActive   *bool  `form:"is_active"`
	Q          string `form:"q"`
	Page       int    `form:"page"`
	Size       int    `form:"size"`
}

type CreateTrainingRecordRequest struct {
	CourseID       uuid.UUID `json:"course_id"       binding:"required"`
	CompletionDate *string   `json:"completion_date"`
	CpeHoursEarned float64   `json:"cpe_hours_earned"`
	CertificateURL *string   `json:"certificate_url"`
	Status         string    `json:"status"`
	Notes          *string   `json:"notes"`
}

type UpdateTrainingRecordRequest struct {
	CompletionDate *string `json:"completion_date"`
	CpeHoursEarned float64 `json:"cpe_hours_earned"`
	CertificateURL *string `json:"certificate_url"`
	Status         string  `json:"status" binding:"required"`
	Notes          *string `json:"notes"`
}

type CreateCPERequirementRequest struct {
	RoleCode          string          `json:"role_code"      binding:"required"`
	Year              int16           `json:"year"           binding:"required"`
	RequiredHours     float64         `json:"required_hours" binding:"required"`
	CategoryBreakdown json.RawMessage `json:"category_breakdown"`
	Notes             *string         `json:"notes"`
}

type UpdateCPERequirementRequest struct {
	RequiredHours     float64         `json:"required_hours" binding:"required"`
	CategoryBreakdown json.RawMessage `json:"category_breakdown"`
	Notes             *string         `json:"notes"`
}

type ListCPERequirementsRequest struct {
	RoleCode string `form:"role_code"`
	Year     int    `form:"year"`
}

type ExpiringCertsRequest struct {
	Days int `form:"days"`
}

// ─── Converters ───────────────────────────────────────────────────────────────

func toCertResponse(c *domain.Certification) CertificationResponse {
	r := CertificationResponse{
		ID:               c.ID,
		EmployeeID:       c.EmployeeID,
		CertType:         c.CertType,
		CertName:         c.CertName,
		CertNumber:       c.CertNumber,
		IssuingAuthority: c.IssuingAuthority,
		Status:           c.Status,
		DocumentURL:      c.DocumentURL,
		Notes:            c.Notes,
		CreatedBy:        c.CreatedBy,
		CreatedAt:        c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        c.UpdatedAt.Format(time.RFC3339),
	}
	if c.IssuedDate != nil {
		s := c.IssuedDate.Format("2006-01-02")
		r.IssuedDate = &s
	}
	if c.ExpiryDate != nil {
		s := c.ExpiryDate.Format("2006-01-02")
		r.ExpiryDate = &s
	}
	return r
}

func toCourseResponse(c *domain.TrainingCourse) TrainingCourseResponse {
	return TrainingCourseResponse{
		ID:          c.ID,
		Code:        c.Code,
		Name:        c.Name,
		Provider:    c.Provider,
		Description: c.Description,
		CpeHours:    c.CpeHours,
		CourseType:  c.CourseType,
		IsActive:    c.IsActive,
		Notes:       c.Notes,
		CreatedBy:   c.CreatedBy,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}
}

func toRecordResponse(r *domain.TrainingRecord) TrainingRecordResponse {
	resp := TrainingRecordResponse{
		ID:             r.ID,
		EmployeeID:     r.EmployeeID,
		CourseID:       r.CourseID,
		CpeHoursEarned: r.CpeHoursEarned,
		CertificateURL: r.CertificateURL,
		Status:         r.Status,
		Notes:          r.Notes,
		CreatedBy:      r.CreatedBy,
		CreatedAt:      r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      r.UpdatedAt.Format(time.RFC3339),
	}
	if r.CompletionDate != nil {
		s := r.CompletionDate.Format("2006-01-02")
		resp.CompletionDate = &s
	}
	return resp
}

func toCPEResponse(c *domain.CPERequirement) CPERequirementResponse {
	resp := CPERequirementResponse{
		ID:            c.ID,
		RoleCode:      c.RoleCode,
		Year:          c.Year,
		RequiredHours: c.RequiredHours,
		Notes:         c.Notes,
		CreatedBy:     c.CreatedBy,
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     c.UpdatedAt.Format(time.RFC3339),
	}
	if len(c.CategoryBreakdown) > 0 {
		resp.CategoryBreakdown = json.RawMessage(c.CategoryBreakdown)
	}
	return resp
}

func parseDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}
	return &t, nil
}

// ─── ProfessionalUseCase ──────────────────────────────────────────────────────

type ProfessionalUseCase struct {
	certRepo   domain.CertificationRepository
	courseRepo domain.TrainingCourseRepository
	recordRepo domain.TrainingRecordRepository
	cpeRepo    domain.CPERequirementRepository
	empRepo    domain.EmployeeRepository
	auditLog   *audit.Logger
}

func NewProfessionalUseCase(
	certRepo domain.CertificationRepository,
	courseRepo domain.TrainingCourseRepository,
	recordRepo domain.TrainingRecordRepository,
	cpeRepo domain.CPERequirementRepository,
	empRepo domain.EmployeeRepository,
	auditLog *audit.Logger,
) *ProfessionalUseCase {
	return &ProfessionalUseCase{
		certRepo: certRepo, courseRepo: courseRepo,
		recordRepo: recordRepo, cpeRepo: cpeRepo,
		empRepo: empRepo, auditLog: auditLog,
	}
}

// canAccessEmployee returns true if the caller is an HR manager or the employee themselves.
func (uc *ProfessionalUseCase) canAccessEmployee(ctx context.Context, employeeID uuid.UUID, callerID uuid.UUID, callerRoles []string) (bool, error) {
	if isHRManager(callerRoles) {
		return true, nil
	}
	// Self-read: look up the employee's linked user_id
	emp, err := uc.empRepo.FindByID(ctx, employeeID)
	if err != nil {
		return false, err
	}
	if emp.UserID != nil && *emp.UserID == callerID {
		return true, nil
	}
	return false, nil
}

// ─── Certification methods ─────────────────────────────────────────────────────

func (uc *ProfessionalUseCase) ListCertifications(ctx context.Context, employeeID uuid.UUID, callerID uuid.UUID, callerRoles []string) ([]CertificationResponse, error) {
	ok, err := uc.canAccessEmployee(ctx, employeeID, callerID, callerRoles)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrInsufficientPermission
	}
	certs, err := uc.certRepo.ListByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("ListCertifications: %w", err)
	}
	out := make([]CertificationResponse, 0, len(certs))
	for _, c := range certs {
		out = append(out, toCertResponse(c))
	}
	return out, nil
}

func (uc *ProfessionalUseCase) CreateCertification(ctx context.Context, employeeID uuid.UUID, req CreateCertificationRequest, callerID uuid.UUID, ip string) (*CertificationResponse, error) {
	issuedDate, err := parseDate(req.IssuedDate)
	if err != nil {
		return nil, domain.ErrValidation
	}
	expiryDate, err := parseDate(req.ExpiryDate)
	if err != nil {
		return nil, domain.ErrValidation
	}

	status := req.Status
	if status == "" {
		status = "ACTIVE"
	}

	c, err := uc.certRepo.Create(ctx, domain.CreateCertificationParams{
		EmployeeID:       employeeID,
		CertType:         req.CertType,
		CertName:         req.CertName,
		CertNumber:       req.CertNumber,
		IssuedDate:       issuedDate,
		ExpiryDate:       expiryDate,
		IssuingAuthority: req.IssuingAuthority,
		Status:           status,
		DocumentURL:      req.DocumentURL,
		Notes:            req.Notes,
		CreatedBy:        &callerID,
	})
	if err != nil {
		return nil, fmt.Errorf("CreateCertification: %w", err)
	}

	cid := c.ID
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "certifications",
		ResourceID: &cid, Action: "CERTIFICATION_CREATED",
		NewValue: c, IPAddress: ip,
	})

	resp := toCertResponse(c)
	return &resp, nil
}

func (uc *ProfessionalUseCase) GetCertification(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRoles []string) (*CertificationResponse, error) {
	c, err := uc.certRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ok, err := uc.canAccessEmployee(ctx, c.EmployeeID, callerID, callerRoles)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrInsufficientPermission
	}
	resp := toCertResponse(c)
	return &resp, nil
}

func (uc *ProfessionalUseCase) UpdateCertification(ctx context.Context, id uuid.UUID, req UpdateCertificationRequest, callerID uuid.UUID, ip string) (*CertificationResponse, error) {
	issuedDate, err := parseDate(req.IssuedDate)
	if err != nil {
		return nil, domain.ErrValidation
	}
	expiryDate, err := parseDate(req.ExpiryDate)
	if err != nil {
		return nil, domain.ErrValidation
	}

	c, err := uc.certRepo.Update(ctx, domain.UpdateCertificationParams{
		ID:               id,
		CertType:         req.CertType,
		CertName:         req.CertName,
		CertNumber:       req.CertNumber,
		IssuedDate:       issuedDate,
		ExpiryDate:       expiryDate,
		IssuingAuthority: req.IssuingAuthority,
		Status:           req.Status,
		DocumentURL:      req.DocumentURL,
		Notes:            req.Notes,
	})
	if err != nil {
		return nil, err
	}

	cid := c.ID
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "certifications",
		ResourceID: &cid, Action: "CERTIFICATION_UPDATED",
		NewValue: c, IPAddress: ip,
	})

	resp := toCertResponse(c)
	return &resp, nil
}

func (uc *ProfessionalUseCase) DeleteCertification(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) error {
	if err := uc.certRepo.SoftDelete(ctx, id); err != nil {
		return err
	}
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "certifications",
		ResourceID: &id, Action: "CERTIFICATION_DELETED", IPAddress: ip,
	})
	return nil
}

func (uc *ProfessionalUseCase) ListExpiringCertifications(ctx context.Context, days int) ([]CertificationResponse, error) {
	if days <= 0 {
		days = 90
	}
	certs, err := uc.certRepo.ListExpiring(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("ListExpiringCertifications: %w", err)
	}
	out := make([]CertificationResponse, 0, len(certs))
	for _, c := range certs {
		out = append(out, toCertResponse(c))
	}
	return out, nil
}

// ─── Training course methods ───────────────────────────────────────────────────

func (uc *ProfessionalUseCase) ListTrainingCourses(ctx context.Context, req ListTrainingCoursesRequest) (*pagination.OffsetResult[TrainingCourseResponse], error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}
	courses, total, err := uc.courseRepo.List(ctx, domain.ListTrainingCoursesFilter{
		CourseType: req.CourseType,
		IsActive:   req.IsActive,
		Q:          req.Q,
		Page:       req.Page,
		Size:       req.Size,
	})
	if err != nil {
		return nil, fmt.Errorf("ListTrainingCourses: %w", err)
	}
	items := make([]TrainingCourseResponse, 0, len(courses))
	for _, c := range courses {
		items = append(items, toCourseResponse(c))
	}
	result := pagination.NewOffsetResult(items, total, req.Page, req.Size)
	return &result, nil
}

func (uc *ProfessionalUseCase) CreateTrainingCourse(ctx context.Context, req CreateTrainingCourseRequest, callerID uuid.UUID, ip string) (*TrainingCourseResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	c, err := uc.courseRepo.Create(ctx, domain.CreateTrainingCourseParams{
		Code:        req.Code,
		Name:        req.Name,
		Provider:    req.Provider,
		Description: req.Description,
		CpeHours:    req.CpeHours,
		CourseType:  req.CourseType,
		IsActive:    isActive,
		Notes:       req.Notes,
		CreatedBy:   &callerID,
	})
	if err != nil {
		return nil, err
	}
	cid := c.ID
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "training_courses",
		ResourceID: &cid, Action: "TRAINING_COURSE_CREATED", NewValue: c, IPAddress: ip,
	})
	resp := toCourseResponse(c)
	return &resp, nil
}

func (uc *ProfessionalUseCase) GetTrainingCourse(ctx context.Context, id uuid.UUID) (*TrainingCourseResponse, error) {
	c, err := uc.courseRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toCourseResponse(c)
	return &resp, nil
}

func (uc *ProfessionalUseCase) UpdateTrainingCourse(ctx context.Context, id uuid.UUID, req UpdateTrainingCourseRequest, callerID uuid.UUID, ip string) (*TrainingCourseResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	c, err := uc.courseRepo.Update(ctx, domain.UpdateTrainingCourseParams{
		ID:          id,
		Name:        req.Name,
		Provider:    req.Provider,
		Description: req.Description,
		CpeHours:    req.CpeHours,
		CourseType:  req.CourseType,
		IsActive:    isActive,
		Notes:       req.Notes,
	})
	if err != nil {
		return nil, err
	}
	cid := c.ID
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "training_courses",
		ResourceID: &cid, Action: "TRAINING_COURSE_UPDATED", NewValue: c, IPAddress: ip,
	})
	resp := toCourseResponse(c)
	return &resp, nil
}

func (uc *ProfessionalUseCase) DeleteTrainingCourse(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) error {
	if err := uc.courseRepo.Delete(ctx, id); err != nil {
		return err
	}
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "training_courses",
		ResourceID: &id, Action: "TRAINING_COURSE_DELETED", IPAddress: ip,
	})
	return nil
}

// ─── Training record methods ──────────────────────────────────────────────────

func (uc *ProfessionalUseCase) ListTrainingRecords(ctx context.Context, employeeID uuid.UUID, callerID uuid.UUID, callerRoles []string) ([]TrainingRecordResponse, error) {
	ok, err := uc.canAccessEmployee(ctx, employeeID, callerID, callerRoles)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrInsufficientPermission
	}
	records, err := uc.recordRepo.ListByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("ListTrainingRecords: %w", err)
	}
	out := make([]TrainingRecordResponse, 0, len(records))
	for _, r := range records {
		out = append(out, toRecordResponse(r))
	}
	return out, nil
}

func (uc *ProfessionalUseCase) CreateTrainingRecord(ctx context.Context, employeeID uuid.UUID, req CreateTrainingRecordRequest, callerID uuid.UUID, ip string) (*TrainingRecordResponse, error) {
	completionDate, err := parseDate(req.CompletionDate)
	if err != nil {
		return nil, domain.ErrValidation
	}

	status := req.Status
	if status == "" {
		status = "ENROLLED"
	}

	rec, err := uc.recordRepo.Create(ctx, domain.CreateTrainingRecordParams{
		EmployeeID:     employeeID,
		CourseID:       req.CourseID,
		CompletionDate: completionDate,
		CpeHoursEarned: req.CpeHoursEarned,
		CertificateURL: req.CertificateURL,
		Status:         status,
		Notes:          req.Notes,
		CreatedBy:      &callerID,
	})
	if err != nil {
		return nil, fmt.Errorf("CreateTrainingRecord: %w", err)
	}

	rid := rec.ID
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "training_records",
		ResourceID: &rid, Action: "TRAINING_RECORD_CREATED", NewValue: rec, IPAddress: ip,
	})

	resp := toRecordResponse(rec)
	return &resp, nil
}

func (uc *ProfessionalUseCase) GetTrainingRecord(ctx context.Context, id uuid.UUID, callerID uuid.UUID, callerRoles []string) (*TrainingRecordResponse, error) {
	rec, err := uc.recordRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ok, err := uc.canAccessEmployee(ctx, rec.EmployeeID, callerID, callerRoles)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrInsufficientPermission
	}
	resp := toRecordResponse(rec)
	return &resp, nil
}

func (uc *ProfessionalUseCase) UpdateTrainingRecord(ctx context.Context, id uuid.UUID, req UpdateTrainingRecordRequest, callerID uuid.UUID, ip string) (*TrainingRecordResponse, error) {
	completionDate, err := parseDate(req.CompletionDate)
	if err != nil {
		return nil, domain.ErrValidation
	}

	rec, err := uc.recordRepo.Update(ctx, domain.UpdateTrainingRecordParams{
		ID:             id,
		CompletionDate: completionDate,
		CpeHoursEarned: req.CpeHoursEarned,
		CertificateURL: req.CertificateURL,
		Status:         req.Status,
		Notes:          req.Notes,
	})
	if err != nil {
		return nil, err
	}

	rid := rec.ID
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "training_records",
		ResourceID: &rid, Action: "TRAINING_RECORD_UPDATED", NewValue: rec, IPAddress: ip,
	})

	resp := toRecordResponse(rec)
	return &resp, nil
}

func (uc *ProfessionalUseCase) DeleteTrainingRecord(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) error {
	if err := uc.recordRepo.SoftDelete(ctx, id); err != nil {
		return err
	}
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "training_records",
		ResourceID: &id, Action: "TRAINING_RECORD_DELETED", IPAddress: ip,
	})
	return nil
}

func (uc *ProfessionalUseCase) GetCPESummary(ctx context.Context, employeeID uuid.UUID, year int, callerID uuid.UUID, callerRoles []string) (*CPESummaryResponse, error) {
	ok, err := uc.canAccessEmployee(ctx, employeeID, callerID, callerRoles)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrInsufficientPermission
	}
	if year <= 0 {
		year = time.Now().Year()
	}
	s, err := uc.recordRepo.GetCPESummary(ctx, employeeID, year)
	if err != nil {
		return nil, fmt.Errorf("GetCPESummary: %w", err)
	}
	return &CPESummaryResponse{
		EmployeeID:    s.EmployeeID,
		Year:          s.Year,
		TotalHours:    s.TotalHours,
		RequiredHours: s.RequiredHours,
		ByCategory:    json.RawMessage(s.ByCategory),
	}, nil
}

// ─── CPE requirement methods ───────────────────────────────────────────────────

func (uc *ProfessionalUseCase) ListCPERequirements(ctx context.Context, req ListCPERequirementsRequest) ([]CPERequirementResponse, error) {
	items, err := uc.cpeRepo.List(ctx, req.RoleCode, req.Year)
	if err != nil {
		return nil, fmt.Errorf("ListCPERequirements: %w", err)
	}
	out := make([]CPERequirementResponse, 0, len(items))
	for _, c := range items {
		out = append(out, toCPEResponse(c))
	}
	return out, nil
}

func (uc *ProfessionalUseCase) CreateCPERequirement(ctx context.Context, req CreateCPERequirementRequest, callerID uuid.UUID, ip string) (*CPERequirementResponse, error) {
	var breakdown []byte
	if len(req.CategoryBreakdown) > 0 {
		breakdown = []byte(req.CategoryBreakdown)
	}
	c, err := uc.cpeRepo.Create(ctx, domain.CreateCPERequirementParams{
		RoleCode:          req.RoleCode,
		Year:              req.Year,
		RequiredHours:     req.RequiredHours,
		CategoryBreakdown: breakdown,
		Notes:             req.Notes,
		CreatedBy:         &callerID,
	})
	if err != nil {
		return nil, err
	}
	cid := c.ID
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "cpe_requirements",
		ResourceID: &cid, Action: "CPE_REQUIREMENT_CREATED", NewValue: c, IPAddress: ip,
	})
	resp := toCPEResponse(c)
	return &resp, nil
}

func (uc *ProfessionalUseCase) UpdateCPERequirement(ctx context.Context, id uuid.UUID, req UpdateCPERequirementRequest, callerID uuid.UUID, ip string) (*CPERequirementResponse, error) {
	var breakdown []byte
	if len(req.CategoryBreakdown) > 0 {
		breakdown = []byte(req.CategoryBreakdown)
	}
	c, err := uc.cpeRepo.Update(ctx, domain.UpdateCPERequirementParams{
		ID:                id,
		RequiredHours:     req.RequiredHours,
		CategoryBreakdown: breakdown,
		Notes:             req.Notes,
	})
	if err != nil {
		return nil, err
	}
	cid := c.ID
	uc.auditLog.Log(ctx, audit.Entry{ //nolint:errcheck
		UserID: &callerID, Module: "hrm", Resource: "cpe_requirements",
		ResourceID: &cid, Action: "CPE_REQUIREMENT_UPDATED", NewValue: c, IPAddress: ip,
	})
	resp := toCPEResponse(c)
	return &resp, nil
}
