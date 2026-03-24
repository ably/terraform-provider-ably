package control

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListNamespaces returns all namespaces for an app. The full list is
// returned in a single request (no pagination).
func (c *Client) ListNamespaces(ctx context.Context, appID string) ([]NamespaceResponse, error) {
	var result []NamespaceResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/namespaces", url.PathEscape(appID)), nil, &result)
	return result, err
}

// CreateNamespace creates a namespace and returns its full
// representation.
func (c *Client) CreateNamespace(ctx context.Context, appID string, body NamespacePost) (NamespaceResponse, error) {
	var result NamespaceResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/namespaces", url.PathEscape(appID)), body, &result)
	return result, err
}

// UpdateNamespace performs a partial update (PATCH). Only non-nil
// pointer fields in body are sent.
func (c *Client) UpdateNamespace(ctx context.Context, appID string, nsID string, body NamespacePatch) (NamespaceResponse, error) {
	var result NamespaceResponse
	err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("apps/%s/namespaces/%s", url.PathEscape(appID), url.PathEscape(nsID)), body, &result)
	return result, err
}

// DeleteNamespace deletes a namespace. Returns [*Error] with
// StatusCode 404 if the namespace does not exist.
func (c *Client) DeleteNamespace(ctx context.Context, appID string, nsID string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("apps/%s/namespaces/%s", url.PathEscape(appID), url.PathEscape(nsID)), nil, nil)
}
