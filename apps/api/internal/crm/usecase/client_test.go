package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/crm/domain"
	"github.com/mdh/erp-audit/api/internal/crm/usecase"
)

// ── mock repo ─────────────────────────────────────────────────────────────────

type mockClientRepo struct {
	created    *domain.Client
	found      *domain.Client
	updated    *domain.Client
	createErr  error
	findErr    error
	updateErr  error
	deleteErr  error
	listItems  []*domain.Client
	listTotal  int64
}

func (m *mockClientRepo) Create(_ context.Context, _ domain.CreateClientParams) (*domain.Client, error) {
	return m.created, m.createErr
}
func (m *mockClientRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Client, error) {
	return m.found, m.findErr
}
func (m *mockClientRepo) Update(_ context.Context, _ domain.UpdateClientParams) (*domain.Client, error) {
	return m.updated, m.updateErr
}
func (m *mockClientRepo) SoftDelete(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return m.deleteErr
}
func (m *mockClientRepo) List(_ context.Context, _ domain.ListClientsFilter) ([]*domain.Client, int64, error) {
	return m.listItems, m.listTotal, m.findErr
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestClientUseCase_Create(t *testing.T) {
	t.Parallel()

	clientID := uuid.New()
	biz := "Acme Corp"
	eng := "Acme Corp EN"

	tests := []struct {
		name    string
		repo    *mockClientRepo
		req     usecase.ClientCreateRequest
		wantErr error
	}{
		{
			name: "success",
			repo: &mockClientRepo{
				created: &domain.Client{
					ID:           clientID,
					TaxCode:      "0123456789",
					BusinessName: biz,
					EnglishName:  &eng,
					Status:       domain.ClientStatusProspect,
				},
			},
			req: usecase.ClientCreateRequest{
				TaxCode:      "0123456789",
				BusinessName: biz,
				EnglishName:  &eng,
			},
		},
		{
			name:    "duplicate tax code → DUPLICATE_TAX_CODE",
			repo:    &mockClientRepo{createErr: domain.ErrDuplicateTaxCode},
			req:     usecase.ClientCreateRequest{TaxCode: "0123456789", BusinessName: biz},
			wantErr: domain.ErrDuplicateTaxCode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewClientUseCase(tt.repo, nil)
			resp, err := uc.Create(context.Background(), tt.req, nil, "")
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want err %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.TaxCode != tt.req.TaxCode {
				t.Errorf("want tax_code %q, got %q", tt.req.TaxCode, resp.TaxCode)
			}
		})
	}
}

func TestClientUseCase_GetByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name    string
		repo    *mockClientRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockClientRepo{
				found: &domain.Client{ID: id, TaxCode: "0123456789", BusinessName: "X", Status: domain.ClientStatusProspect},
			},
		},
		{
			name:    "not found → CLIENT_NOT_FOUND",
			repo:    &mockClientRepo{findErr: domain.ErrClientNotFound},
			wantErr: domain.ErrClientNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewClientUseCase(tt.repo, nil)
			_, err := uc.GetByID(context.Background(), id)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want err %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClientUseCase_Update(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name    string
		repo    *mockClientRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockClientRepo{
				updated: &domain.Client{ID: id, TaxCode: "0123456789", BusinessName: "Updated", Status: domain.ClientStatusProspect},
			},
		},
		{
			name:    "not found → CLIENT_NOT_FOUND",
			repo:    &mockClientRepo{updateErr: domain.ErrClientNotFound},
			wantErr: domain.ErrClientNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewClientUseCase(tt.repo, nil)
			_, err := uc.Update(context.Background(), id, usecase.ClientUpdateRequest{BusinessName: "Updated"}, nil, "")
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want err %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClientUseCase_Delete(t *testing.T) {
	t.Parallel()

	id := uuid.New()

	tests := []struct {
		name    string
		repo    *mockClientRepo
		wantErr error
	}{
		{name: "success", repo: &mockClientRepo{}},
		{
			name:    "not found → CLIENT_NOT_FOUND",
			repo:    &mockClientRepo{deleteErr: domain.ErrClientNotFound},
			wantErr: domain.ErrClientNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := usecase.NewClientUseCase(tt.repo, nil)
			err := uc.Delete(context.Background(), id, nil, "")
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("want err %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClientUseCase_Create_WithSalesFields(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	refID := uuid.New()
	clientID := uuid.New()

	repo := &mockClientRepo{
		created: &domain.Client{
			ID:           clientID,
			TaxCode:      "0123456789",
			BusinessName: "Acme",
			Status:       domain.ClientStatusProspect,
			SalesOwnerID: &ownerID,
			ReferrerID:   &refID,
		},
	}
	uc := usecase.NewClientUseCase(repo, nil)
	resp, err := uc.Create(context.Background(), usecase.ClientCreateRequest{
		TaxCode:      "0123456789",
		BusinessName: "Acme",
		SalesOwnerID: &ownerID,
		ReferrerID:   &refID,
	}, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SalesOwnerID == nil || *resp.SalesOwnerID != ownerID {
		t.Errorf("want sales_owner_id %s, got %v", ownerID, resp.SalesOwnerID)
	}
	if resp.ReferrerID == nil || *resp.ReferrerID != refID {
		t.Errorf("want referrer_id %s, got %v", refID, resp.ReferrerID)
	}
}

func TestClientUseCase_List(t *testing.T) {
	t.Parallel()

	items := []*domain.Client{
		{ID: uuid.New(), TaxCode: "0123456789", BusinessName: "A", Status: domain.ClientStatusProspect},
		{ID: uuid.New(), TaxCode: "0987654321", BusinessName: "B", Status: domain.ClientStatusAccepted},
	}

	uc := usecase.NewClientUseCase(&mockClientRepo{listItems: items, listTotal: 2}, nil)
	result, err := uc.List(context.Background(), usecase.ClientListRequest{Page: 1, Size: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Errorf("want total 2, got %d", result.Total)
	}
	if len(result.Data) != 2 {
		t.Errorf("want 2 items, got %d", len(result.Data))
	}
}
