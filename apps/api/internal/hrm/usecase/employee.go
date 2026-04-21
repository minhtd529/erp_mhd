package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

const masked = "***"

// EmployeeListItem is the lean DTO for list endpoints.
type EmployeeListItem struct {
	ID             uuid.UUID  `json:"id"`
	EmployeeCode   *string    `json:"employee_code,omitempty"`
	FullName       string     `json:"full_name"`
	DisplayName    *string    `json:"display_name,omitempty"`
	Email          string     `json:"email"`
	Grade          string     `json:"grade"`
	Status         string     `json:"status"`
	BranchID       *uuid.UUID `json:"branch_id,omitempty"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty"`
	PositionTitle  *string    `json:"position_title,omitempty"`
	EmploymentType string     `json:"employment_type"`
	HiredDate      *string    `json:"hired_date,omitempty"`
	WorkLocation   string     `json:"work_location"`
	CreatedAt      string     `json:"created_at"`
}

// EmployeeResponse is the full DTO for detail endpoints (sensitive fields masked).
type EmployeeResponse struct {
	ID                  uuid.UUID  `json:"id"`
	EmployeeCode        *string    `json:"employee_code,omitempty"`
	FullName            string     `json:"full_name"`
	DisplayName         *string    `json:"display_name,omitempty"`
	Email               string     `json:"email"`
	Phone               *string    `json:"phone,omitempty"`
	DateOfBirth         *string    `json:"date_of_birth,omitempty"`
	Grade               string     `json:"grade"`
	Status              string     `json:"status"`
	ManagerID           *uuid.UUID `json:"manager_id,omitempty"`
	BranchID            *uuid.UUID `json:"branch_id,omitempty"`
	DepartmentID        *uuid.UUID `json:"department_id,omitempty"`
	PositionTitle       *string    `json:"position_title,omitempty"`
	EmploymentType      string     `json:"employment_type"`
	HiredDate           *string    `json:"hired_date,omitempty"`
	ProbationEndDate    *string    `json:"probation_end_date,omitempty"`
	TerminationDate     *string    `json:"termination_date,omitempty"`
	TerminationReason   *string    `json:"termination_reason,omitempty"`
	CurrentContractID   *uuid.UUID `json:"current_contract_id,omitempty"`
	Gender              *string    `json:"gender,omitempty"`
	PlaceOfBirth        *string    `json:"place_of_birth,omitempty"`
	Nationality         *string    `json:"nationality,omitempty"`
	Ethnicity           *string    `json:"ethnicity,omitempty"`
	PersonalEmail       *string    `json:"personal_email,omitempty"`
	PersonalPhone       *string    `json:"personal_phone,omitempty"`
	WorkPhone           *string    `json:"work_phone,omitempty"`
	CurrentAddress      *string    `json:"current_address,omitempty"`
	PermanentAddress    *string    `json:"permanent_address,omitempty"`
	CccdEncrypted       string     `json:"cccd_encrypted"`
	CccdIssuedDate      *string    `json:"cccd_issued_date,omitempty"`
	CccdIssuedPlace     *string    `json:"cccd_issued_place,omitempty"`
	PassportNumber      *string    `json:"passport_number,omitempty"`
	PassportExpiry      *string    `json:"passport_expiry,omitempty"`
	HiredSource         *string    `json:"hired_source,omitempty"`
	ReferrerEmployeeID  *uuid.UUID `json:"referrer_employee_id,omitempty"`
	ProbationSalaryPct  *float64   `json:"probation_salary_pct,omitempty"`
	WorkLocation        string     `json:"work_location"`
	RemoteDaysPerWeek   *int16     `json:"remote_days_per_week,omitempty"`
	EducationLevel      *string    `json:"education_level,omitempty"`
	EducationMajor      *string    `json:"education_major,omitempty"`
	EducationSchool     *string    `json:"education_school,omitempty"`
	EducationGraduationYear *int16 `json:"education_graduation_year,omitempty"`
	VnCpaNumber         *string    `json:"vn_cpa_number,omitempty"`
	VnCpaIssuedDate     *string    `json:"vn_cpa_issued_date,omitempty"`
	VnCpaExpiryDate     *string    `json:"vn_cpa_expiry_date,omitempty"`
	PracticingCertNumber *string   `json:"practicing_certificate_number,omitempty"`
	PracticingCertExpiry *string   `json:"practicing_certificate_expiry,omitempty"`
	BaseSalary          *float64   `json:"base_salary,omitempty"`
	SalaryCurrency      *string    `json:"salary_currency,omitempty"`
	SalaryEffectiveDate *string    `json:"salary_effective_date,omitempty"`
	BankAccountEncrypted string    `json:"bank_account_encrypted"`
	BankName            *string    `json:"bank_name,omitempty"`
	BankBranch          *string    `json:"bank_branch,omitempty"`
	MstCaNhanEncrypted  string     `json:"mst_ca_nhan_encrypted"`
	CommissionRate      *float64   `json:"commission_rate,omitempty"`
	CommissionType      string     `json:"commission_type"`
	SalesTargetYearly   *float64   `json:"sales_target_yearly,omitempty"`
	BizDevRegion        *string    `json:"biz_dev_region,omitempty"`
	SoBhxhEncrypted             string  `json:"so_bhxh_encrypted"`
	BhxhRegisteredDate          *string `json:"bhxh_registered_date,omitempty"`
	BhxhProvinceCode            *string `json:"bhxh_province_code,omitempty"`
	BhytCardNumber              *string `json:"bhyt_card_number,omitempty"`
	BhytExpiryDate              *string `json:"bhyt_expiry_date,omitempty"`
	BhytRegisteredHospitalCode  *string `json:"bhyt_registered_hospital_code,omitempty"`
	BhytRegisteredHospitalName  *string `json:"bhyt_registered_hospital_name,omitempty"`
	TncnRegistered              bool    `json:"tncn_registered"`
	CreatedAt           string     `json:"created_at"`
	UpdatedAt           string     `json:"updated_at"`
}

