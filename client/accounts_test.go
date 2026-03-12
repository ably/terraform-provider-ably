package ably

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// GetAccountStats
// ---------------------------------------------------------------------------

func TestGetAccountStats_Success_WithParams(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("GET /accounts/acc-123/stats", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		assert.Equal(t, "hour", r.URL.Query().Get("unit"))
		assert.Equal(t, "10", r.URL.Query().Get("limit"))

		writeJSON(w, http.StatusOK, []StatsResponse{
			{
				IntervalID: "2026-03-05:00:00",
				Unit:       "hour",
				Schema:     "https://schemas.ably.com/json/app-stats-0.0.1.json",
				Entries:    map[string]interface{}{"messages.all.all.count": float64(100)},
				AccountID:  "acc-123",
			},
		})
	})

	stats, err := client.GetAccountStats(context.Background(), "acc-123", &StatsParams{
		Unit:  "hour",
		Limit: ptr(10),
	})
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, "2026-03-05:00:00", stats[0].IntervalID)
	assert.Equal(t, "hour", stats[0].Unit)
	assert.Equal(t, "acc-123", stats[0].AccountID)
}

func TestGetAccountStats_Success_NoParams(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("GET /accounts/acc-456/stats", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		assert.Empty(t, r.URL.RawQuery)
		writeJSON(w, http.StatusOK, []StatsResponse{
			{IntervalID: "2026-03-05:00:00", Unit: "minute", Schema: "https://schemas.ably.com/json/app-stats-0.0.1.json", Entries: map[string]interface{}{}, AccountID: "acc-456"},
		})
	})

	stats, err := client.GetAccountStats(context.Background(), "acc-456", nil)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, "acc-456", stats[0].AccountID)
}

func TestGetAccountStats_Errors(t *testing.T) {
	t.Parallel()
	call := func(acctID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.GetAccountStats(ctx, acctID, nil)
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "AuthFailure", pattern: "GET /accounts/acc-123/stats", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("acc-123")},
		{name: "NotFound", pattern: "GET /accounts/nonexistent/stats", status: 404, message: "Account not found", code: 40400, call: call("nonexistent")},
		{name: "BadRequest", pattern: "GET /accounts/acc-123/stats", status: 400, message: "Invalid unit parameter", code: 40000, call: call("acc-123")},
		{name: "ServerError", pattern: "GET /accounts/acc-123/stats", status: 500, message: "Internal server error", code: 50000, call: call("acc-123")},
	})
}

func TestGetAccountStats_ContextCancelled(t *testing.T) {
	testContextCanceled(t, "GET /accounts/acc-123/stats", []StatsResponse{}, func(ctx context.Context, c *Client) error {
		_, err := c.GetAccountStats(ctx, "acc-123", nil)
		return err
	})
}

// ---------------------------------------------------------------------------
// Integration: Account Stats
// ---------------------------------------------------------------------------

func TestIntegration_AccountStats(t *testing.T) {
	client := integrationClient(t)

	me, err := client.Me(context.Background())
	require.NoError(t, err)

	_, err = client.GetAccountStats(context.Background(), me.Account.ID, &StatsParams{Unit: "hour", Limit: ptr(1)})
	require.NoError(t, err)
}

func TestIntegration_AccountStats_NotFound(t *testing.T) {
	client := integrationClient(t)
	_, err := client.GetAccountStats(context.Background(), "nonexistent-account-id", nil)
	assertAPIError(t, err, 404)
}

func TestIntegration_AccountStats_BadToken(t *testing.T) {
	client := badTokenClient(t)
	_, err := client.GetAccountStats(context.Background(), "any-account-id", nil)
	assertAPIError(t, err, 401)
}
