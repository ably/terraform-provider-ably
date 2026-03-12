package ably

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ListRules
// ---------------------------------------------------------------------------

func TestListRules_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	want := []RuleResponse{
		{ID: "rule001", AppID: "app123", Status: "enabled", RuleType: "http", RequestMode: "single", Source: &RuleSource{ChannelFilter: "^my-channel", Type: "channel.message"}},
		{ID: "rule002", AppID: "app123", Status: "enabled", RuleType: "http"},
	}

	mux.HandleFunc("GET /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		writeJSON(w, http.StatusOK, want)
	})

	got, err := client.ListRules(context.Background(), "app123")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "rule001", got[0].ID)
	assert.Equal(t, "http", got[0].RuleType)
	assert.Equal(t, "single", got[0].RequestMode)
	require.NotNil(t, got[0].Source)
	assert.Equal(t, "^my-channel", got[0].Source.ChannelFilter)
	assert.Equal(t, "channel.message", got[0].Source.Type)
	assert.Equal(t, "rule002", got[1].ID)
}

func TestListRules_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.ListRules(ctx, appID)
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "GET /apps/app123/rules", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123")},
		{name: "NotFound", pattern: "GET /apps/app999/rules", status: 404, message: "App not found", code: 40400, call: call("app999")},
		{name: "ServerError", pattern: "GET /apps/app123/rules", status: 500, message: "Internal server error", code: 50000, call: call("app123")},
		{name: "GatewayTimeout", pattern: "GET /apps/app123/rules", status: 504, message: "Gateway timeout", code: 50400, call: call("app123")},
	})
}

func TestListRules_ContextCancelled(t *testing.T) {
	testContextCanceled(t, "GET /apps/app123/rules", []RuleResponse{}, func(ctx context.Context, c *Client) error {
		_, err := c.ListRules(ctx, "app123")
		return err
	})
}

// ---------------------------------------------------------------------------
// CreateRule
// ---------------------------------------------------------------------------

func TestCreateRule_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := HTTPRulePost{
		RuleType: "http", RequestMode: "single",
		Source: RuleSource{ChannelFilter: "^my-channel", Type: "channel.message"},
		Target: HTTPRuleTarget{URL: "https://example.com/webhook", Format: "json"},
	}
	wantResp := RuleResponse{
		ID: "rule789", AppID: "app123", Status: "enabled", RuleType: "http", RequestMode: "single",
		Source: &RuleSource{ChannelFilter: "^my-channel", Type: "channel.message"},
	}

	mux.HandleFunc("POST /apps/app123/rules", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got HTTPRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "http", got.RuleType)
		assert.Equal(t, "single", got.RequestMode)
		assert.Equal(t, "^my-channel", got.Source.ChannelFilter)
		assert.Equal(t, "https://example.com/webhook", got.Target.URL)
		writeJSON(w, http.StatusCreated, wantResp)
	})

	got, err := client.CreateRule(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "rule789", got.ID)
	assert.Equal(t, "enabled", got.Status)
	assert.Equal(t, "http", got.RuleType)
	require.NotNil(t, got.Source)
	assert.Equal(t, "^my-channel", got.Source.ChannelFilter)
}

func TestCreateRule_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.CreateRule(ctx, appID, HTTPRulePost{RuleType: "http"})
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "POST /apps/app123/rules", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123")},
		{name: "NotFound", pattern: "POST /apps/app999/rules", status: 404, message: "App not found", code: 40400, call: call("app999")},
		{name: "BadRequest", pattern: "POST /apps/app123/rules", status: 400, message: "Missing required field: source", code: 40000, call: call("app123")},
		{name: "Forbidden", pattern: "POST /apps/app123/rules", status: 403, message: "Rule creation not allowed for this account", code: 40300, call: call("app123")},
		{name: "UnprocessableEntity", pattern: "POST /apps/app123/rules", status: 422, message: "Invalid rule configuration", code: 42200, call: call("app123")},
		{name: "ServerError", pattern: "POST /apps/app123/rules", status: 500, message: "Internal server error", code: 50000, call: call("app123")},
		{name: "GatewayTimeout", pattern: "POST /apps/app123/rules", status: 504, message: "Gateway timeout", code: 50400, call: call("app123")},
	})
}