type CreateEmployeeRequest struct {
	FullName       string     `json:"full_name"       binding:"required,max=200"`
	Email          string     `json:"email"           binding:"required,email,max=255"`
	Phone          *string    `json:"phone"`
	DateOfBirth    *string    `json:"date_of_birth"`
	Grade          string     `json:"grade"           binding:"required,oneof=EXECUTIVE PARTNER DIRECTOR MANAGER SENIOR JUNIOR INTERN SUPPORT"`
	ManagerID      *uuid.UUID `json:"manager_id"`
	Status         string     `json:"status"          binding:"omitempty,oneof=ACTIVE INACTIVE ON_LEAVE TERMINATED"`
	BranchID       *uuid.UUID `json:"branch_id"`
	DepartmentID   *uuid.UUID `json:"department_id"`
	PositionTitle  *string    `json:"position_title"`
	EmploymentType string     `json:"employment_type" binding:"omitempty,oneof=FULL_TIME PART_TIME INTERN"`
	HiredDate      *string    `json:"hired_date"`
	DisplayName    *string    `json:"display_name"`
	Gender         *string    `json:"gender"          binding:"omitempty,oneof=MALE FEMALE OTHER"`
	PersonalEmail  *string    `json:"personal_email"`
	PersonalPhone  *string    `json:"personal_phone"`
	WorkLocation   string     `json:"work_location"   binding:"omitempty,oneof=OFFICE REMOTE HYBRID"`
	HiredSource    *string    `json:"hired_source"    binding:"omitempty,oneof=REFERRAL PORTAL DIRECT AGENCY"`
	EducationLevel *string    `json:"education_level" binding:"omitempty,oneof=BACHELOR MASTER PHD COLLEGE OTHER"`
	CommissionType string     `json:"commission_type" binding:"omitempty,oneof=FIXED TIERED NONE"`
}

