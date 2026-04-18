package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/workingpaper/domain"
	"github.com/mdh/erp-audit/api/internal/workingpaper/usecase"
)

// ── fake FolderRepository ─────────────────────────────────────────────────────

type fakeFolderRepo struct {
	folders map[uuid.UUID]*domain.WorkingPaperFolder
	createErr error
}

func newFakeFolderRepo() *fakeFolderRepo {
	return &fakeFolderRepo{folders: map[uuid.UUID]*domain.WorkingPaperFolder{}}
}

func (r *fakeFolderRepo) Create(_ context.Context, p domain.CreateFolderParams) (*domain.WorkingPaperFolder, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	f := &domain.WorkingPaperFolder{
		ID:           uuid.New(),
		EngagementID: p.EngagementID,
		FolderName:   p.FolderName,
		CreatedBy:    p.CreatedBy,
		CreatedAt:    time.Now(),
	}
	r.folders[f.ID] = f
	return f, nil
}

func (r *fakeFolderRepo) ListByEngagement(_ context.Context, engagementID uuid.UUID) ([]*domain.WorkingPaperFolder, error) {
	var result []*domain.WorkingPaperFolder
	for _, f := range r.folders {
		if f.EngagementID == engagementID {
			result = append(result, f)
		}
	}
	return result, nil
}

// ── fake TemplateRepository ───────────────────────────────────────────────────

type fakeTemplateRepo struct {
	templates map[uuid.UUID]*domain.AuditTemplate
	createErr error
	updateErr error
	retireErr error
	findErr   error
}

func newFakeTemplateRepo() *fakeTemplateRepo {
	return &fakeTemplateRepo{templates: map[uuid.UUID]*domain.AuditTemplate{}}
}

