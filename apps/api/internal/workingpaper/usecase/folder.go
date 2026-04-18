package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// FolderUseCase handles working paper folder management.
type FolderUseCase struct {
	folderRepo domain.FolderRepository
	auditLog   *audit.Logger
}

// NewFolderUseCase constructs a FolderUseCase.
func NewFolderUseCase(folderRepo domain.FolderRepository, auditLog *audit.Logger) *FolderUseCase {
	return &FolderUseCase{folderRepo: folderRepo, auditLog: auditLog}
}

func (uc *FolderUseCase) Create(ctx context.Context, engagementID uuid.UUID, req FolderCreateRequest, callerID uuid.UUID, ip string) (*FolderResponse, error) {
	folder, err := uc.folderRepo.Create(ctx, domain.CreateFolderParams{
		EngagementID: engagementID,
		FolderName:   req.FolderName,
		CreatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "working_papers", Resource: "working_paper_folders",
		ResourceID: &folder.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toFolderResponse(folder)
	return &resp, nil
}

// FolderListRequest carries pagination params for folder listing.
type FolderListRequest struct {
	Page int
	Size int
}

func (uc *FolderUseCase) ListByEngagement(ctx context.Context, engagementID uuid.UUID, req FolderListRequest) (PaginatedResult[FolderResponse], error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	folders, err := uc.folderRepo.ListByEngagement(ctx, engagementID)
	if err != nil {
		return PaginatedResult[FolderResponse]{}, err
	}
	all := make([]FolderResponse, len(folders))
	for i, f := range folders {
		all[i] = toFolderResponse(f)
	}
	return paginateSlice(all, req.Page, req.Size), nil
}