type UpdateEmployeeRequest struct {
	FullName                *string    `json:"full_name"                binding:"omitempty,max=200"`
	Phone                   *string    `json:"phone"`
	Grade                   *string    `json:"grade"                    binding:"omitempty,oneof=EXECUTIVE PARTNER DIRECTOR MANAGER SENIOR JUNIOR INTERN SUPPORT"`
	ManagerID               *uuid.UUID `json:"manager_id"`
	Status                  *string    `json:"status"                   binding:"omitempty,oneof=ACTIVE INACTIVE ON_LEAVE TERMINATED"`
	BranchID                *uuid.UUID `json:"branch_id"`
	DepartmentID            *uuid.UUID `json:"department_id"`
	PositionTitle           *string    `json:"position_title"`
	EmploymentType          *string    `json:"employment_type"          binding:"omitempty,oneof=FULL_TIME PART_TIME INTERN"`
	HiredDate               *string    `json:"hired_date"`
	ProbationEndDate        *string    `json:"probation_end_date"`
	TerminationDate         *string    `json:"termination_date"`
	TerminationReason       *string    `json:"termination_reason"`
	DisplayName             *string    `json:"display_name"`
	Gender                  *string    `json:"gender"                   binding:"omitempty,oneof=MALE FEMALE OTHER"`
	PersonalEmail           *string    `json:"personal_email"`
	PersonalPhone           *string    `json:"personal_phone"`
	WorkPhone               *string    `json:"work_phone"`
	CurrentAddress          *string    `json:"current_address"`
	PermanentAddress        *string    `json:"permanent_address"`
	WorkLocation            *string    `json:"work_location"            binding:"omitempty,oneof=OFFICE REMOTE HYBRID"`
	RemoteDaysPerWeek       *int16     `json:"remote_days_per_week"`
	HiredSource             *string    `json:"hired_source"             binding:"omitempty,oneof=REFERRAL PORTAL DIRECT AGENCY"`
	EducationLevel          *string    `json:"education_level"          binding:"omitempty,oneof=BACHELOR MASTER PHD COLLEGE OTHER"`
	EducationMajor          *string    `json:"education_major"`
	EducationSchool         *string    `json:"education_school"`
	EducationGraduationYear *int16     `json:"education_graduation_year"`
	VnCpaNumber             *string    `json:"vn_cpa_number"`
	PracticingCertNumber    *string    `json:"practicing_certificate_number"`
	CommissionType          *string    `json:"commission_type"          binding:"omitempty,oneof=FIXED TIERED NONE"`
	CommissionRate          *float64   `json:"commission_rate"`
	BizDevRegion            *string    `json:"biz_dev_region"`
	Nationality             *string    `json:"nationality"`
	Ethnicity               *string    `json:"ethnicity"`
}

type ListEmployeeRequest struct {
	Page         int        `form:"page,default=1"  binding:"min=1"`
	Size         int        `form:"size,default=20" binding:"min=1,max=100"`
	BranchID     *uuid.UUID `form:"branch_id"`
	DepartmentID *uuid.UUID `form:"department_id"`
	Status       *string    `form:"status"`
	Grade        *string    `form:"grade"`
	Q            string     `form:"q"`
}

// ─── Scope helpers ────────────────────────────────────────────────────────────

type empScope int

const (
	empScopeAll    empScope = iota
	empScopeBranch          // HEAD_OF_BRANCH — own branch only
	empScopeSelf            // JUNIOR_AUDITOR, SENIOR_AUDITOR, ACCOUNTANT
)

func getEmpScope(roles []string) empScope {
	for _, r := range roles {
		switch r {
		case "SUPER_ADMIN", "CHAIRMAN", "CEO", "HR_MANAGER", "HR_STAFF", "PARTNER", "AUDIT_MANAGER":
			return empScopeAll
		case "HEAD_OF_BRANCH":
			return empScopeBranch
		}
	}
	return empScopeSelf
}