// ---------------------------------------------------------------------------
// GetRule
// ---------------------------------------------------------------------------

func TestGetRule_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	wantResp := RuleResponse{
		ID: "rule789", AppID: "app123", Status: "enabled", RuleType: "http", RequestMode: "single",
		Source: &RuleSource{ChannelFilter: "^my-channel", Type: "channel.message"},
	}

	mux.HandleFunc("GET /apps/app123/rules/rule789", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.GetRule(context.Background(), "app123", "rule789")
	require.NoError(t, err)
	assert.Equal(t, "rule789", got.ID)
	assert.Equal(t, "http", got.RuleType)
	require.NotNil(t, got.Source)
	assert.Equal(t, "^my-channel", got.Source.ChannelFilter)
}

func TestGetRule_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID, ruleID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.GetRule(ctx, appID, ruleID)
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "GET /apps/app123/rules/rule789", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123", "rule789")},
		{name: "NotFound", pattern: "GET /apps/app123/rules/nonexistent", status: 404, message: "Rule not found", code: 40400, call: call("app123", "nonexistent")},
		{name: "ServerError", pattern: "GET /apps/app123/rules/rule789", status: 500, message: "Internal server error", code: 50000, call: call("app123", "rule789")},
		{name: "GatewayTimeout", pattern: "GET /apps/app123/rules/rule789", status: 504, message: "Gateway timeout", code: 50400, call: call("app123", "rule789")},
	})
}

// ---------------------------------------------------------------------------
// UpdateRule
// ---------------------------------------------------------------------------

func TestUpdateRule_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := HTTPRulePost{
		RuleType: "http", RequestMode: "single",
		Source: RuleSource{ChannelFilter: "^updated-channel", Type: "channel.message"},
		Target: HTTPRuleTarget{URL: "https://example.com/updated-webhook", Format: "json"},
	}
	wantResp := RuleResponse{
		ID: "rule789", AppID: "app123", Status: "enabled", RuleType: "http", RequestMode: "single",
		Source: &RuleSource{ChannelFilter: "^updated-channel", Type: "channel.message"},
	}

	mux.HandleFunc("PATCH /apps/app123/rules/rule789", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got HTTPRulePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "^updated-channel", got.Source.ChannelFilter)
		assert.Equal(t, "https://example.com/updated-webhook", got.Target.URL)
		writeJSON(w, http.StatusOK, wantResp)
	})

	got, err := client.UpdateRule(context.Background(), "app123", "rule789", body)
	require.NoError(t, err)
	assert.Equal(t, "rule789", got.ID)
	assert.Equal(t, "http", got.RuleType)
	require.NotNil(t, got.Source)
	assert.Equal(t, "^updated-channel", got.Source.ChannelFilter)
}

func TestUpdateRule_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID, ruleID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.UpdateRule(ctx, appID, ruleID, HTTPRulePost{RuleType: "http"})
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "PATCH /apps/app123/rules/rule789", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123", "rule789")},
		{name: "NotFound", pattern: "PATCH /apps/app123/rules/nonexistent", status: 404, message: "Rule not found", code: 40400, call: call("app123", "nonexistent")},
		{name: "BadRequest", pattern: "PATCH /apps/app123/rules/rule789", status: 400, message: "Invalid field value", code: 40000, call: call("app123", "rule789")},
		{name: "UnprocessableEntity", pattern: "PATCH /apps/app123/rules/rule789", status: 422, message: "Invalid rule configuration", code: 42200, call: call("app123", "rule789")},
		{name: "ServerError", pattern: "PATCH /apps/app123/rules/rule789", status: 500, message: "Internal server error", code: 50000, call: call("app123", "rule789")},
		{name: "GatewayTimeout", pattern: "PATCH /apps/app123/rules/rule789", status: 504, message: "Gateway timeout", code: 50400, call: call("app123", "rule789")},
	})
}

// ---------------------------------------------------------------------------
// DeleteRule
// ---------------------------------------------------------------------------

