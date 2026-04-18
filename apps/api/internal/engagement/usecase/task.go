package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/pkg/audit"
)

// TaskUseCase manages engagement tasks.
type TaskUseCase struct {
	taskRepo       domain.TaskRepository
	engagementRepo domain.EngagementRepository
	auditLog       *audit.Logger
}

// NewTaskUseCase constructs a TaskUseCase.
func NewTaskUseCase(taskRepo domain.TaskRepository, engagementRepo domain.EngagementRepository, auditLog *audit.Logger) *TaskUseCase {
	return &TaskUseCase{taskRepo: taskRepo, engagementRepo: engagementRepo, auditLog: auditLog}
}

// TaskListRequest carries filter and pagination params for task listing.
type TaskListRequest struct {
	Phase domain.TaskPhase
	Page  int
	Size  int
}

func (uc *TaskUseCase) List(ctx context.Context, engagementID uuid.UUID, req TaskListRequest) (PaginatedResult[TaskResponse], error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	tasks, err := uc.taskRepo.ListByEngagement(ctx, engagementID, req.Phase)
	if err != nil {
		return PaginatedResult[TaskResponse]{}, err
	}
	all := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		all[i] = toTaskResponse(t)
	}
	page, size := req.Page, req.Size
	start := (page - 1) * size
	if start > len(all) {
		start = len(all)
	}
	end := start + size
	if end > len(all) {
		end = len(all)
	}
	return newPaginatedResult(all[start:end], int64(len(all)), page, size), nil
}

func (uc *TaskUseCase) Create(ctx context.Context, engagementID uuid.UUID, req TaskCreateRequest, callerID uuid.UUID, ip string) (*TaskResponse, error) {
	if _, err := uc.engagementRepo.FindByID(ctx, engagementID); err != nil {
		return nil, err
	}

	t, err := uc.taskRepo.Create(ctx, domain.CreateTaskParams{
		EngagementID: engagementID,
		Phase:        req.Phase,
		Title:        req.Title,
		AssignedTo:   req.AssignedTo,
		DueDate:      req.DueDate,
		CreatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagement_tasks",
		ResourceID: &t.ID, Action: "CREATE", IPAddress: ip,
	})

	resp := toTaskResponse(t)
	return &resp, nil
}

func (uc *TaskUseCase) Update(ctx context.Context, engagementID uuid.UUID, taskID uuid.UUID, req TaskUpdateRequest, callerID uuid.UUID, ip string) (*TaskResponse, error) {
	t, err := uc.taskRepo.Update(ctx, domain.UpdateTaskParams{
		ID:           taskID,
		EngagementID: engagementID,
		Title:        req.Title,
		AssignedTo:   req.AssignedTo,
		Status:       req.Status,
		DueDate:      req.DueDate,
		UpdatedBy:    callerID,
	})
	if err != nil {
		return nil, err
	}

	_, _ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagement_tasks",
		ResourceID: &taskID, Action: "UPDATE", IPAddress: ip,
	})

	resp := toTaskResponse(t)
	return &resp, nil
}
