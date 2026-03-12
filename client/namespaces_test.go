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
// ListNamespaces
// ---------------------------------------------------------------------------

func TestListNamespaces_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	want := []NamespaceResponse{
		{ID: "ns1", AppID: "app123", Persisted: true},
		{ID: "ns2", AppID: "app123", TLSOnly: true},
	}

	mux.HandleFunc("GET /apps/app123/namespaces", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		writeJSON(w, http.StatusOK, want)
	})

	got, err := client.ListNamespaces(context.Background(), "app123")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "ns1", got[0].ID)
	assert.True(t, got[0].Persisted)
	assert.Equal(t, "ns2", got[1].ID)
	assert.True(t, got[1].TLSOnly)
}

func TestListNamespaces_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.ListNamespaces(ctx, appID)
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "GET /apps/app123/namespaces", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123")},
		{name: "NotFound", pattern: "GET /apps/app999/namespaces", status: 404, message: "App not found", code: 40400, call: call("app999")},
		{name: "ServerError", pattern: "GET /apps/app123/namespaces", status: 500, message: "Internal server error", code: 50000, call: call("app123")},
		{name: "GatewayTimeout", pattern: "GET /apps/app123/namespaces", status: 504, message: "Gateway timeout", code: 50400, call: call("app123")},
	})
}

func TestListNamespaces_ContextCancelled(t *testing.T) {
	testContextCanceled(t, "GET /apps/app123/namespaces", []NamespaceResponse{}, func(ctx context.Context, c *Client) error {
		_, err := c.ListNamespaces(ctx, "app123")
		return err
	})
}

// ---------------------------------------------------------------------------
// CreateNamespace
// ---------------------------------------------------------------------------

func TestCreateNamespace_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := NamespacePost{
		ID:          "chat",
		Persisted:   true,
		PersistLast: true,
		TLSOnly:     true,
	}

	want := NamespaceResponse{
		ID:          "chat",
		AppID:       "app123",
		Persisted:   true,
		PersistLast: true,
		TLSOnly:     true,
		Created:     1700000000000,
		Modified:    1700000000000,
	}

	mux.HandleFunc("POST /apps/app123/namespaces", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got NamespacePost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "chat", got.ID)
		assert.True(t, got.Persisted)
		assert.True(t, got.PersistLast)
		assert.True(t, got.TLSOnly)
		writeJSON(w, http.StatusCreated, want)
	})

	got, err := client.CreateNamespace(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "chat", got.ID)
	assert.Equal(t, "app123", got.AppID)
	assert.True(t, got.Persisted)
	assert.True(t, got.PersistLast)
	assert.True(t, got.TLSOnly)
}

func TestCreateNamespace_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.CreateNamespace(ctx, appID, NamespacePost{ID: "chat"})
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "POST /apps/app123/namespaces", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123")},
		{name: "NotFound", pattern: "POST /apps/app999/namespaces", status: 404, message: "App not found", code: 40400, call: call("app999")},
		{name: "BadRequest", pattern: "POST /apps/app123/namespaces", status: 400, message: "Missing required field: id", code: 40000, call: call("app123")},
		{name: "UnprocessableEntity", pattern: "POST /apps/app123/namespaces", status: 422, message: "Namespace ID already exists", code: 42200, call: call("app123")},
		{name: "ServerError", pattern: "POST /apps/app123/namespaces", status: 500, message: "Internal server error", code: 50000, call: call("app123")},
	})
}

// ---------------------------------------------------------------------------
// UpdateNamespace
// ---------------------------------------------------------------------------

func TestUpdateNamespace_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := NamespacePatch{Persisted: ptr(true)}
	want := NamespaceResponse{
		ID:        "chat",
		AppID:     "app123",
		Persisted: true,
		Modified:  1700000001000,
	}

	mux.HandleFunc("PATCH /apps/app123/namespaces/chat", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		var got NamespacePatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		require.NotNil(t, got.Persisted)
		assert.True(t, *got.Persisted)
		writeJSON(w, http.StatusOK, want)
	})

	got, err := client.UpdateNamespace(context.Background(), "app123", "chat", body)
	require.NoError(t, err)
	assert.Equal(t, "chat", got.ID)
	assert.True(t, got.Persisted)
}

