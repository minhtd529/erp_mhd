package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/hrm/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// UpdateProfileRequest contains only the self-editable fields.
type UpdateProfileRequest struct {
	DisplayName      *string `json:"display_name"`
	PersonalPhone    *string `json:"personal_phone"`
	PersonalEmail    *string `json:"personal_email"`
	CurrentAddress   *string `json:"current_address"`
	PermanentAddress *string `json:"permanent_address"`
}

// ProfileUseCase handles /me/profile endpoints.
type ProfileUseCase struct {
	repo     domain.EmployeeRepository
	auditLog *audit.Logger
}

func NewProfileUseCase(repo domain.EmployeeRepository, auditLog *audit.Logger) *ProfileUseCase {
	return &ProfileUseCase{repo: repo, auditLog: auditLog}
}

func (uc *ProfileUseCase) GetMyProfile(ctx context.Context, userID uuid.UUID) (*EmployeeResponse, error) {
	e, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := toEmployeeResponse(e)
	return &resp, nil
}

func (uc *ProfileUseCase) UpdateMyProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest, ip string) (*EmployeeResponse, error) {
	callerID := userID
	p := domain.UpdateProfileParams{
		UserID:           userID,
		DisplayName:      req.DisplayName,
		PersonalPhone:    req.PersonalPhone,
		PersonalEmail:    req.PersonalEmail,
		CurrentAddress:   req.CurrentAddress,
		PermanentAddress: req.PermanentAddress,
		UpdatedBy:        &callerID,
	}
	e, err := uc.repo.UpdateProfile(ctx, p)
	if err != nil {
		return nil, err
	}
	eid := e.ID
	if uc.auditLog != nil {
		_, _ = uc.auditLog.Log(ctx, audit.Entry{
			UserID: &callerID, Module: "hrm", Resource: "employees",
			ResourceID: &eid, Action: "UPDATE_PROFILE", IPAddress: ip,
		})
	}
	resp := toEmployeeResponse(e)
	return &resp, nil
}
