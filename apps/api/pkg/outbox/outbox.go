// Package outbox implements the transactional outbox pattern.
// Writers call Publish (or PublishTx for same-transaction writes); a background
// Poller reads PENDING rows and dispatches them to Asynq.
package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventType identifies the domain event kind.
type EventType string

const (
	EventTimesheetSubmitted EventType = "TimesheetSubmitted"
	EventTimesheetApproved  EventType = "TimesheetApproved"
	EventTimesheetRejected  EventType = "TimesheetRejected"
	EventTimesheetLocked    EventType = "TimesheetLocked"

	EventEngagementActivated EventType = "EngagementActivated"
	EventEngagementCompleted EventType = "EngagementCompleted"
	EventEngagementSettled   EventType = "EngagementSettled"

	// Billing events consumed by the commission accrual engine.
	EventInvoiceIssued    EventType = "invoice.issued"
	EventInvoiceCancelled EventType = "invoice.cancelled"
	EventPaymentReceived  EventType = "payment.received"
	EventCreditNoteIssued EventType = "credit_note.issued"
)

// MessageStatus represents the processing state of an outbox row.
type MessageStatus string

const (
	StatusPending    MessageStatus = "PENDING"
	StatusProcessing MessageStatus = "PROCESSING"
	StatusProcessed  MessageStatus = "PROCESSED"
	StatusFailed     MessageStatus = "FAILED"
)

// Message is a single outbox row.
type Message struct {
	ID            uuid.UUID
	AggregateType string
	AggregateID   uuid.UUID
	EventType     EventType
	Payload       json.RawMessage
	Status        MessageStatus
	Attempts      int
	LastError     *string
	CreatedAt     time.Time
	ProcessedAt   *time.Time
}

const insertSQL = `
	INSERT INTO outbox_messages (aggregate_type, aggregate_id, event_type, payload)
	VALUES ($1, $2, $3, $4)`

// Publisher writes domain events to the outbox_messages table.
// Nil-safe: all methods are no-ops when p is nil (useful in unit tests).
type Publisher struct {
	pool *pgxpool.Pool
}

// New returns a Publisher backed by pool.
func New(pool *pgxpool.Pool) *Publisher {
	return &Publisher{pool: pool}
}

// Publish writes a message using the shared pool (outside any explicit transaction).
func (p *Publisher) Publish(ctx context.Context, aggregateType string, aggregateID uuid.UUID, eventType EventType, payload any) error {
	if p == nil {
		return nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = p.pool.Exec(ctx, insertSQL, aggregateType, aggregateID, string(eventType), raw)
	return err
}

// PublishTx writes a message within the provided transaction, giving atomicity
// with the surrounding domain mutation.
func (p *Publisher) PublishTx(ctx context.Context, tx pgx.Tx, aggregateType string, aggregateID uuid.UUID, eventType EventType, payload any) error {
	if p == nil {
		return nil
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, insertSQL, aggregateType, aggregateID, string(eventType), raw)
	return err
}
