package ably

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListQueues lists all queues for an app.
func (c *Client) ListQueues(ctx context.Context, appID string) ([]QueueResponse, error) {
	var result []QueueResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/queues", url.PathEscape(appID)), nil, &result)
	return result, err
}

// CreateQueue creates a new queue for an app.
func (c *Client) CreateQueue(ctx context.Context, appID string, body Queue) (QueueResponse, error) {
	var result QueueResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/queues", url.PathEscape(appID)), body, &result)
	return result, err
}

// DeleteQueue deletes a queue.
func (c *Client) DeleteQueue(ctx context.Context, appID string, queueID string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("apps/%s/queues/%s", url.PathEscape(appID), url.PathEscape(queueID)), nil, nil)
}
