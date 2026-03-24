package control

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ListKeys
// ---------------------------------------------------------------------------

func TestListKeys_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	want := []KeyResponse{
		{ID: "key1", AppID: "app123", Name: "Key One", Status: 0, Key: "app123.key1:secret1"},
		{ID: "key2", AppID: "app123", Name: "Key Two", Status: 0, Key: "app123.key2:secret2"},
	}

	mux.HandleFunc("GET /apps/app123/keys", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		writeJSON(w, http.StatusOK, want)
	})

	got, err := client.ListKeys(context.Background(), "app123")
	require.NoError(t, err)
	require.Len(t, got, len(want))
	for i := range want {
		assert.Equal(t, want[i].ID, got[i].ID)
		assert.Equal(t, want[i].Name, got[i].Name)
		assert.Equal(t, want[i].Key, got[i].Key)
	}
}

func TestListKeys_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.ListKeys(ctx, appID)
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "GET /apps/app123/keys", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123")},
		{name: "NotFound", pattern: "GET /apps/app999/keys", status: 404, message: "App not found", code: 40400, call: call("app999")},
		{name: "ServerError", pattern: "GET /apps/app123/keys", status: 500, message: "Internal server error", code: 50000, call: call("app123")},
		{name: "GatewayTimeout", pattern: "GET /apps/app123/keys", status: 504, message: "Gateway timeout", code: 50400, call: call("app123")},
	})
}

func TestListKeys_ContextCancelled(t *testing.T) {
	testContextCanceled(t, "GET /apps/app123/keys", []KeyResponse{}, func(ctx context.Context, c *Client) error {
		_, err := c.ListKeys(ctx, "app123")
		return err
	})
}

// ---------------------------------------------------------------------------
// CreateKey
// ---------------------------------------------------------------------------

func TestCreateKey_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := KeyPost{
		Name: "My New Key",
		Capability: map[string][]string{
			"channel:*": {"publish", "subscribe"},
		},
	}

	want := KeyResponse{
		ID:    "key789",
		AppID: "app123",
		Name:  "My New Key",
		Key:   "app123.key789:newsecret",
		Capability: map[string][]string{
			"channel:*": {"publish", "subscribe"},
		},
	}

	mux.HandleFunc("POST /apps/app123/keys", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}

		var got KeyPost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, body.Name, got.Name)

		writeJSON(w, http.StatusCreated, want)
	})

	got, err := client.CreateKey(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.AppID, got.AppID)
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.Key, got.Key)
}

func TestCreateKey_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.CreateKey(ctx, appID, KeyPost{Name: "x"})
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "POST /apps/app123/keys", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123")},
		{name: "NotFound", pattern: "POST /apps/app999/keys", status: 404, message: "App not found", code: 40400, call: call("app999")},
		{name: "BadRequest", pattern: "POST /apps/app123/keys", status: 400, message: "Missing required field: capability", code: 40000, call: call("app123")},
		{name: "UnprocessableEntity", pattern: "POST /apps/app123/keys", status: 422, message: "Invalid capability", code: 42200, call: call("app123")},
		{name: "ServerError", pattern: "POST /apps/app123/keys", status: 500, message: "Internal server error", code: 50000, call: call("app123")},
	})
}

// ---------------------------------------------------------------------------
// UpdateKey
// ---------------------------------------------------------------------------

func TestUpdateKey_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := KeyPatch{Name: "Updated Key Name"}
	want := KeyResponse{
		ID:    "key456",
		AppID: "app123",
		Name:  "Updated Key Name",
		Key:   "app123.key456:secret",
	}

	mux.HandleFunc("PATCH /apps/app123/keys/key456", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}

		var got KeyPatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, body.Name, got.Name)

		writeJSON(w, http.StatusOK, want)
	})

	got, err := client.UpdateKey(context.Background(), "app123", "key456", body)
	require.NoError(t, err)
	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.Key, got.Key)
}

