package ably

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListKeys lists all keys for an app.
func (c *Client) ListKeys(ctx context.Context, appID string) ([]KeyResponse, error) {
	var result []KeyResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/keys", url.PathEscape(appID)), nil, &result)
	return result, err
}

// CreateKey creates a new key for an app.
func (c *Client) CreateKey(ctx context.Context, appID string, body KeyPost) (KeyResponse, error) {
	var result KeyResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/keys", url.PathEscape(appID)), body, &result)
	return result, err
}

// UpdateKey updates an existing key.
func (c *Client) UpdateKey(ctx context.Context, appID string, keyID string, body KeyPatch) (KeyResponse, error) {
	var result KeyResponse
	err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("apps/%s/keys/%s", url.PathEscape(appID), url.PathEscape(keyID)), body, &result)
	return result, err
}

// RevokeKey revokes an existing key.
func (c *Client) RevokeKey(ctx context.Context, appID string, keyID string) error {
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/keys/%s/revoke", url.PathEscape(appID), url.PathEscape(keyID)), nil, nil)
}
