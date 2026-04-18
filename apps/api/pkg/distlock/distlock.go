// Package distlock provides a simple Redis-backed distributed lock.
// Uses SET NX PX (single Redis node) — suitable for dev and single-instance prod.
package distlock

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ErrLockNotAcquired is returned when the lock is already held.
var ErrLockNotAcquired = errors.New("LOCK_NOT_ACQUIRED")

// Lock represents a held distributed lock.
type Lock struct {
	client *redis.Client
	key    string
	token  string
}

// Acquirer is the interface for acquiring distributed locks.
// Implementations may use Redis (production) or be stubbed (tests).
type Acquirer interface {
	Acquire(ctx context.Context, key string) (*Lock, error)
}

// Locker acquires and releases distributed locks backed by Redis.
type Locker struct {
	client *redis.Client
	ttl    time.Duration
}

// New creates a Locker with the given Redis client and lock TTL.
func New(client *redis.Client, ttl time.Duration) *Locker {
	return &Locker{client: client, ttl: ttl}
}

// Acquire tries to acquire the lock for key. Returns ErrLockNotAcquired if
// already held. The caller must call Release when done.
func (l *Locker) Acquire(ctx context.Context, key string) (*Lock, error) {
	token := uuid.New().String()
	ok, err := l.client.SetNX(ctx, key, token, l.ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLockNotAcquired
	}
	return &Lock{client: l.client, key: key, token: token}, nil
}

// Release releases the lock, but only if this holder still owns it.
// Uses a Lua script for atomic check-and-delete.
func (lk *Lock) Release(ctx context.Context) error {
	if lk.client == nil {
		return nil // no-op lock used in tests
	}
	const script = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end`
	return lk.client.Eval(ctx, script, []string{lk.key}, lk.token).Err()
}

// NewNoopLock returns a Lock that does nothing on Release.
// Useful in unit tests where Redis is not available.
func NewNoopLock() *Lock {
	return &Lock{}
}
