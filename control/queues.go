package control

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListQueues returns all queues for an app, including their current
// state and connection details. The full list is returned in a single
// request (no pagination).
func (c *Client) ListQueues(ctx context.Context, appID string) ([]QueueResponse, error) {
	var result []QueueResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/queues", url.PathEscape(appID)), nil, &result)
	return result, err
}

// CreateQueue creates a queue and returns its full representation,
// including AMQP and STOMP connection details.
func (c *Client) CreateQueue(ctx context.Context, appID string, body Queue) (QueueResponse, error) {
	var result QueueResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/queues", url.PathEscape(appID)), body, &result)
	return result, err
}

// DeleteQueue deletes a queue. Returns [*Error] with StatusCode 404 if
// the queue does not exist.
func (c *Client) DeleteQueue(ctx context.Context, appID string, queueID string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("apps/%s/queues/%s", url.PathEscape(appID), url.PathEscape(queueID)), nil, nil)
}