// ─── UseCase ─────────────────────────────────────────────────────────────────

type EmployeeUseCase struct {
	repo     domain.EmployeeRepository
	auditLog *audit.Logger
}

func NewEmployeeUseCase(repo domain.EmployeeRepository, auditLog *audit.Logger) *EmployeeUseCase {
	return &EmployeeUseCase{repo: repo, auditLog: auditLog}
}

func (uc *EmployeeUseCase) ListEmployees(
	ctx context.Context,
	req ListEmployeeRequest,
	callerUserID uuid.UUID,
	callerRoles []string,
	callerBranchID *uuid.UUID,
) (pagination.OffsetResult[EmployeeListItem], error) {
	f := domain.ListEmployeesFilter{
		Page:         req.Page,
		Size:         req.Size,
		BranchID:     req.BranchID,
		DepartmentID: req.DepartmentID,
		Status:       req.Status,
		Grade:        req.Grade,
		Q:            req.Q,
	}

	scope := getEmpScope(callerRoles)
	switch scope {
	case empScopeBranch:
		f.BranchScope = callerBranchID
	case empScopeSelf:
		f.UserID = &callerUserID
	}

	employees, total, err := uc.repo.List(ctx, f)
	if err != nil {
		return pagination.OffsetResult[EmployeeListItem]{}, fmt.Errorf("employee.List: %w", err)
	}
	items := make([]EmployeeListItem, len(employees))
	for i, e := range employees {
		items[i] = toEmployeeListItem(e)
	}
	return pagination.NewOffsetResult(items, total, req.Page, req.Size), nil
}

func (uc *EmployeeUseCase) GetEmployee(
	ctx context.Context,
	id uuid.UUID,
	callerUserID uuid.UUID,
	callerRoles []string,
	callerBranchID *uuid.UUID,
) (*EmployeeResponse, error) {
	e, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	scope := getEmpScope(callerRoles)
	switch scope {
	case empScopeBranch:
		if callerBranchID == nil || e.BranchID == nil || *callerBranchID != *e.BranchID {
			return nil, domain.ErrInsufficientPermission
		}
	case empScopeSelf:
		if e.UserID == nil || *e.UserID != callerUserID {
			return nil, domain.ErrInsufficientPermission
		}
	}

	resp := toEmployeeResponse(e)
	return &resp, nil
}

func (uc *EmployeeUseCase) CreateEmployee(
	ctx context.Context,
	req CreateEmployeeRequest,
	callerID *uuid.UUID,
	ip string,
) (*EmployeeResponse, error) {
	// Apply DB-backed defaults for optional fields
	if req.Status == "" {
		req.Status = "ACTIVE"
	}
	if req.EmploymentType == "" {
		req.EmploymentType = "FULL_TIME"
	}
	if req.WorkLocation == "" {
		req.WorkLocation = "OFFICE"
	}
	if req.CommissionType == "" {
		req.CommissionType = "NONE"
	}

	p := domain.CreateEmployeeParams{
		FullName:       req.FullName,
		Email:          req.Email,
		Phone:          req.Phone,
		Grade:          req.Grade,
		ManagerID:      req.ManagerID,
		Status:         req.Status,
		BranchID:       req.BranchID,
		DepartmentID:   req.DepartmentID,
		PositionTitle:  req.PositionTitle,
		EmploymentType: req.EmploymentType,
		DisplayName:    req.DisplayName,
		Gender:         req.Gender,
		PersonalEmail:  req.PersonalEmail,
		PersonalPhone:  req.PersonalPhone,
		WorkLocation:   req.WorkLocation,
		HiredSource:    req.HiredSource,
		EducationLevel: req.EducationLevel,
		CommissionType: req.CommissionType,
		CreatedBy:      callerID,
	}
	if req.DateOfBirth != nil {
		if t, err := time.Parse("2006-01-02", *req.DateOfBirth); err == nil {
			p.DateOfBirth = &t
		}
	}
	if req.HiredDate != nil {
		if t, err := time.Parse("2006-01-02", *req.HiredDate); err == nil {
			p.HiredDate = &t
		}
	}

	e, err := uc.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	eid := e.ID
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "employees",
			ResourceID: &eid, Action: "CREATE", IPAddress: ip,
		})
	}
	resp := toEmployeeResponse(e)
	return &resp, nil
}

