package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Poller reads PENDING outbox_messages and enqueues them as Asynq tasks.
// It uses SELECT FOR UPDATE SKIP LOCKED for safe concurrent polling.
type Poller struct {
	pool     *pgxpool.Pool
	client   *asynq.Client
	interval time.Duration
	batch    int
}

// NewPoller creates a Poller.
// interval is how often to poll; batch is the max rows per poll cycle.
func NewPoller(pool *pgxpool.Pool, client *asynq.Client, interval time.Duration, batch int) *Poller {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if batch <= 0 {
		batch = 50
	}
	return &Poller{pool: pool, client: client, interval: interval, batch: batch}
}

// Run blocks until ctx is cancelled, polling on each interval tick.
func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.poll(ctx); err != nil {
				// Non-fatal — log and continue. Caller can observe via structured logging.
				fmt.Printf("outbox poller error: %v\n", err)
			}
		}
	}
}

func (p *Poller) poll(ctx context.Context) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	rows, err := tx.Query(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, attempts
		FROM outbox_messages
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, p.batch)
	if err != nil {
		return err
	}

	var msgs []Message
	for rows.Next() {
		var m Message
		var aggID uuid.UUID
		var eventType string
		var raw json.RawMessage
		if err := rows.Scan(&m.ID, &m.AggregateType, &aggID, &eventType, &raw, &m.Attempts); err != nil {
			rows.Close()
			return err
		}
		m.AggregateID = aggID
		m.EventType = EventType(eventType)
		m.Payload = raw
		msgs = append(msgs, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for i := range msgs {
		m := &msgs[i]
		task := asynq.NewTask(string(m.EventType), m.Payload,
			asynq.TaskID(m.ID.String()),
			asynq.MaxRetry(5),
			asynq.Queue("events"),
		)
		if _, err := p.client.Enqueue(task); err != nil && !isAlreadyEnqueued(err) {
			// Mark failed so we don't loop forever on broken tasks.
			errMsg := err.Error()
			_, _ = tx.Exec(ctx, `
				UPDATE outbox_messages
				SET status = 'FAILED', attempts = attempts + 1, last_error = $1
				WHERE id = $2
			`, errMsg, m.ID)
			continue
		}
		_, _ = tx.Exec(ctx, `
			UPDATE outbox_messages
			SET status = 'PROCESSED', attempts = attempts + 1, processed_at = NOW()
			WHERE id = $1
		`, m.ID)
	}

	return tx.Commit(ctx)
}

// isAlreadyEnqueued returns true when Asynq reports task ID conflict (idempotent).
func isAlreadyEnqueued(err error) bool {
	return err != nil && err.Error() == asynq.ErrTaskIDConflict.Error()
}