func TestDeleteRule_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)
	mux.HandleFunc("DELETE /apps/app123/rules/rule789", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteRule(context.Background(), "app123", "rule789")
	require.NoError(t, err)
}

func TestDeleteRule_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID, ruleID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			return c.DeleteRule(ctx, appID, ruleID)
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "DELETE /apps/app123/rules/rule789", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123", "rule789")},
		{name: "NotFound", pattern: "DELETE /apps/app123/rules/nonexistent", status: 404, message: "Rule not found", code: 40400, call: call("app123", "nonexistent")},
		{name: "ServerError", pattern: "DELETE /apps/app123/rules/rule789", status: 500, message: "Internal server error", code: 50000, call: call("app123", "rule789")},
		{name: "GatewayTimeout", pattern: "DELETE /apps/app123/rules/rule789", status: 504, message: "Gateway timeout", code: 50400, call: call("app123", "rule789")},
	})
}

// ---------------------------------------------------------------------------
// Integration: Rules CRUD
// ---------------------------------------------------------------------------

func TestIntegration_RulesCRUD(t *testing.T) {
	client := integrationClient(t)
	appID := createTestApp(t, client, "integ-rules-"+testSuffix())

	var ruleID string

	t.Run("Create", func(t *testing.T) {
		rule, err := client.CreateRule(context.Background(), appID, HTTPRulePost{
			RuleType: "http", RequestMode: "single",
			Source: RuleSource{ChannelFilter: "^test", Type: "channel.message"},
			Target: HTTPRuleTarget{URL: "https://example.com/webhook", Format: "json"},
		})
		require.NoError(t, err)
		require.NotEmpty(t, rule.ID)
		ruleID = rule.ID
	})

	t.Run("List", func(t *testing.T) {
		rules, err := client.ListRules(context.Background(), appID)
		require.NoError(t, err)
		assert.True(t, slices.ContainsFunc(rules, func(r RuleResponse) bool { return r.ID == ruleID }),
			"created rule %s not found in list", ruleID)
	})

	t.Run("Get", func(t *testing.T) {
		got, err := client.GetRule(context.Background(), appID, ruleID)
		require.NoError(t, err)
		assert.Equal(t, ruleID, got.ID)
	})

	t.Run("Update", func(t *testing.T) {
		updated, err := client.UpdateRule(context.Background(), appID, ruleID, HTTPRulePost{
			RuleType: "http", RequestMode: "single",
			Source: RuleSource{ChannelFilter: "^updated", Type: "channel.message"},
			Target: HTTPRuleTarget{URL: "https://example.com/updated", Format: "json"},
		})
		require.NoError(t, err)
		assert.Equal(t, ruleID, updated.ID)
	})

	t.Run("Delete", func(t *testing.T) {
		err := client.DeleteRule(context.Background(), appID, ruleID)
		require.NoError(t, err)
	})
}

func TestIntegration_Rules_NotFound(t *testing.T) {
	client := integrationClient(t)

	_, err := client.GetRule(context.Background(), "nonexistent-app-id", "nonexistent-rule-id")
	assertAPIError(t, err, 404)

	err = client.DeleteRule(context.Background(), "nonexistent-app-id", "nonexistent-rule-id")
	assertAPIError(t, err, 404)
}

func TestIntegration_Rules_BadToken(t *testing.T) {
	client := badTokenClient(t)

	_, err := client.ListRules(context.Background(), "any-app-id")
	assertAPIError(t, err, 401)

	_, err = client.CreateRule(context.Background(), "any-app-id", HTTPRulePost{RuleType: "http"})
	assertAPIError(t, err, 401)

	_, err = client.GetRule(context.Background(), "any-app-id", "any-rule-id")
	assertAPIError(t, err, 401)

	_, err = client.UpdateRule(context.Background(), "any-app-id", "any-rule-id", HTTPRulePost{RuleType: "http"})
	assertAPIError(t, err, 401)

	err = client.DeleteRule(context.Background(), "any-app-id", "any-rule-id")
	assertAPIError(t, err, 401)
}