func (uc *EmployeeUseCase) UpdateEmployee(
	ctx context.Context,
	id uuid.UUID,
	req UpdateEmployeeRequest,
	callerID *uuid.UUID,
	ip string,
) (*EmployeeResponse, error) {
	p := domain.UpdateEmployeeParams{
		ID:                      id,
		FullName:                req.FullName,
		Phone:                   req.Phone,
		Grade:                   req.Grade,
		ManagerID:               req.ManagerID,
		Status:                  req.Status,
		BranchID:                req.BranchID,
		DepartmentID:            req.DepartmentID,
		PositionTitle:           req.PositionTitle,
		EmploymentType:          req.EmploymentType,
		TerminationReason:       req.TerminationReason,
		DisplayName:             req.DisplayName,
		Gender:                  req.Gender,
		PersonalEmail:           req.PersonalEmail,
		PersonalPhone:           req.PersonalPhone,
		WorkPhone:               req.WorkPhone,
		CurrentAddress:          req.CurrentAddress,
		PermanentAddress:        req.PermanentAddress,
		WorkLocation:            req.WorkLocation,
		RemoteDaysPerWeek:       req.RemoteDaysPerWeek,
		HiredSource:             req.HiredSource,
		EducationLevel:          req.EducationLevel,
		EducationMajor:          req.EducationMajor,
		EducationSchool:         req.EducationSchool,
		EducationGraduationYear: req.EducationGraduationYear,
		VnCpaNumber:             req.VnCpaNumber,
		PracticingCertNumber:    req.PracticingCertNumber,
		CommissionType:          req.CommissionType,
		CommissionRate:          req.CommissionRate,
		BizDevRegion:            req.BizDevRegion,
		Nationality:             req.Nationality,
		Ethnicity:               req.Ethnicity,
		UpdatedBy:               callerID,
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
	p.HiredDate = parseDate(req.HiredDate)
	p.ProbationEndDate = parseDate(req.ProbationEndDate)
	p.TerminationDate = parseDate(req.TerminationDate)

	e, err := uc.repo.Update(ctx, p)
	if err != nil {
		return nil, err
	}

	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "employees",
			ResourceID: &id, Action: "UPDATE", IPAddress: ip,
		})
	}
	resp := toEmployeeResponse(e)
	return &resp, nil
}

func (uc *EmployeeUseCase) DeleteEmployee(ctx context.Context, id uuid.UUID, callerID *uuid.UUID, ip string) error {
	if err := uc.repo.SoftDelete(ctx, id, callerID); err != nil {
		return err
	}
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: callerID, Module: "hrm", Resource: "employees",
			ResourceID: &id, Action: "DELETE", IPAddress: ip,
		})
	}
	return nil
}

// ─── Converters ───────────────────────────────────────────────────────────────

func dateStr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func toEmployeeListItem(e *domain.Employee) EmployeeListItem {
	return EmployeeListItem{
		ID:             e.ID,
		EmployeeCode:   e.EmployeeCode,
		FullName:       e.FullName,
		DisplayName:    e.DisplayName,
		Email:          e.Email,
		Grade:          e.Grade,
		Status:         e.Status,
		BranchID:       e.BranchID,
		DepartmentID:   e.DepartmentID,
		PositionTitle:  e.PositionTitle,
		EmploymentType: e.EmploymentType,
		HiredDate:      dateStr(e.HiredDate),
		WorkLocation:   e.WorkLocation,
		CreatedAt:      e.CreatedAt.Format(time.RFC3339),
	}
}

