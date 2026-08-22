package service

import (
	"errors"
	"testing"

	"github.com/500wango/arcmux/relaykit/types"
	"github.com/stretchr/testify/require"
)

func TestIsChannelUnavailableErrorRecognizesExhaustedUpstreamBalance(t *testing.T) {
	testCases := []struct {
		name string
		err  *types.NewAPIError
	}{
		{
			name: "credit balance message on bad request",
			err:  types.NewOpenAIError(errors.New("Your credit balance is too low to access the API"), types.ErrorCodeBadResponseStatusCode, 400),
		},
		{
			name: "stable insufficient quota code",
			err: types.WithOpenAIError(types.OpenAIError{
				Message: "request rejected",
				Code:    "insufficient_quota",
			}, 400),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, IsChannelUnavailableError(tc.err))
		})
	}
}

func TestIsChannelUnavailableErrorDoesNotTurnClientErrorsIntoFailover(t *testing.T) {
	err := types.NewOpenAIError(errors.New("invalid model parameter"), types.ErrorCodeBadResponseStatusCode, 400)
	require.False(t, IsChannelUnavailableError(err))

	err = types.NewError(errors.New("provider-specific failure"), types.ErrorCodeChannelParamOverrideInvalid, types.ErrOptionWithSkipRetry())
	require.False(t, IsChannelUnavailableError(err))
}
