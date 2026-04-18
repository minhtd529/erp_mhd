package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/engagement/domain"
	"github.com/mdh/erp-audit/api/internal/engagement/usecase"
)

// ── mock TaskRepository ───────────────────────────────────────────────────────

type mockTaskRepo struct {
	created   *domain.EngagementTask
	updated   *domain.EngagementTask
	listed    []*domain.EngagementTask
	createErr error
	updateErr error
	listErr   error
}

func (m *mockTaskRepo) Create(_ context.Context, _ domain.CreateTaskParams) (*domain.EngagementTask, error) {
	return m.created, m.createErr
}
func (m *mockTaskRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.EngagementTask, error) {
	return nil, nil
}
func (m *mockTaskRepo) Update(_ context.Context, _ domain.UpdateTaskParams) (*domain.EngagementTask, error) {
	return m.updated, m.updateErr
}
func (m *mockTaskRepo) ListByEngagement(_ context.Context, _ uuid.UUID, _ domain.TaskPhase) ([]*domain.EngagementTask, error) {
	return m.listed, m.listErr
}

// ── TaskUseCase tests ─────────────────────────────────────────────────────────

func buildTaskUC(taskRepo *mockTaskRepo, engageRepo *mockEngagementRepo) *usecase.TaskUseCase {
	return usecase.NewTaskUseCase(taskRepo, engageRepo, nil)
}