func toEmployeeResponse(e *domain.Employee) EmployeeResponse {
	r := EmployeeResponse{
		ID:                  e.ID,
		EmployeeCode:        e.EmployeeCode,
		FullName:            e.FullName,
		DisplayName:         e.DisplayName,
		Email:               e.Email,
		Phone:               e.Phone,
		DateOfBirth:         dateStr(e.DateOfBirth),
		Grade:               e.Grade,
		Status:              e.Status,
		ManagerID:           e.ManagerID,
		BranchID:            e.BranchID,
		DepartmentID:        e.DepartmentID,
		PositionTitle:       e.PositionTitle,
		EmploymentType:      e.EmploymentType,
		HiredDate:           dateStr(e.HiredDate),
		ProbationEndDate:    dateStr(e.ProbationEndDate),
		TerminationDate:     dateStr(e.TerminationDate),
		TerminationReason:   e.TerminationReason,
		CurrentContractID:   e.CurrentContractID,
		Gender:              e.Gender,
		PlaceOfBirth:        e.PlaceOfBirth,
		Nationality:         e.Nationality,
		Ethnicity:           e.Ethnicity,
		PersonalEmail:       e.PersonalEmail,
		PersonalPhone:       e.PersonalPhone,
		WorkPhone:           e.WorkPhone,
		CurrentAddress:      e.CurrentAddress,
		PermanentAddress:    e.PermanentAddress,
		CccdEncrypted:       masked,
		CccdIssuedDate:      dateStr(e.CccdIssuedDate),
		CccdIssuedPlace:     e.CccdIssuedPlace,
		PassportNumber:      e.PassportNumber,
		PassportExpiry:      dateStr(e.PassportExpiry),
		HiredSource:         e.HiredSource,
		ReferrerEmployeeID:  e.ReferrerEmployeeID,
		ProbationSalaryPct:  e.ProbationSalaryPct,
		WorkLocation:        e.WorkLocation,
		RemoteDaysPerWeek:   e.RemoteDaysPerWeek,
		EducationLevel:      e.EducationLevel,
		EducationMajor:      e.EducationMajor,
		EducationSchool:     e.EducationSchool,
		EducationGraduationYear: e.EducationGraduationYear,
		VnCpaNumber:         e.VnCpaNumber,
		VnCpaIssuedDate:     dateStr(e.VnCpaIssuedDate),
		VnCpaExpiryDate:     dateStr(e.VnCpaExpiryDate),
		PracticingCertNumber: e.PracticingCertNumber,
		PracticingCertExpiry: dateStr(e.PracticingCertExpiry),
		BaseSalary:          e.BaseSalary,
		SalaryCurrency:      e.SalaryCurrency,
		SalaryEffectiveDate: dateStr(e.SalaryEffectiveDate),
		BankAccountEncrypted: masked,
		BankName:            e.BankName,
		BankBranch:          e.BankBranch,
		MstCaNhanEncrypted:  masked,
		CommissionRate:      e.CommissionRate,
		CommissionType:      e.CommissionType,
		SalesTargetYearly:   e.SalesTargetYearly,
		BizDevRegion:        e.BizDevRegion,
		SoBhxhEncrypted:             masked,
		BhxhRegisteredDate:          dateStr(e.BhxhRegisteredDate),
		BhxhProvinceCode:            e.BhxhProvinceCode,
		BhytCardNumber:              e.BhytCardNumber,
		BhytExpiryDate:              dateStr(e.BhytExpiryDate),
		BhytRegisteredHospitalCode:  e.BhytRegisteredHospitalCode,
		BhytRegisteredHospitalName:  e.BhytRegisteredHospitalName,
		TncnRegistered:              e.TncnRegistered,
		CreatedAt:           e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           e.UpdatedAt.Format(time.RFC3339),
	}
	return r
}
