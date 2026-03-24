package control

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
)

// ListApps returns all apps in the given account. The full list is
// returned in a single request (no pagination).
func (c *Client) ListApps(ctx context.Context, accountID string) ([]AppResponse, error) {
	var result []AppResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("accounts/%s/apps", url.PathEscape(accountID)), nil, &result)
	return result, err
}

// CreateApp creates an app and returns its full representation,
// including the server-assigned ID.
func (c *Client) CreateApp(ctx context.Context, accountID string, body AppPost) (AppResponse, error) {
	var result AppResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("accounts/%s/apps", url.PathEscape(accountID)), body, &result)
	return result, err
}

// UpdateApp performs a partial update (PATCH). Only non-zero fields in
// body are sent; unset fields are left unchanged on the server.
func (c *Client) UpdateApp(ctx context.Context, appID string, body AppPatch) (AppResponse, error) {
	var result AppResponse
	err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("apps/%s", url.PathEscape(appID)), body, &result)
	return result, err
}

// DeleteApp permanently deletes an app and all of its associated
// resources (keys, namespaces, queues, rules). Returns [*Error] with
// StatusCode 404 if the app does not exist.
func (c *Client) DeleteApp(ctx context.Context, appID string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("apps/%s", url.PathEscape(appID)), nil, nil)
}

// UpdateAppPKCS12 uploads a DER-encoded PKCS#12 archive as the app's
// push notification certificate (APNs). p12Pass is the password that
// protects the archive. The upload replaces any previously configured
// certificate.
func (c *Client) UpdateAppPKCS12(ctx context.Context, appID string, p12Data []byte, p12Pass string) (AppResponse, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("p12File", "cert.p12")
	if err != nil {
		return AppResponse{}, err
	}
	if _, err = fw.Write(p12Data); err != nil {
		return AppResponse{}, err
	}

	pw, err := w.CreateFormField("p12Pass")
	if err != nil {
		return AppResponse{}, err
	}
	if _, err = pw.Write([]byte(p12Pass)); err != nil {
		return AppResponse{}, err
	}

	if err = w.Close(); err != nil {
		return AppResponse{}, err
	}

	var result AppResponse
	err = c.doCustom(ctx, http.MethodPost, fmt.Sprintf("apps/%s/pkcs12", url.PathEscape(appID)), w.FormDataContentType(), &buf, &result)
	return result, err
}

// GetAppStats returns usage statistics for the given app. Pass nil for
// params to use API defaults.
func (c *Client) GetAppStats(ctx context.Context, appID string, params *StatsParams) ([]StatsResponse, error) {
	path := addStatsParams(fmt.Sprintf("apps/%s/stats", url.PathEscape(appID)), params)
	var result []StatsResponse
	err := c.doJSON(ctx, http.MethodGet, path, nil, &result)
	return result, err
}