func TestUpdateKey_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID, keyID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.UpdateKey(ctx, appID, keyID, KeyPatch{Name: "x"})
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "PATCH /apps/app123/keys/key456", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123", "key456")},
		{name: "NotFound", pattern: "PATCH /apps/app123/keys/nonexistent", status: 404, message: "Key not found", code: 40400, call: call("app123", "nonexistent")},
		{name: "BadRequest", pattern: "PATCH /apps/app123/keys/key456", status: 400, message: "Invalid field", code: 40000, call: call("app123", "key456")},
		{name: "UnprocessableEntity", pattern: "PATCH /apps/app123/keys/key456", status: 422, message: "Invalid capability format", code: 42200, call: call("app123", "key456")},
		{name: "ServerError", pattern: "PATCH /apps/app123/keys/key456", status: 500, message: "Internal server error", code: 50000, call: call("app123", "key456")},
		{name: "GatewayTimeout", pattern: "PATCH /apps/app123/keys/key456", status: 504, message: "Gateway timeout", code: 50400, call: call("app123", "key456")},
	})
}

// ---------------------------------------------------------------------------
// RevokeKey
// ---------------------------------------------------------------------------

func TestRevokeKey_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("POST /apps/app123/keys/key456/revoke", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	err := client.RevokeKey(context.Background(), "app123", "key456")
	require.NoError(t, err)
}

func TestRevokeKey_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID, keyID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			return c.RevokeKey(ctx, appID, keyID)
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "POST /apps/app123/keys/key456/revoke", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123", "key456")},
		{name: "NotFound", pattern: "POST /apps/app123/keys/nonexistent/revoke", status: 404, message: "Key not found", code: 40400, call: call("app123", "nonexistent")},
		{name: "ServerError", pattern: "POST /apps/app123/keys/key456/revoke", status: 500, message: "Internal server error", code: 50000, call: call("app123", "key456")},
		{name: "GatewayTimeout", pattern: "POST /apps/app123/keys/key456/revoke", status: 504, message: "Gateway timeout", code: 50400, call: call("app123", "key456")},
	})
}

// ---------------------------------------------------------------------------
// Integration: Keys CRUD
// ---------------------------------------------------------------------------

func TestIntegration_KeysCRUD(t *testing.T) {
	client := integrationClient(t)
	suffix := testSuffix()
	appID := createTestApp(t, client, "integ-keys-"+suffix)

	var keyID string
	keyName := "test-key-" + suffix

	t.Run("Create", func(t *testing.T) {
		key, err := client.CreateKey(context.Background(), appID, KeyPost{
			Name: keyName,
			Capability: map[string][]string{
				"[*]*": {
					"subscribe", "publish", "presence",
					"object-subscribe", "object-publish",
					"annotation-subscribe", "annotation-publish",
					"message-update-own", "message-update-any",
					"message-delete-own", "message-delete-any",
					"history", "statistics",
					"push-subscribe", "push-admin",
					"channel-metadata", "privileged-headers",
				},
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, key.ID)
		keyID = key.ID
		assert.Equal(t, keyName, key.Name)
	})

	t.Run("List", func(t *testing.T) {
		keys, err := client.ListKeys(context.Background(), appID)
		require.NoError(t, err)
		assert.NotEmpty(t, keys)
	})

	t.Run("Update", func(t *testing.T) {
		updated, err := client.UpdateKey(context.Background(), appID, keyID, KeyPatch{Name: keyName + "-updated"})
		require.NoError(t, err)
		assert.Equal(t, keyName+"-updated", updated.Name)
	})

	t.Run("Revoke", func(t *testing.T) {
		err := client.RevokeKey(context.Background(), appID, keyID)
		require.NoError(t, err)
	})
}

func TestIntegration_Keys_NotFound(t *testing.T) {
	client := integrationClient(t)
	_, err := client.ListKeys(context.Background(), "nonexistent-app-id")
	assertAPIError(t, err, 404)
}

func TestIntegration_Keys_BadToken(t *testing.T) {
	client := badTokenClient(t)

	_, err := client.ListKeys(context.Background(), "any-app-id")
	assertAPIError(t, err, 401)

	_, err = client.CreateKey(context.Background(), "any-app-id", KeyPost{Name: "x", Capability: map[string][]string{"*": {"publish"}}})
	assertAPIError(t, err, 401)

	_, err = client.UpdateKey(context.Background(), "any-app-id", "any-key-id", KeyPatch{Name: "x"})
	assertAPIError(t, err, 401)

	err = client.RevokeKey(context.Background(), "any-app-id", "any-key-id")
	assertAPIError(t, err, 401)
}
