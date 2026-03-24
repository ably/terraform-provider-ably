package control

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListKeys returns all API keys for an app. The full list is returned
// in a single request (no pagination).
func (c *Client) ListKeys(ctx context.Context, appID string) ([]KeyResponse, error) {
	var result []KeyResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/keys", url.PathEscape(appID)), nil, &result)
	return result, err
}

// CreateKey creates an API key and returns its full representation,
// including the Key field which contains the full API key string. This
// is the only time the full key is available.
func (c *Client) CreateKey(ctx context.Context, appID string, body KeyPost) (KeyResponse, error) {
	var result KeyResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/keys", url.PathEscape(appID)), body, &result)
	return result, err
}

// UpdateKey performs a partial update (PATCH) on an API key. Only
// non-zero fields in body are sent.
func (c *Client) UpdateKey(ctx context.Context, appID string, keyID string, body KeyPatch) (KeyResponse, error) {
	var result KeyResponse
	err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("apps/%s/keys/%s", url.PathEscape(appID), url.PathEscape(keyID)), body, &result)
	return result, err
}

// RevokeKey permanently revokes an API key. Revoked keys cannot be
// reinstated. Returns [*Error] with StatusCode 404 if the key does
// not exist.
func (c *Client) RevokeKey(ctx context.Context, appID string, keyID string) error {
	return c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/keys/%s/revoke", url.PathEscape(appID), url.PathEscape(keyID)), nil, nil)
}
