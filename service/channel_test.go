package service

import (
	"errors"
	"testing"

	"github.com/500wango/arcmux/relaykit/types"
	"github.com/500wango/arcmux/setting/operation_setting"
	"github.com/stretchr/testify/assert"
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

func TestIsChannelUnavailableMessageMatchesStableSignals(t *testing.T) {
	testCases := []struct {
		name      string
		message   string
		errorCode string
		want      bool
	}{
		{name: "error code is normalized", message: "request rejected", errorCode: "  Insufficient_Quota ", want: true},
		{name: "message fragment match", message: "You have insufficient quota for this model", errorCode: "", want: true},
		{name: "plain client error", message: "invalid model parameter", errorCode: "", want: false},
		{name: "empty inputs", message: "", errorCode: "", want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsChannelUnavailableMessage(tc.message, tc.errorCode))
		})
	}
}

// The configured automatic-disable keywords double as a failover signal even
// when auto-disable is off; matching must come from the keyword list itself,
// not the built-in balance fragments.
func TestIsChannelUnavailableMessageUsesConfiguredDisableKeywords(t *testing.T) {
	originalKeywords := operation_setting.AutomaticDisableKeywords
	operation_setting.AutomaticDisableKeywords = []string{"organization suspended"}
	t.Cleanup(func() { operation_setting.AutomaticDisableKeywords = originalKeywords })

	assert.True(t, IsChannelUnavailableMessage("request failed: organization suspended for abuse", ""))
	assert.False(t, IsChannelUnavailableMessage("upstream returned an unrelated failure", ""))
}
