// cmd/worker runs the Asynq background worker and outbox poller.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/mdh/erp-audit/api/pkg/config"
	"github.com/mdh/erp-audit/api/pkg/database"
	"github.com/mdh/erp-audit/api/pkg/outbox"
	"github.com/mdh/erp-audit/api/pkg/worker"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("worker: failed to load config: %v", err)
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatalf("worker: failed to connect to database: %v", err)
	}
	defer db.Close()

	redisOpts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		log.Fatalf("worker: failed to parse Redis URL: %v", err)
	}
	redisAddr := redisOpts.Addr

	// ── Asynq worker server ───────────────────────────────────────────────────
	srv := worker.New(worker.Config{
		RedisAddr:   redisAddr,
		Concurrency: 10,
		Queue:       "events",
	})
	srv.Register(outbox.EventTimesheetSubmitted, worker.HandleTimesheetSubmitted)
	srv.Register(outbox.EventTimesheetApproved, worker.HandleTimesheetApproved)
	srv.Register(outbox.EventTimesheetRejected, worker.HandleTimesheetRejected)
	srv.Register(outbox.EventTimesheetLocked, worker.HandleTimesheetLocked)
	srv.Register(outbox.EventEngagementActivated, worker.HandleEngagementActivated)

	// ── Outbox poller ─────────────────────────────────────────────────────────
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqClient.Close()

	poller := outbox.NewPoller(db.Pool, asynqClient, 5*time.Second, 50)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go poller.Run(ctx)

	log.Println("worker: starting asynq server")
	if err := srv.Run(); err != nil {
		log.Printf("worker: server error: %v", err)
	}

	<-ctx.Done()
	srv.Shutdown()
	log.Println("worker: shut down")
}
