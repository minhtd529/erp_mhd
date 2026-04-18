package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// TemplateUseCase handles audit template CRUD and application.
type TemplateUseCase struct {
	templateRepo domain.TemplateRepository
	wpRepo       domain.WorkingPaperRepository
	auditLog     *audit.Logger
}

// NewTemplateUseCase constructs a TemplateUseCase.
func NewTemplateUseCase(
	templateRepo domain.TemplateRepository,
	wpRepo domain.WorkingPaperRepository,
	auditLog *audit.Logger,
) *TemplateUseCase {
	return &TemplateUseCase{templateRepo: templateRepo, wpRepo: wpRepo, auditLog: auditLog}
}

func (uc *TemplateUseCase) Create(ctx context.Context, req TemplateCreateRequest, callerID uuid.UUID, ip string) (*TemplateResponse, error) {
	version := req.Version
	if version == "" {
		version = "1.0"
	}
	tmpl, err := uc.templateRepo.Create(ctx, domain.CreateTemplateParams{
		TemplateType: req.TemplateType,
		Title:        req.Title,
		Version:      version,
		Content:      req.Content,
		VSACompliant: req.VSACompliant,
		CreatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "audit_templates",
		ResourceID: &tmpl.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toTemplateResponse(tmpl)
	return &resp, nil
}

func (uc *TemplateUseCase) Update(ctx context.Context, id uuid.UUID, req TemplateUpdateRequest, callerID uuid.UUID, ip string) (*TemplateResponse, error) {
	tmpl, err := uc.templateRepo.Update(ctx, domain.UpdateTemplateParams{
		ID:           id,
		Title:        req.Title,
		Content:      req.Content,
		VSACompliant: req.VSACompliant,
		UpdatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "audit_templates",
		ResourceID: &id, Action: "UPDATE", IPAddress: ip,
	})

	resp := toTemplateResponse(tmpl)
	return &resp, nil
}

func (uc *TemplateUseCase) Retire(ctx context.Context, id uuid.UUID, callerID uuid.UUID, ip string) error {
	if err := uc.templateRepo.Retire(ctx, id, callerID); err != nil {
		return err
	}
	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "audit_templates",
		ResourceID: &id, Action: "DELETE", IPAddress: ip,
	})
	return nil
}

func (uc *TemplateUseCase) List(ctx context.Context, req TemplateListRequest) (*PaginatedResult[TemplateResponse], error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}
	templates, total, err := uc.templateRepo.List(ctx, req.ActiveOnly, req.Page, req.Size)
	if err != nil {
		return nil, err
	}
	data := make([]TemplateResponse, len(templates))
	for i, t := range templates {
		data[i] = toTemplateResponse(t)
	}
	result := pagination.NewOffsetResult(data, total, req.Page, req.Size)
	return &result, nil
}

// ApplyToEngagement creates one DRAFT working paper per template for the engagement.
func (uc *TemplateUseCase) ApplyToEngagement(ctx context.Context, templateID uuid.UUID, engagementID uuid.UUID, callerID uuid.UUID, ip string) (*ApplyTemplateResponse, error) {
	tmpl, err := uc.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	wp, err := uc.wpRepo.Create(ctx, domain.CreateWPParams{
		EngagementID: engagementID,
		DocumentType: domain.DocProcedures,
		Title:        tmpl.Title,
		CreatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_papers",
		ResourceID: &wp.ID, Action: "CREATE", IPAddress: ip,
		NewValue: map[string]string{"from_template": templateID.String()},
	})

	return &ApplyTemplateResponse{
		Created: 1,
		Papers:  []WPResponse{toWPResponse(wp)},
	}, nil
}
