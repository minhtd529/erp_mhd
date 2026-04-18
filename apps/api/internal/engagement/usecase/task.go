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

func (uc *TaskUseCase) List(ctx context.Context, engagementID uuid.UUID, phase domain.TaskPhase) ([]TaskResponse, error) {
	tasks, err := uc.taskRepo.ListByEngagement(ctx, engagementID, phase)
	if err != nil {
		return nil, err
	}
	out := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		out[i] = toTaskResponse(t)
	}
	return out, nil
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

	_ = uc.auditLog.Log(ctx, audit.Entry{
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

	_ = uc.auditLog.Log(ctx, audit.Entry{
		UserID: &callerID, Module: "engagement", Resource: "engagement_tasks",
		ResourceID: &taskID, Action: "UPDATE", IPAddress: ip,
	})

	resp := toTaskResponse(t)
	return &resp, nil
}
