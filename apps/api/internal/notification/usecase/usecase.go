// Package usecase implements the notification use cases.
package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mdh/erp-audit/api/internal/notification/domain"
	"github.com/mdh/erp-audit/api/pkg/pagination"
)

// UseCase coordinates notification business logic.
type UseCase struct{ repo domain.Repository }

func New(repo domain.Repository) *UseCase { return &UseCase{repo: repo} }

// List returns paginated notifications for the authenticated user.
func (uc *UseCase) List(ctx context.Context, userID uuid.UUID, page, size int) (pagination.OffsetResult[*domain.Notification], error) {
	items, total, err := uc.repo.ListByUserID(ctx, userID, page, size)
	if err != nil {
		return pagination.OffsetResult[*domain.Notification]{}, err
	}
	return pagination.NewOffsetResult(items, total, page, size), nil
}

// MarkRead marks a single notification as read.
func (uc *UseCase) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return uc.repo.MarkRead(ctx, id, userID)
}
