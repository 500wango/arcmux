package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/setting"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserConcurrencySharesSlotsAcrossKeysUntilStreamEnds(t *testing.T) {
	for _, storage := range []string{"memory", "redis"} {
		t.Run(storage, func(t *testing.T) {
			previousLimit, previousRedis := setting.UserConcurrencyLimit.Load(), common.RedisEnabled
			t.Cleanup(func() {
				setting.UserConcurrencyLimit.Store(previousLimit)
				common.RedisEnabled = previousRedis
			})
			common.RedisEnabled = false
			if storage == "redis" {
				useRateLimitMiniRedis(t)
			}
			setting.UserConcurrencyLimit.Store(1)
			started, finished := make(chan struct{}), make(chan struct{})
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("id", 42)
				if c.GetHeader("Authorization") == "other-user" {
					c.Set("id", 43)
				}
			}, UserConcurrencyLimit())
			router.GET("/responses", func(c *gin.Context) {
				if c.Query("stream") == "true" {
					c.Header("Content-Type", "text/event-stream")
					c.Writer.WriteHeader(http.StatusOK)
					c.Writer.Flush()
					close(started)
					<-c.Request.Context().Done()
					return
				}
				c.Status(http.StatusNoContent)
			})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			request := httptest.NewRequest(http.MethodGet, "/responses?stream=true", nil).WithContext(ctx)
			request.Header.Set("Authorization", "first-key")
			go func() {
				defer close(finished)
				router.ServeHTTP(httptest.NewRecorder(), request)
			}()
			<-started
			request = httptest.NewRequest(http.MethodGet, "/responses", nil)
			request.Header.Set("Authorization", "second-key")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assert.Equal(t, http.StatusTooManyRequests, response.Code)
			assert.Contains(t, response.Body.String(), "user_concurrency_limit_exceeded")
			request.Header.Set("Authorization", "other-user")
			response = httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assert.Equal(t, http.StatusNoContent, response.Code)
			cancel()
			<-finished
			request.Header.Set("Authorization", "second-key")
			response = httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assert.Equal(t, http.StatusNoContent, response.Code, "canceled streams must release their slots")
		})
	}
}

func TestUserConcurrencyRedisRecoversAbandonedLeaseAndFailsClosed(t *testing.T) {
	server, client := useRateLimitMiniRedis(t)
	ctx := context.Background()
	key := "concurrency:user:42"
	allowed, err := acquireUserConcurrencyScript.Run(ctx, client, []string{key}, 1, "abandoned", userConcurrencyLeaseSeconds).Int()
	require.NoError(t, err)
	require.Equal(t, 1, allowed)

	// Renew a live stream's lease before its original deadline, then simulate a crash.
	now := time.Now()
	server.SetTime(now.Add(90 * time.Second))
	renewed, err := renewUserConcurrencyScript.Run(ctx, client, []string{key}, "abandoned", userConcurrencyLeaseSeconds).Int()
	require.NoError(t, err)
	require.Equal(t, 1, renewed)
	server.SetTime(now.Add(150 * time.Second))
	allowed, err = acquireUserConcurrencyScript.Run(ctx, client, []string{key}, 1, "next", userConcurrencyLeaseSeconds).Int()
	require.NoError(t, err)
	assert.Zero(t, allowed, "a renewed stream still occupies a slot beyond the original deadline")
	server.SetTime(now.Add(220 * time.Second))
	allowed, err = acquireUserConcurrencyScript.Run(ctx, client, []string{key}, 1, "next", userConcurrencyLeaseSeconds).Int()
	require.NoError(t, err)
	assert.Equal(t, 1, allowed, "an abandoned lease must not permanently block the user")

	previousLimit := setting.UserConcurrencyLimit.Load()
	t.Cleanup(func() { setting.UserConcurrencyLimit.Store(previousLimit) })
	setting.UserConcurrencyLimit.Store(1)
	require.NoError(t, client.Close())
	router := gin.New()
	router.GET("/responses", func(c *gin.Context) { c.Set("id", 42) }, UserConcurrencyLimit(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/responses", nil))
	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	setting.UserConcurrencyLimit.Store(0)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/responses", nil))
	assert.Equal(t, http.StatusNoContent, response.Code, "zero disables the limit without requiring Redis")
}