func TestUpdateNamespace_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID, nsID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.UpdateNamespace(ctx, appID, nsID, NamespacePatch{})
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "PATCH /apps/app123/namespaces/chat", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123", "chat")},
		{name: "NotFound", pattern: "PATCH /apps/app123/namespaces/nonexistent", status: 404, message: "Namespace not found", code: 40400, call: call("app123", "nonexistent")},
		{name: "BadRequest", pattern: "PATCH /apps/app123/namespaces/chat", status: 400, message: "Invalid field value", code: 40000, call: call("app123", "chat")},
		{name: "ServerError", pattern: "PATCH /apps/app123/namespaces/chat", status: 500, message: "Internal server error", code: 50000, call: call("app123", "chat")},
		{name: "GatewayTimeout", pattern: "PATCH /apps/app123/namespaces/chat", status: 504, message: "Gateway timeout", code: 50400, call: call("app123", "chat")},
	})
}

// ---------------------------------------------------------------------------
// DeleteNamespace
// ---------------------------------------------------------------------------

func TestDeleteNamespace_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("DELETE /apps/app123/namespaces/chat", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteNamespace(context.Background(), "app123", "chat")
	require.NoError(t, err)
}

func TestDeleteNamespace_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID, nsID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			return c.DeleteNamespace(ctx, appID, nsID)
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "DELETE /apps/app123/namespaces/chat", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123", "chat")},
		{name: "NotFound", pattern: "DELETE /apps/app123/namespaces/nonexistent", status: 404, message: "Namespace not found", code: 40400, call: call("app123", "nonexistent")},
		{name: "ServerError", pattern: "DELETE /apps/app123/namespaces/chat", status: 500, message: "Internal server error", code: 50000, call: call("app123", "chat")},
		{name: "GatewayTimeout", pattern: "DELETE /apps/app123/namespaces/chat", status: 504, message: "Gateway timeout", code: 50400, call: call("app123", "chat")},
	})
}

// ---------------------------------------------------------------------------
// Integration: Namespaces CRUD
// ---------------------------------------------------------------------------

func TestIntegration_NamespacesCRUD(t *testing.T) {
	client := integrationClient(t)
	suffix := testSuffix()
	appID := createTestApp(t, client, "integ-ns-"+suffix)
	nsID := "test-ns-" + suffix

	t.Run("Create", func(t *testing.T) {
		ns, err := client.CreateNamespace(context.Background(), appID, NamespacePost{ID: nsID, Persisted: true})
		require.NoError(t, err)
		assert.Equal(t, nsID, ns.ID)
		assert.True(t, ns.Persisted)
	})

	t.Run("List", func(t *testing.T) {
		namespaces, err := client.ListNamespaces(context.Background(), appID)
		require.NoError(t, err)
		assert.True(t, slices.ContainsFunc(namespaces, func(n NamespaceResponse) bool { return n.ID == nsID }),
			"created namespace not found in list")
	})

	t.Run("Update", func(t *testing.T) {
		updated, err := client.UpdateNamespace(context.Background(), appID, nsID, NamespacePatch{PersistLast: ptr(true)})
		require.NoError(t, err)
		assert.True(t, updated.PersistLast)
	})

	t.Run("Delete", func(t *testing.T) {
		err := client.DeleteNamespace(context.Background(), appID, nsID)
		require.NoError(t, err)
	})
}

func TestIntegration_Namespaces_NotFound(t *testing.T) {
	client := integrationClient(t)

	_, err := client.UpdateNamespace(context.Background(), "nonexistent-app-id", "ns", NamespacePatch{})
	assertAPIError(t, err, 404)

	err = client.DeleteNamespace(context.Background(), "nonexistent-app-id", "ns")
	assertAPIError(t, err, 404)
}

func TestIntegration_Namespaces_BadToken(t *testing.T) {
	client := badTokenClient(t)

	_, err := client.ListNamespaces(context.Background(), "any-app-id")
	assertAPIError(t, err, 401)

	_, err = client.CreateNamespace(context.Background(), "any-app-id", NamespacePost{ID: "x"})
	assertAPIError(t, err, 401)

	_, err = client.UpdateNamespace(context.Background(), "any-app-id", "any-ns-id", NamespacePatch{})
	assertAPIError(t, err, 401)

	err = client.DeleteNamespace(context.Background(), "any-app-id", "any-ns-id")
	assertAPIError(t, err, 401)
}
