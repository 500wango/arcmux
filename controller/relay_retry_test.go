package controller

import (
	"errors"
	"net/http/httptest"
	"testing"

	taskdto "github.com/500wango/arcmux/dto"
	"github.com/500wango/arcmux/relaykit/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestShouldRetryTreatsExhaustedUpstreamBalanceAsFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	err := types.NewOpenAIError(
		errors.New("Your credit balance is too low to access the API"),
		types.ErrorCodeBadResponseStatusCode,
		400,
	)

	require.True(t, shouldRetry(ctx, err, 1), "a 400 balance exhaustion must fail over when a retry remains")
	require.False(t, shouldRetry(ctx, err, 0), "retry configuration must still cap failover attempts")
}

func TestShouldRetryTaskRelayTreatsExhaustedUpstreamBalanceAsFailover(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	taskErr := &taskdto.TaskError{
		Code:       "insufficient_quota",
		Message:    "request rejected",
		StatusCode: 400,
	}

	require.True(t, shouldRetryTaskRelay(ctx, 0, taskErr, 1))
	taskErr.LocalError = true
	require.False(t, shouldRetryTaskRelay(ctx, 0, taskErr, 1))
}
