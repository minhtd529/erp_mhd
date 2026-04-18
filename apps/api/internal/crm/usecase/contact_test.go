package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
	"github.com/mdh/erp-audit/api/internal/crm/usecase"
)

// ── fake ContactRepository ────────────────────────────────────────────────────

type fakeContactRepo struct {
	contacts  map[uuid.UUID]*domain.ClientContact
	createErr error
	updateErr error
	deleteErr error
}

func newFakeContactRepo() *fakeContactRepo {
	return &fakeContactRepo{contacts: map[uuid.UUID]*domain.ClientContact{}}
}

func (r *fakeContactRepo) Create(_ context.Context, p domain.CreateContactParams) (*domain.ClientContact, error) {
	if r.createErr != nil {
		return nil, r.createErr
	}
	cc := &domain.ClientContact{
		ID:        uuid.New(),
		ClientID:  p.ClientID,
		FullName:  p.FullName,
		Title:     p.Title,
		Phone:     p.Phone,
		Email:     p.Email,
		IsPrimary: p.IsPrimary,
		CreatedBy: p.CreatedBy,
		UpdatedBy: p.CreatedBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	r.contacts[cc.ID] = cc
	return cc, nil
}

func (r *fakeContactRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.ClientContact, error) {
	cc, ok := r.contacts[id]
	if !ok {
		return nil, domain.ErrContactNotFound
	}
	return cc, nil
}

func (r *fakeContactRepo) Update(_ context.Context, p domain.UpdateContactParams) (*domain.ClientContact, error) {
	if r.updateErr != nil {
		return nil, r.updateErr
	}
	cc, ok := r.contacts[p.ID]
	if !ok {
		return nil, domain.ErrContactNotFound
	}
	cc.FullName = p.FullName
	cc.Title = p.Title
	cc.Phone = p.Phone
	cc.Email = p.Email
	cc.IsPrimary = p.IsPrimary
	cc.UpdatedBy = p.UpdatedBy
	cc.UpdatedAt = time.Now()
	return cc, nil
}

func (r *fakeContactRepo) SoftDelete(_ context.Context, id uuid.UUID, _ uuid.UUID, _ *uuid.UUID) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	if _, ok := r.contacts[id]; !ok {
		return domain.ErrContactNotFound
	}
	delete(r.contacts, id)
	return nil
}

func (r *fakeContactRepo) ListByClient(_ context.Context, clientID uuid.UUID) ([]*domain.ClientContact, error) {
	var result []*domain.ClientContact
	for _, cc := range r.contacts {
		if cc.ClientID == clientID {
			result = append(result, cc)
		}
	}
	return result, nil
}

// ── ContactUseCase tests ──────────────────────────────────────────────────────

func TestContact_Create_HappyPath(t *testing.T) {
	repo := newFakeContactRepo()
	uc := usecase.NewContactUseCase(repo, nil)

	clientID := uuid.New()
	caller := uuid.New()
	title := "CFO"

	resp, err := uc.Create(context.Background(), clientID, usecase.ContactCreateRequest{
		FullName:  "Jane Doe",
		Title:     &title,
		IsPrimary: true,
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if resp.FullName != "Jane Doe" {
		t.Errorf("want 'Jane Doe', got %q", resp.FullName)
	}
	if !resp.IsPrimary {
		t.Error("want IsPrimary=true")
	}
	if resp.ClientID != clientID {
		t.Error("wrong clientID")
	}
}

func TestContact_Create_RepoError(t *testing.T) {
	repo := newFakeContactRepo()
	repo.createErr = errors.New("DB_ERROR")
	uc := usecase.NewContactUseCase(repo, nil)

	_, err := uc.Create(context.Background(), uuid.New(), usecase.ContactCreateRequest{
		FullName: "X",
	}, uuid.New(), "127.0.0.1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestContact_ListByClient_HappyPath(t *testing.T) {
	repo := newFakeContactRepo()
	uc := usecase.NewContactUseCase(repo, nil)

	clientID := uuid.New()
	otherClientID := uuid.New()
	caller := uuid.New()

	// 2 contacts for target client, 1 for another
	for i := 0; i < 2; i++ {
		_, _ = uc.Create(context.Background(), clientID, usecase.ContactCreateRequest{FullName: "Contact"}, caller, "127.0.0.1")
	}
	_, _ = uc.Create(context.Background(), otherClientID, usecase.ContactCreateRequest{FullName: "Other"}, caller, "127.0.0.1")

	result, err := uc.ListByClient(context.Background(), clientID)
	if err != nil {
		t.Fatalf("ListByClient: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("want 2 contacts, got %d", len(result))
	}
}

func TestContact_ListByClient_Empty(t *testing.T) {
	repo := newFakeContactRepo()
	uc := usecase.NewContactUseCase(repo, nil)

	result, err := uc.ListByClient(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("want 0, got %d", len(result))
	}
}

func TestContact_Update_HappyPath(t *testing.T) {
	repo := newFakeContactRepo()
	uc := usecase.NewContactUseCase(repo, nil)

	clientID := uuid.New()
	caller := uuid.New()
	created, err := uc.Create(context.Background(), clientID, usecase.ContactCreateRequest{
		FullName: "Old Name",
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	newTitle := "CEO"
	updated, err := uc.Update(context.Background(), clientID, created.ID, usecase.ContactUpdateRequest{
		FullName: "New Name",
		Title:    &newTitle,
	}, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.FullName != "New Name" {
		t.Errorf("want 'New Name', got %q", updated.FullName)
	}
}

func TestContact_Update_NotFound(t *testing.T) {
	repo := newFakeContactRepo()
	uc := usecase.NewContactUseCase(repo, nil)

	_, err := uc.Update(context.Background(), uuid.New(), uuid.New(), usecase.ContactUpdateRequest{
		FullName: "X",
	}, uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrContactNotFound) {
		t.Errorf("want ErrContactNotFound, got %v", err)
	}
}

func TestContact_Delete_HappyPath(t *testing.T) {
	repo := newFakeContactRepo()
	uc := usecase.NewContactUseCase(repo, nil)

	clientID := uuid.New()
	caller := uuid.New()
	created, _ := uc.Create(context.Background(), clientID, usecase.ContactCreateRequest{
		FullName: "To Delete",
	}, caller, "127.0.0.1")

	err := uc.Delete(context.Background(), clientID, created.ID, caller, "127.0.0.1")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Should be gone from list
	contacts, _ := uc.ListByClient(context.Background(), clientID)
	if len(contacts) != 0 {
		t.Error("contact should be deleted")
	}
}

func TestContact_Delete_NotFound(t *testing.T) {
	repo := newFakeContactRepo()
	uc := usecase.NewContactUseCase(repo, nil)

	err := uc.Delete(context.Background(), uuid.New(), uuid.New(), uuid.New(), "127.0.0.1")
	if !errors.Is(err, domain.ErrContactNotFound) {
		t.Errorf("want ErrContactNotFound, got %v", err)
	}
}
