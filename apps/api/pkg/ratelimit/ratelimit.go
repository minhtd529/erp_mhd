// Package ratelimit provides Redis-backed sliding-window rate limiting middleware.
// Two limits are enforced per request:
//   - Per authenticated user: 100 req/min (keyed by user ID from JWT claims)
//   - Per IP address:         1000 req/min (keyed by client IP)
//
// The sliding window is implemented with a single Redis INCR + EXPIRE call:
//   - First request in a window sets the counter and TTL atomically via a Lua script.
//   - Subsequent requests increment the counter; TTL is preserved.
//   - When a limit is exceeded the handler aborts with HTTP 429 and Retry-After header.
package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	UserLimitPerMin = 100
	IPLimitPerMin   = 1000
	window          = time.Minute
)

// luaIncr atomically increments a counter and sets TTL on first call.
// Returns the new counter value.
const luaIncr = `
local v = redis.call("INCR", KEYS[1])
if v == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return v`

// RateLimiter holds the Redis client and limit thresholds.
type RateLimiter struct {
	rdb          *redis.Client
	userLimit    int
	ipLimit      int
	windowMillis int64
}

// New creates a RateLimiter with default thresholds.
func New(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{
		rdb:          rdb,
		userLimit:    UserLimitPerMin,
		ipLimit:      IPLimitPerMin,
		windowMillis: window.Milliseconds(),
	}
}

// Middleware returns a gin.HandlerFunc that enforces per-user and per-IP limits.
// AuthMiddleware must run before this middleware so user ID is available in context.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Per-IP limit (always enforced)
		ip := c.ClientIP()
		ipKey := "rl:ip:" + ip
		if exceeded, retryAfter := rl.check(ctx, ipKey, rl.ipLimit); exceeded {
			c.Header("Retry-After", strconv.FormatFloat(retryAfter.Seconds(), 'f', 0, 64))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "RATE_LIMIT_EXCEEDED",
				"message": fmt.Sprintf("Too many requests from IP %s; retry after %ds", ip, int(retryAfter.Seconds())),
			})
			return
		}

		// Per-user limit (only when authenticated)
		if userID, exists := c.Get("user_id"); exists && userID != nil {
			userKey := fmt.Sprintf("rl:user:%v", userID)
			if exceeded, retryAfter := rl.check(ctx, userKey, rl.userLimit); exceeded {
				c.Header("Retry-After", strconv.FormatFloat(retryAfter.Seconds(), 'f', 0, 64))
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error":   "RATE_LIMIT_EXCEEDED",
					"message": fmt.Sprintf("Too many requests; retry after %ds", int(retryAfter.Seconds())),
				})
				return
			}
		}

		c.Next()
	}
}

// check increments the counter for key and returns (exceeded, retryAfter).
func (rl *RateLimiter) check(ctx context.Context, key string, limit int) (bool, time.Duration) {
	if rl.rdb == nil {
		return false, 0
	}
	result, err := rl.rdb.Eval(ctx, luaIncr, []string{key}, rl.windowMillis).Int64()
	if err != nil {
		// On Redis errors, allow the request through (fail-open).
		return false, 0
	}
	if int(result) > limit {
		ttl, err := rl.rdb.PTTL(ctx, key).Result()
		if err != nil || ttl <= 0 {
			ttl = window
		}
		return true, ttl
	}
	return false, 0
}
