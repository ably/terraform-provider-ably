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
// ListQueues
// ---------------------------------------------------------------------------

func TestListQueues_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	want := []QueueResponse{
		{ID: "queue1", AppID: "app123", Name: "my-queue", Region: "us-east-1", State: "running", TTL: 60},
		{ID: "queue2", AppID: "app123", Name: "other-queue", Region: "eu-west-1", State: "running", TTL: 120},
	}

	mux.HandleFunc("GET /apps/app123/queues", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		writeJSON(w, http.StatusOK, want)
	})

	got, err := client.ListQueues(context.Background(), "app123")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "queue1", got[0].ID)
	assert.Equal(t, "my-queue", got[0].Name)
	assert.Equal(t, "queue2", got[1].ID)
	assert.Equal(t, "other-queue", got[1].Name)
}

func TestListQueues_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.ListQueues(ctx, appID)
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "GET /apps/app123/queues", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123")},
		{name: "NotFound", pattern: "GET /apps/app999/queues", status: 404, message: "App not found", code: 40400, call: call("app999")},
		{name: "ServerError", pattern: "GET /apps/app123/queues", status: 500, message: "Internal server error", code: 50000, call: call("app123")},
		{name: "ServiceUnavailable", pattern: "GET /apps/app123/queues", status: 503, message: "Service unavailable", code: 50300, call: call("app123")},
		{name: "GatewayTimeout", pattern: "GET /apps/app123/queues", status: 504, message: "Gateway timeout", code: 50400, call: call("app123")},
	})
}

func TestListQueues_ContextCancelled(t *testing.T) {
	testContextCanceled(t, "GET /apps/app123/queues", []QueueResponse{}, func(ctx context.Context, c *Client) error {
		_, err := c.ListQueues(ctx, "app123")
		return err
	})
}

// ---------------------------------------------------------------------------
// CreateQueue
// ---------------------------------------------------------------------------

func TestCreateQueue_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	body := Queue{Name: "my-queue", TTL: 60, MaxLength: 10000, Region: "us-east-1"}
	want := QueueResponse{ID: "queue1", AppID: "app123", Name: "my-queue", Region: "us-east-1", State: "running", TTL: 60, MaxLength: 10000}

	mux.HandleFunc("POST /apps/app123/queues", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}

		var got Queue
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "my-queue", got.Name)
		assert.Equal(t, 60, got.TTL)
		assert.Equal(t, 10000, got.MaxLength)
		assert.Equal(t, "us-east-1", got.Region)

		writeJSON(w, http.StatusCreated, want)
	})

	got, err := client.CreateQueue(context.Background(), "app123", body)
	require.NoError(t, err)
	assert.Equal(t, "queue1", got.ID)
	assert.Equal(t, "app123", got.AppID)
	assert.Equal(t, "my-queue", got.Name)
	assert.Equal(t, "us-east-1", got.Region)
	assert.Equal(t, 60, got.TTL)
	assert.Equal(t, 10000, got.MaxLength)
}

func TestCreateQueue_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.CreateQueue(ctx, appID, Queue{Name: "test"})
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "POST /apps/app123/queues", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123")},
		{name: "NotFound", pattern: "POST /apps/app999/queues", status: 404, message: "App not found", code: 40400, call: call("app999")},
		{name: "BadRequest", pattern: "POST /apps/app123/queues", status: 400, message: "Missing required field: region", code: 40000, call: call("app123")},
		{name: "UnprocessableEntity", pattern: "POST /apps/app123/queues", status: 422, message: "Queue name already exists", code: 42200, call: call("app123")},
		{name: "ServerError", pattern: "POST /apps/app123/queues", status: 500, message: "Internal server error", code: 50000, call: call("app123")},
	})
}

// ---------------------------------------------------------------------------
// DeleteQueue
// ---------------------------------------------------------------------------

func TestDeleteQueue_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("DELETE /apps/app123/queues/queue1", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteQueue(context.Background(), "app123", "queue1")
	require.NoError(t, err)
}

func TestDeleteQueue_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID, queueID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			return c.DeleteQueue(ctx, appID, queueID)
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "Unauthorized", pattern: "DELETE /apps/app123/queues/queue1", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app123", "queue1")},
		{name: "NotFound", pattern: "DELETE /apps/app123/queues/nonexistent", status: 404, message: "Queue not found", code: 40400, call: call("app123", "nonexistent")},
		{name: "BadRequest", pattern: "DELETE /apps/app123/queues/queue1", status: 400, message: "Cannot delete active queue", code: 40000, call: call("app123", "queue1")},
		{name: "ServerError", pattern: "DELETE /apps/app123/queues/queue1", status: 500, message: "Internal server error", code: 50000, call: call("app123", "queue1")},
		{name: "ServiceUnavailable", pattern: "DELETE /apps/app123/queues/queue1", status: 503, message: "Service unavailable", code: 50300, call: call("app123", "queue1")},
	})
}

// ---------------------------------------------------------------------------
// Integration: Queues CRUD
// ---------------------------------------------------------------------------

func TestIntegration_QueuesCRUD(t *testing.T) {
	client := integrationClient(t)
	suffix := testSuffix()
	appID := createTestApp(t, client, "integ-queues-"+suffix)
	queueName := "test-queue-" + suffix

	var queueID string

	t.Run("Create", func(t *testing.T) {
		queue, err := client.CreateQueue(context.Background(), appID, Queue{Name: queueName, TTL: 60, MaxLength: 10000, Region: "us-east-1-a"})
		require.NoError(t, err)
		require.NotEmpty(t, queue.ID)
		queueID = queue.ID
		assert.Equal(t, queueName, queue.Name)
	})

	t.Run("List", func(t *testing.T) {
		queues, err := client.ListQueues(context.Background(), appID)
		require.NoError(t, err)
		assert.True(t, slices.ContainsFunc(queues, func(q QueueResponse) bool { return q.ID == queueID }),
			"created queue %s not found in list", queueID)
	})

	t.Run("Delete", func(t *testing.T) {
		err := client.DeleteQueue(context.Background(), appID, queueID)
		require.NoError(t, err)
	})
}

func TestIntegration_Queues_NotFound(t *testing.T) {
	client := integrationClient(t)
	_, err := client.ListQueues(context.Background(), "nonexistent-app-id")
	assertAPIError(t, err, 404)
}

func TestIntegration_Queues_BadToken(t *testing.T) {
	client := badTokenClient(t)

	_, err := client.ListQueues(context.Background(), "any-app-id")
	assertAPIError(t, err, 401)

	_, err = client.CreateQueue(context.Background(), "any-app-id", Queue{Name: "x"})
	assertAPIError(t, err, 401)

	err = client.DeleteQueue(context.Background(), "any-app-id", "any-queue-id")
	assertAPIError(t, err, 401)
}
