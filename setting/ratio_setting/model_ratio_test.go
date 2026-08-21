package ratio_setting

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultModelRatioIncludesCurrentGPTModels(t *testing.T) {
	InitRatioSettings()

	tests := []struct {
		model string
		ratio float64
	}{
		{model: "gpt-5.2", ratio: 0.875},
		{model: "gpt-5.4", ratio: 1.25},
		{model: "gpt-5.4-mini", ratio: 0.375},
		{model: "gpt-5.5", ratio: 2.5},
		{model: "gpt-5.6-terra", ratio: 1.25},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			ratio, ok, matchedName := GetModelRatio(tt.model)
			require.True(t, ok)
			require.Equal(t, tt.model, matchedName)
			require.Equal(t, tt.ratio, ratio)
		})
	}
}

func TestCurrentGPTCompletionRatios(t *testing.T) {
	tests := []struct {
		model string
		ratio float64
	}{
		{model: "gpt-5.2", ratio: 8},
		{model: "gpt-5.4", ratio: 6},
		{model: "gpt-5.4-mini", ratio: 6},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			require.Equal(t, tt.ratio, GetCompletionRatio(tt.model))
		})
	}
}
