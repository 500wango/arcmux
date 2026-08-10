package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMissingFrontendAssetSeparatesAssetsFromClientRoutes(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{path: "/static/js/old-chunk.js", expected: true},
		{path: "/static/css/missing.css", expected: true},
		{path: "/favicon.ico", expected: true},
		{path: "/logo.png", expected: true},
		{path: "/logo.svg", expected: true},
		{path: "/logo-icon.svg", expected: true},
		{path: "/logo-full.svg", expected: true},
		{path: "/channels", expected: false},
		{path: "/dashboard/overview", expected: false},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			assert.Equal(t, test.expected, isMissingFrontendAsset(test.path))
		})
	}
}