func TestTask_List_HappyPath(t *testing.T) {
	engID := uuid.New()
	tasks := []*domain.EngagementTask{
		{ID: uuid.New(), EngagementID: engID, Phase: domain.PhasePlanning, Title: "Plan scope"},
		{ID: uuid.New(), EngagementID: engID, Phase: domain.PhasePlanning, Title: "Risk assessment"},
	}
	taskRepo := &mockTaskRepo{listed: tasks}
	uc := buildTaskUC(taskRepo, &mockEngagementRepo{})

	result, err := uc.List(context.Background(), engID, usecase.TaskListRequest{Phase: domain.PhasePlanning, Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("want 2 tasks, got %d", len(result.Data))
	}
}

func TestTask_List_RepoError(t *testing.T) {
	taskRepo := &mockTaskRepo{listErr: errors.New("DB_ERROR")}
	uc := buildTaskUC(taskRepo, &mockEngagementRepo{})

	_, err := uc.List(context.Background(), uuid.New(), usecase.TaskListRequest{Phase: domain.PhasePlanning, Page: 1, Size: 20})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTask_Create_HappyPath(t *testing.T) {
	engID := uuid.New()
	taskID := uuid.New()
	phase := domain.PhasePlanning
	engagement := &domain.Engagement{ID: engID, Status: domain.StatusActive}
	task := &domain.EngagementTask{ID: taskID, EngagementID: engID, Phase: phase, Title: "New task"}

	taskRepo := &mockTaskRepo{created: task}
	engRepo := &mockEngagementRepo{found: engagement}
	uc := buildTaskUC(taskRepo, engRepo)

	caller := uuid.New()
	resp, err := uc.Create(context.Background(), engID, usecase.TaskCreateRequest{
		Phase: phase,
		Title: "New task",
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if resp.Title != "New task" {
		t.Errorf("want 'New task', got %q", resp.Title)
	}
}

func TestTask_Create_EngagementNotFound(t *testing.T) {
	taskRepo := &mockTaskRepo{}
	engRepo := &mockEngagementRepo{findErr: domain.ErrEngagementNotFound}
	uc := buildTaskUC(taskRepo, engRepo)

	_, err := uc.Create(context.Background(), uuid.New(), usecase.TaskCreateRequest{
		Phase: domain.PhasePlanning, Title: "X",
	}, uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrEngagementNotFound) {
		t.Errorf("want ErrEngagementNotFound, got %v", err)
	}
}

func TestTask_Update_HappyPath(t *testing.T) {
	engID := uuid.New()
	taskID := uuid.New()
	updated := &domain.EngagementTask{
		ID:           taskID,
		EngagementID: engID,
		Phase:        domain.PhaseFieldwork,
		Title:        "Updated",
		Status:       domain.TaskCompleted,
		UpdatedAt:    time.Now(),
	}
	taskRepo := &mockTaskRepo{updated: updated}
	uc := buildTaskUC(taskRepo, &mockEngagementRepo{})

	caller := uuid.New()
	resp, err := uc.Update(context.Background(), engID, taskID, usecase.TaskUpdateRequest{
		Title:  "Updated",
		Status: domain.TaskCompleted,
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if resp.Status != domain.TaskCompleted {
		t.Errorf("want COMPLETED status, got %s", resp.Status)
	}
}

func TestTask_Update_RepoError(t *testing.T) {
	taskRepo := &mockTaskRepo{updateErr: domain.ErrTaskNotFound}
	uc := buildTaskUC(taskRepo, &mockEngagementRepo{})

	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), usecase.TaskUpdateRequest{
		Title: "X", Status: domain.TaskInProgress,
	}, uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrTaskNotFound) {
		t.Errorf("want ErrTaskNotFound, got %v", err)
	}
}

// ── TeamUseCase extra tests (List and Update) ─────────────────────────────────

func TestTeam_List_HappyPath(t *testing.T) {
	engID := uuid.New()
	members := []*domain.EngagementMember{
		{ID: uuid.New(), EngagementID: engID, StaffID: uuid.New(), Role: domain.RoleManager},
		{ID: uuid.New(), EngagementID: engID, StaffID: uuid.New(), Role: domain.RoleAuditor},
	}
	memberRepo := &mockMemberRepo{listItems: members}
	uc := usecase.NewTeamUseCase(memberRepo, &mockEngagementRepo{}, nil)

	result, err := uc.List(context.Background(), engID, usecase.MemberListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("want 2 members, got %d", len(result.Data))
	}
}

func TestTeam_List_RepoError(t *testing.T) {
	memberRepo := &mockMemberRepo{findErr: errors.New("DB_ERROR")}
	uc := usecase.NewTeamUseCase(memberRepo, &mockEngagementRepo{}, nil)

	_, err := uc.List(context.Background(), uuid.New(), usecase.MemberListRequest{Page: 1, Size: 20})
	if err == nil {
		t.Fatal("expected error from list")
	}
}

func TestTeam_Update_HappyPath(t *testing.T) {
	engID := uuid.New()
	memberID := uuid.New()
	rate := 500_000.0
	updated := &domain.EngagementMember{
		ID:                memberID,
		EngagementID:      engID,
		StaffID:           uuid.New(),
		Role:              domain.RoleManager,
		HourlyRate:        &rate,
		AllocationPercent: 80,
	}
	memberRepo := &mockMemberRepo{updated: updated}
	uc := usecase.NewTeamUseCase(memberRepo, &mockEngagementRepo{}, nil)

	resp, err := uc.Update(context.Background(), engID, memberID, usecase.MemberUpdateRequest{
		Role:              domain.RoleManager,
		HourlyRate:        &rate,
		AllocationPercent: 80,
	}, uuid.New(), "127.0.0.1")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if resp.AllocationPercent != 80 {
		t.Errorf("want 80%% allocation, got %d", resp.AllocationPercent)
	}
}

func TestTeam_Update_RepoError(t *testing.T) {
	memberRepo := &mockMemberRepo{updateErr: domain.ErrMemberNotFound}
	uc := usecase.NewTeamUseCase(memberRepo, &mockEngagementRepo{}, nil)

	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), usecase.MemberUpdateRequest{
		Role: domain.RoleAuditor, AllocationPercent: 50,
	}, uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrMemberNotFound) {
		t.Errorf("want ErrMemberNotFound, got %v", err)
	}
}

// ── Engagement Update/Delete extra tests ─────────────────────────────────────

func TestEngagement_Update_HappyPath(t *testing.T) {
	engID := uuid.New()
	clientID := uuid.New()
	existing := &domain.Engagement{
		ID: engID, ClientID: clientID,
		Status: domain.StatusDraft,
	}
	updated := &domain.Engagement{
		ID: engID, ClientID: clientID,
		ServiceType: domain.ServiceAudit, FeeType: domain.FeeFixed, FeeAmount: 99_000_000,
		Status: domain.StatusDraft,
	}
	engRepo := &mockEngagementRepo{found: existing, updated: updated}
	uc := usecase.NewEngagementUseCase(engRepo, nil, nil)

	resp, err := uc.Update(context.Background(), engID, usecase.EngagementUpdateRequest{
		ServiceType: domain.ServiceAudit,
		FeeType:     domain.FeeFixed,
		FeeAmount:   99_000_000,
	}, uuid.New(), "127.0.0.1")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if resp.FeeAmount != 99_000_000 {
		t.Errorf("fee not updated: got %v", resp.FeeAmount)
	}
}

func TestEngagement_Update_NotFound(t *testing.T) {
	engRepo := &mockEngagementRepo{updateErr: domain.ErrEngagementNotFound}
	uc := usecase.NewEngagementUseCase(engRepo, nil, nil)

	_, err := uc.Update(context.Background(), uuid.New(), usecase.EngagementUpdateRequest{
		ServiceType: domain.ServiceAudit, FeeType: domain.FeeFixed, FeeAmount: 1,
	}, uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrEngagementNotFound) {
		t.Errorf("want ErrEngagementNotFound, got %v", err)
	}
}

func TestEngagement_Delete_HappyPath(t *testing.T) {
	engID := uuid.New()
	engRepo := &mockEngagementRepo{deleteErr: nil}
	uc := usecase.NewEngagementUseCase(engRepo, nil, nil)

	err := uc.Delete(context.Background(), engID, uuid.New(), "127.0.0.1")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestEngagement_Delete_Error(t *testing.T) {
	engRepo := &mockEngagementRepo{deleteErr: domain.ErrEngagementNotFound}
	uc := usecase.NewEngagementUseCase(engRepo, nil, nil)

	err := uc.Delete(context.Background(), uuid.New(), uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrEngagementNotFound) {
		t.Errorf("want ErrEngagementNotFound, got %v", err)
	}
}

// ── CostUseCase extra tests (List and Create) ─────────────────────────────────

func TestCost_List_HappyPath(t *testing.T) {
	engID := uuid.New()
	costs := []*domain.DirectCost{
		{ID: uuid.New(), EngagementID: engID, CostType: domain.CostTravel, Amount: 500_000, Status: domain.CostDraft},
		{ID: uuid.New(), EngagementID: engID, CostType: domain.CostOther, Amount: 200_000, Status: domain.CostDraft},
	}
	costRepo := &mockCostRepo{listItems: costs}
	uc := usecase.NewCostUseCase(costRepo, &mockEngagementRepo{}, nil)

	result, err := uc.List(context.Background(), engID, usecase.CostListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("want 2 costs, got %d", len(result.Data))
	}
}

func TestCost_List_RepoError(t *testing.T) {
	costRepo := &mockCostRepo{findErr: errors.New("DB_ERROR")}
	uc := usecase.NewCostUseCase(costRepo, &mockEngagementRepo{}, nil)

	_, err := uc.List(context.Background(), uuid.New(), usecase.CostListRequest{Page: 1, Size: 20})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCost_Create_HappyPath(t *testing.T) {
	engID := uuid.New()
	engagement := &domain.Engagement{ID: engID, Status: domain.StatusActive}
	cost := &domain.DirectCost{
		ID:           uuid.New(),
		EngagementID: engID,
		CostType:     domain.CostTravel,
		Description:  "Business trip",
		Amount:       1_500_000,
		Status:       domain.CostDraft,
	}
	costRepo := &mockCostRepo{created: cost}
	engRepo := &mockEngagementRepo{found: engagement}
	uc := usecase.NewCostUseCase(costRepo, engRepo, nil)

	resp, err := uc.Create(context.Background(), engID, usecase.CostCreateRequest{
		CostType:    domain.CostTravel,
		Description: "Business trip",
		Amount:      1_500_000,
	}, uuid.New(), "127.0.0.1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if resp.Amount != 1_500_000 {
		t.Errorf("wrong amount: %v", resp.Amount)
	}
}

func TestCost_Create_EngagementNotFound(t *testing.T) {
	costRepo := &mockCostRepo{}
	engRepo := &mockEngagementRepo{findErr: domain.ErrEngagementNotFound}
	uc := usecase.NewCostUseCase(costRepo, engRepo, nil)

	_, err := uc.Create(context.Background(), uuid.New(), usecase.CostCreateRequest{
		CostType: domain.CostTravel, Description: "X", Amount: 1,
	}, uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrEngagementNotFound) {
		t.Errorf("want ErrEngagementNotFound, got %v", err)
	}
}

func TestCost_Reject_HappyPath(t *testing.T) {
	reason := "Expense not approved"
	engID := uuid.New()
	cost := &domain.DirectCost{
		ID:           uuid.New(),
		EngagementID: engID,
		Status:       domain.CostSubmitted,
	}
	costRepo := &mockCostRepo{found: cost, updated: &domain.DirectCost{
		ID:           cost.ID,
		EngagementID: engID,
		Status:       domain.CostRejected,
		RejectReason: &reason,
	}}
	uc := usecase.NewCostUseCase(costRepo, &mockEngagementRepo{}, nil)

	resp, err := uc.Reject(context.Background(), engID, cost.ID, reason, uuid.New(), "127.0.0.1")
	if err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if resp.Status != domain.CostRejected {
		t.Errorf("want REJECTED, got %s", resp.Status)
	}
}