func (r *fakeTemplateRepo) Create(_ context.Context, p domain.CreateTemplateParams) (*domain.AuditTemplate, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	t := &domain.AuditTemplate{
		ID:           uuid.New(),
		TemplateType: p.TemplateType,
		Title:        p.Title,
		Version:      p.Version,
		Content:      p.Content,
		VSACompliant: p.VSACompliant,
		IsActive:     true,
		CreatedBy:    p.CreatedBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	r.templates[t.ID] = t
	return t, nil
}

func (r *fakeTemplateRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.AuditTemplate, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	t, ok := r.templates[id]
	if !ok {
		return nil, domain.ErrTemplateNotFound
	}
	return t, nil
}

func (r *fakeTemplateRepo) Update(_ context.Context, p domain.UpdateTemplateParams) (*domain.AuditTemplate, error) {
	if r.updateErr != nil {
		return nil, r.updateErr
	}
	t, ok := r.templates[p.ID]
	if !ok {
		return nil, domain.ErrTemplateNotFound
	}
	t.Title = p.Title
	t.Content = p.Content
	t.VSACompliant = p.VSACompliant
	return t, nil
}

func (r *fakeTemplateRepo) Retire(_ context.Context, id uuid.UUID, _ uuid.UUID) error {
	if r.retireErr != nil {
		return r.retireErr
	}
	t, ok := r.templates[id]
	if !ok {
		return domain.ErrTemplateNotFound
	}
	t.IsActive = false
	return nil
}

func (r *fakeTemplateRepo) List(_ context.Context, activeOnly bool, _, _ int) ([]*domain.AuditTemplate, int64, error) {
	var result []*domain.AuditTemplate
	for _, t := range r.templates {
		if !activeOnly || t.IsActive {
			result = append(result, t)
		}
	}
	return result, int64(len(result)), nil
}

// ── FolderUseCase tests ───────────────────────────────────────────────────────

func TestFolder_Create_HappyPath(t *testing.T) {
	folderRepo := newFakeFolderRepo()
	uc := usecase.NewFolderUseCase(folderRepo, nil)

	caller := uuid.New()
	engID := uuid.New()
	resp, err := uc.Create(context.Background(), engID, usecase.FolderCreateRequest{
		FolderName: "Working Files",
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if resp.FolderName != "Working Files" {
		t.Errorf("want 'Working Files', got %q", resp.FolderName)
	}
	if resp.EngagementID != engID {
		t.Errorf("wrong engagement ID")
	}
}

func TestFolder_Create_RepoError(t *testing.T) {
	folderRepo := newFakeFolderRepo()
	folderRepo.createErr = errors.New("DB_ERROR")
	uc := usecase.NewFolderUseCase(folderRepo, nil)

	_, err := uc.Create(context.Background(), uuid.New(), usecase.FolderCreateRequest{
		FolderName: "X",
	}, uuid.New(), "127.0.0.1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFolder_ListByEngagement_HappyPath(t *testing.T) {
	folderRepo := newFakeFolderRepo()
	uc := usecase.NewFolderUseCase(folderRepo, nil)

	caller := uuid.New()
	engID := uuid.New()
	otherEngID := uuid.New()

	// Create 2 folders for target engagement, 1 for another
	_, _ = uc.Create(context.Background(), engID, usecase.FolderCreateRequest{FolderName: "A"}, caller, "127.0.0.1")
	_, _ = uc.Create(context.Background(), engID, usecase.FolderCreateRequest{FolderName: "B"}, caller, "127.0.0.1")
	_, _ = uc.Create(context.Background(), otherEngID, usecase.FolderCreateRequest{FolderName: "C"}, caller, "127.0.0.1")

	result, err := uc.ListByEngagement(context.Background(), engID, usecase.FolderListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("ListByEngagement: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("want 2 folders, got %d", len(result.Data))
	}
}

func TestFolder_ListByEngagement_Empty(t *testing.T) {
	folderRepo := newFakeFolderRepo()
	uc := usecase.NewFolderUseCase(folderRepo, nil)

	result, err := uc.ListByEngagement(context.Background(), uuid.New(), usecase.FolderListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("want 0 folders, got %d", len(result.Data))
	}
}

// ── TemplateUseCase tests ─────────────────────────────────────────────────────

func TestTemplate_Create_DefaultVersion(t *testing.T) {
	tmplRepo := newFakeTemplateRepo()
	wpRepo := newFakeWPRepo()
	uc := usecase.NewTemplateUseCase(tmplRepo, wpRepo, nil)

	caller := uuid.New()
	resp, err := uc.Create(context.Background(), usecase.TemplateCreateRequest{
		TemplateType: "AUDIT",
		Title:        "Standard Audit",
		Content:      json.RawMessage(`{"steps":["plan","execute","report"]}`),
		VSACompliant: true,
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if resp.Version != "1.0" {
		t.Errorf("want default version '1.0', got %q", resp.Version)
	}
	if resp.Title != "Standard Audit" {
		t.Errorf("wrong title: %q", resp.Title)
	}
}

func TestTemplate_Create_CustomVersion(t *testing.T) {
	tmplRepo := newFakeTemplateRepo()
	uc := usecase.NewTemplateUseCase(tmplRepo, newFakeWPRepo(), nil)

	resp, err := uc.Create(context.Background(), usecase.TemplateCreateRequest{
		TemplateType: "AUDIT",
		Title:        "v2 Template",
		Version:      "2.0",
		Content:      json.RawMessage(`{}`),
	}, uuid.New(), "127.0.0.1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if resp.Version != "2.0" {
		t.Errorf("want '2.0', got %q", resp.Version)
	}
}

func TestTemplate_Update_HappyPath(t *testing.T) {
	tmplRepo := newFakeTemplateRepo()
	uc := usecase.NewTemplateUseCase(tmplRepo, newFakeWPRepo(), nil)

	caller := uuid.New()
	created, _ := uc.Create(context.Background(), usecase.TemplateCreateRequest{
		TemplateType: "AUDIT",
		Title:        "Old Title",
		Content:      json.RawMessage(`{}`),
	}, caller, "127.0.0.1")

	updated, err := uc.Update(context.Background(), created.ID, usecase.TemplateUpdateRequest{
		Title:        "New Title",
		Content:      json.RawMessage(`{"updated":true}`),
		VSACompliant: true,
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "New Title" {
		t.Errorf("want 'New Title', got %q", updated.Title)
	}
}

func TestTemplate_Update_NotFound(t *testing.T) {
	tmplRepo := newFakeTemplateRepo()
	uc := usecase.NewTemplateUseCase(tmplRepo, newFakeWPRepo(), nil)

	_, err := uc.Update(context.Background(), uuid.New(), usecase.TemplateUpdateRequest{
		Title:   "X",
		Content: json.RawMessage(`{}`),
	}, uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrTemplateNotFound) {
		t.Errorf("want ErrTemplateNotFound, got %v", err)
	}
}

func TestTemplate_Retire_HappyPath(t *testing.T) {
	tmplRepo := newFakeTemplateRepo()
	uc := usecase.NewTemplateUseCase(tmplRepo, newFakeWPRepo(), nil)

	caller := uuid.New()
	created, _ := uc.Create(context.Background(), usecase.TemplateCreateRequest{
		TemplateType: "AUDIT", Title: "To Retire", Content: json.RawMessage(`{}`),
	}, caller, "127.0.0.1")

	if err := uc.Retire(context.Background(), created.ID, caller, "127.0.0.1"); err != nil {
		t.Fatalf("Retire: %v", err)
	}
	// Verify it's no longer active
	list, _ := uc.List(context.Background(), usecase.TemplateListRequest{ActiveOnly: true})
	for _, tmpl := range list.Data {
		if tmpl.ID == created.ID {
			t.Error("retired template still in active list")
		}
	}
}

func TestTemplate_Retire_NotFound(t *testing.T) {
	tmplRepo := newFakeTemplateRepo()
	uc := usecase.NewTemplateUseCase(tmplRepo, newFakeWPRepo(), nil)

	err := uc.Retire(context.Background(), uuid.New(), uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrTemplateNotFound) {
		t.Errorf("want ErrTemplateNotFound, got %v", err)
	}
}

func TestTemplate_List_ActiveOnly(t *testing.T) {
	tmplRepo := newFakeTemplateRepo()
	uc := usecase.NewTemplateUseCase(tmplRepo, newFakeWPRepo(), nil)

	caller := uuid.New()
	for i := 0; i < 3; i++ {
		uc.Create(context.Background(), usecase.TemplateCreateRequest{ //nolint:errcheck
			TemplateType: "AUDIT", Title: "T", Content: json.RawMessage(`{}`),
		}, caller, "127.0.0.1")
	}

	result, err := uc.List(context.Background(), usecase.TemplateListRequest{ActiveOnly: true, Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Data) != 3 {
		t.Errorf("want 3, got %d", len(result.Data))
	}
}

func TestTemplate_ApplyToEngagement_HappyPath(t *testing.T) {
	tmplRepo := newFakeTemplateRepo()
	wpRepo := newFakeWPRepo()
	uc := usecase.NewTemplateUseCase(tmplRepo, wpRepo, nil)

	caller := uuid.New()
	created, _ := uc.Create(context.Background(), usecase.TemplateCreateRequest{
		TemplateType: "AUDIT",
		Title:        "Apply Me",
		Content:      json.RawMessage(`{}`),
	}, caller, "127.0.0.1")

	engID := uuid.New()
	resp, err := uc.ApplyToEngagement(context.Background(), created.ID, engID, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("ApplyToEngagement: %v", err)
	}
	if resp.Created != 1 {
		t.Errorf("want 1 paper created, got %d", resp.Created)
	}
	if resp.Papers[0].EngagementID != engID {
		t.Error("paper has wrong engagement ID")
	}
}

func TestTemplate_ApplyToEngagement_NotFound(t *testing.T) {
	tmplRepo := newFakeTemplateRepo()
	uc := usecase.NewTemplateUseCase(tmplRepo, newFakeWPRepo(), nil)

	_, err := uc.ApplyToEngagement(context.Background(), uuid.New(), uuid.New(), uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrTemplateNotFound) {
		t.Errorf("want ErrTemplateNotFound, got %v", err)
	}
}
