package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchGroupDataPreservesHeartbeatHistoryAndLatestStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/status-page/arcmux":
			_, _ = w.Write([]byte(`{"publicGroupList":[{"name":"Services","monitorList":[{"id":3,"name":"ArcMux API"}]}]}`))
		case "/api/status-page/heartbeat/arcmux":
			_, _ = w.Write([]byte(`{"heartbeatList":{"3":[{"status":0,"time":"2026-09-13 08:30:00"},{"status":1,"time":"2026-09-13 08:28:00"},{"status":2,"time":"2026-09-13 08:29:00"}]},"uptimeList":{"3_24":0.99}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	result := fetchGroupData(context.Background(), server.Client(), map[string]interface{}{
		"url": server.URL, "slug": "arcmux", "categoryName": "Platform",
	})
	require.Len(t, result.Monitors, 1)
	monitor := result.Monitors[0]
	assert.Equal(t, 0, monitor.Status)
	assert.Equal(t, 0.99, monitor.Uptime)
	assert.Equal(t, []UptimeHeartbeat{
		{Status: 1, Time: "2026-09-13 08:28:00"},
		{Status: 2, Time: "2026-09-13 08:29:00"},
		{Status: 0, Time: "2026-09-13 08:30:00"},
	}, monitor.Heartbeats)
}
