// Package worker provides the Asynq-backed task worker server and handler registry.
package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/mdh/erp-audit/api/pkg/outbox"
)

// HandlerFunc processes a single Asynq task.
type HandlerFunc func(ctx context.Context, t *asynq.Task) error

// Server wraps an asynq.Server with a typed handler registry.
type Server struct {
	srv      *asynq.Server
	mux      *asynq.ServeMux
	handlers map[outbox.EventType]HandlerFunc
}

// Config holds the Asynq server configuration.
type Config struct {
	RedisAddr   string
	Concurrency int
	Queue       string
}

// New creates a Server. Concurrency defaults to 10; Queue defaults to "events".
func New(cfg Config) *Server {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 10
	}
	if cfg.Queue == "" {
		cfg.Queue = "events"
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues:      map[string]int{cfg.Queue: 1},
		},
	)
	return &Server{
		srv:      srv,
		mux:      asynq.NewServeMux(),
		handlers: make(map[outbox.EventType]HandlerFunc),
	}
}

// Register adds a handler for the given event type.
func (s *Server) Register(eventType outbox.EventType, fn HandlerFunc) {
	s.handlers[eventType] = fn
	s.mux.HandleFunc(string(eventType), asynq.HandlerFunc(fn))
}

// Run starts the worker and blocks until an error occurs or the server is stopped.
func (s *Server) Run() error {
	if len(s.handlers) == 0 {
		return fmt.Errorf("worker: no handlers registered")
	}
	return s.srv.Run(s.mux)
}

// Shutdown gracefully stops the worker.
func (s *Server) Shutdown() {
	s.srv.Shutdown()
}
