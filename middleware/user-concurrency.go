package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/logger"
	"github.com/500wango/arcmux/setting"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var userConcurrency = struct {
	sync.Mutex
	active map[int]int64
}{active: make(map[int]int64)}

// Each request owns a renewable lease so a crashed process cannot permanently
// consume a user's slots. Redis time keeps leases consistent across instances.
const userConcurrencyLeaseSeconds = 120

var acquireUserConcurrencyScript = redis.NewScript(`
local now = tonumber(redis.call('TIME')[1])
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', now)
if redis.call('ZCARD', KEYS[1]) >= tonumber(ARGV[1]) then return 0 end
redis.call('ZADD', KEYS[1], now + tonumber(ARGV[3]), ARGV[2])
redis.call('EXPIRE', KEYS[1], ARGV[3])
return 1
`)

var renewUserConcurrencyScript = redis.NewScript(`
local now = tonumber(redis.call('TIME')[1])
local expires = redis.call('ZSCORE', KEYS[1], ARGV[1])
if not expires or tonumber(expires) <= now then return 0 end
redis.call('ZADD', KEYS[1], 'XX', now + tonumber(ARGV[2]), ARGV[1])
redis.call('EXPIRE', KEYS[1], ARGV[2])
return 1
`)

// UserConcurrencyLimit holds one slot through the entire handler, including
// streaming and channel retries. It must run after authentication.
func UserConcurrencyLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := setting.UserConcurrencyLimit.Load()
		if limit == 0 {
			c.Next()
			return
		}
		userID := c.GetInt("id")
		if userID <= 0 {
			abortWithOpenAiMessage(c, http.StatusUnauthorized, "Authentication required")
			return
		}

		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()
		release, err := acquireUserConcurrency(ctx, cancel, userID, limit)
		if err != nil {
			logger.LogError(ctx, "user concurrency check failed: "+err.Error())
			abortWithOpenAiMessage(c, http.StatusServiceUnavailable, "Concurrency limit check failed", "concurrency_limit_check_failed")
			return
		}
		if release == nil {
			c.Header("Retry-After", "1")
			abortWithOpenAiMessage(c, http.StatusTooManyRequests,
				fmt.Sprintf("User concurrent request limit reached (%d). Please wait for an active request to finish.", limit), "user_concurrency_limit_exceeded")
			return
		}
		defer release()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func acquireUserConcurrency(ctx context.Context, cancel context.CancelFunc, userID int, limit int64) (func(), error) {
	if !common.RedisEnabled {
		userConcurrency.Lock()
		defer userConcurrency.Unlock()
		if userConcurrency.active[userID] >= limit {
			return nil, nil
		}
		userConcurrency.active[userID]++
		return func() {
			userConcurrency.Lock()
			defer userConcurrency.Unlock()
			userConcurrency.active[userID]--
			if userConcurrency.active[userID] == 0 {
				delete(userConcurrency.active, userID)
			}
		}, nil
	}

	rdb := common.RDB
	key := "concurrency:user:" + strconv.Itoa(userID)
	leaseID := common.GetUUID()
	checkCtx, checkCancel := context.WithTimeout(ctx, 3*time.Second)
	allowed, err := acquireUserConcurrencyScript.Run(checkCtx, rdb, []string{key}, limit, leaseID, userConcurrencyLeaseSeconds).Int()
	checkCancel()
	if err != nil || allowed == 0 {
		return nil, err
	}

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				renewCtx, renewCancel := context.WithTimeout(ctx, 3*time.Second)
				renewed, err := renewUserConcurrencyScript.Run(renewCtx, rdb, []string{key}, leaseID, userConcurrencyLeaseSeconds).Int()
				renewCancel()
				if err != nil || renewed == 0 {
					logger.LogError(ctx, fmt.Sprintf("user concurrency lease lost: renewed=%d, error=%v", renewed, err))
					cancel()
					return
				}
			}
		}
	}()
	return func() {
		close(done)
		// Request cancellation must not cancel cleanup of its slot.
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer releaseCancel()
		if err := rdb.ZRem(releaseCtx, key, leaseID).Err(); err != nil {
			logger.LogError(releaseCtx, "user concurrency release failed: "+err.Error())
		}
	}, nil
}
