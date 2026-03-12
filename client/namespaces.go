package ably

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListNamespaces lists all namespaces for an app.
func (c *Client) ListNamespaces(ctx context.Context, appID string) ([]NamespaceResponse, error) {
	var result []NamespaceResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/namespaces", url.PathEscape(appID)), nil, &result)
	return result, err
}

// CreateNamespace creates a new namespace for an app.
func (c *Client) CreateNamespace(ctx context.Context, appID string, body NamespacePost) (NamespaceResponse, error) {
	var result NamespaceResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/namespaces", url.PathEscape(appID)), body, &result)
	return result, err
}

// UpdateNamespace updates an existing namespace.
func (c *Client) UpdateNamespace(ctx context.Context, appID string, nsID string, body NamespacePatch) (NamespaceResponse, error) {
	var result NamespaceResponse
	err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("apps/%s/namespaces/%s", url.PathEscape(appID), url.PathEscape(nsID)), body, &result)
	return result, err
}

// DeleteNamespace deletes a namespace.
func (c *Client) DeleteNamespace(ctx context.Context, appID string, nsID string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("apps/%s/namespaces/%s", url.PathEscape(appID), url.PathEscape(nsID)), nil, nil)
}
