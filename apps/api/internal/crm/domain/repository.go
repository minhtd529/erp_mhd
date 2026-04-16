package domain

import (
	"context"

	"github.com/google/uuid"
)

// ClientRepository defines the data-access contract for Client.
type ClientRepository interface {
	Create(ctx context.Context, p CreateClientParams) (*Client, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Client, error)
	Update(ctx context.Context, p UpdateClientParams) (*Client, error)
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	List(ctx context.Context, f ListClientsFilter) ([]*Client, int64, error)
}
