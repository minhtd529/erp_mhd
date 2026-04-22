// Package repository provides the PostgreSQL implementation for notifications.
package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdh/erp-audit/api/internal/notification/domain"
)

type Repo struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

const cols = `id, user_id, type, title, body, data, source_ref, is_read, read_at, created_at`

func scan(row interface{ Scan(...any) error }) (*domain.Notification, error) {
	var n domain.Notification
	err := row.Scan(
		&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body,
		&n.Data, &n.SourceRef, &n.IsRead, &n.ReadAt, &n.CreatedAt,
	)
	return &n, err
}

// Insert writes a new notification. Silently ignores duplicates (same source_ref per user).
func (r *Repo) Insert(ctx context.Context, p domain.InsertParams) error {
	data := p.Data
	if data == nil {
		data = []byte("{}")
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notifications (user_id, type, title, body, data, source_ref)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, source_ref) WHERE source_ref != '' DO NOTHING`,
		p.UserID, p.Type, p.Title, p.Body, data, p.SourceRef,
	)
	if err != nil {
		return fmt.Errorf("notification.Insert: %w", err)
	}
	return nil
}

// ListByUserID returns paginated notifications for userID, newest first.
func (r *Repo) ListByUserID(ctx context.Context, userID uuid.UUID, page, size int) ([]*domain.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	var total int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("notification.ListByUserID count: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT `+cols+` FROM notifications
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, size, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("notification.ListByUserID: %w", err)
	}
	defer rows.Close()

	var out []*domain.Notification
	for rows.Next() {
		n, err := scan(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("notification.ListByUserID scan: %w", err)
		}
		out = append(out, n)
	}
	if out == nil {
		out = []*domain.Notification{}
	}
	return out, total, rows.Err()
}

// MarkRead sets is_read=true for a notification belonging to userID.
func (r *Repo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE notifications
		SET    is_read = true, read_at = now()
		WHERE  id = $1 AND user_id = $2 AND is_read = false`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("notification.MarkRead: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Check if it exists at all (vs just already read)
		var exists bool
		_ = r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1 AND user_id = $2)`,
			id, userID,
		).Scan(&exists)
		if !exists {
			return domain.ErrNotificationNotFound
		}
	}
	return nil
}

// ensure Repo implements domain.Repository at compile time.
var _ domain.Repository = (*Repo)(nil)
