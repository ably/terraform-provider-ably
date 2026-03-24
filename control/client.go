// Package control provides a Go client for the Ably Control API
// (https://control.ably.net/v1).
//
// The client manages Ably resources — apps, keys, namespaces, queues,
// and integration rules — via token-authenticated HTTP requests.
//
//	client := control.NewClient("your-control-api-token")
//	apps, err := client.ListApps(ctx, accountID)
//
// All methods return [*Error] for non-2xx API responses. Use [errors.As]
// to inspect status codes and Ably error codes:
//
//	var apiErr *control.Error
//	if errors.As(err, &apiErr) {
//	    log.Printf("status %d, code %d: %s", apiErr.StatusCode, apiErr.Code, apiErr.Message)
//	}
//
// Requests that fail with 5xx status codes are retried automatically
// (up to 4 times by default, with exponential backoff). Client errors
// (4xx) are never retried.
package control

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
)

// Version is the semantic version of this library, used in the default
// User-Agent header.
const Version = "0.1.0"

var defaultUserAgent = "ably-control-api/" + Version

// Client is an Ably Control API client. Use [NewClient] to create one.
type Client struct {
	// BaseURL is the base URL for the API (e.g. https://control.ably.net/v1).
	BaseURL string
	// Token is the bearer token for authentication.
	Token string
	// UserAgent is the User-Agent header sent with every request.
	// Defaults to "ably-control-api/<VERSION>". Consumers such as
	// Terraform or MCP servers can append their own identifier, e.g.
	//   client.UserAgent += " ably-terraform/1.0.0"
	UserAgent string
	// HTTPClient is the retryable HTTP client used for requests.
	// By default it retries up to 4 times with exponential backoff
	// on 5xx responses and connection errors.
	HTTPClient *retryablehttp.Client
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithRetryMax sets the maximum number of retries for failed requests.
// Set to 0 to disable retries. Default is 4.
func WithRetryMax(n int) ClientOption {
	return func(c *Client) {
		c.HTTPClient.RetryMax = n
	}
}

// WithUserAgent sets the User-Agent header. Use this to override the
// default entirely; to append, modify client.UserAgent after creation.
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.UserAgent = ua
	}
}

// WithHTTPClient sets the underlying [net/http.Client] used for transport.
// The retry layer remains active; this only replaces the transport and
// timeout configuration.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient.HTTPClient = hc
	}
}

// NewClient creates a new Ably Control API client.
//
// Defaults:
//   - Base URL: https://control.ably.net/v1
//   - User-Agent: ably-control-api/<Version>
//   - Retry: up to 4 attempts with exponential backoff on 5xx and
//     connection errors; 4xx responses are never retried
//
// Use [WithRetryMax], [WithUserAgent], or [WithHTTPClient] to override.
func NewClient(token string, opts ...ClientOption) *Client {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 4
	rc.Logger = nil // silence default logger
	rc.CheckRetry = retryPolicy
	rc.ErrorHandler = retryablehttp.PassthroughErrorHandler
	c := &Client{
		BaseURL:    "https://control.ably.net/v1",
		Token:      token,
		UserAgent:  defaultUserAgent,
		HTTPClient: rc,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// retryPolicy retries on 5xx and connection errors, but not on 4xx.
func retryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}
	if resp.StatusCode >= 500 {
		return true, nil
	}
	return false, nil
}

// newRequest creates a retryable HTTP request with context, auth, and user-agent headers.
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*retryablehttp.Request, error) {
	u := c.BaseURL + "/" + strings.TrimLeft(path, "/")
	req, err := retryablehttp.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("User-Agent", c.UserAgent)
	return req, nil
}

// checkResponse parses a non-2xx response into an *Error.
// The caller is responsible for closing resp.Body.
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	var apiErr Error
	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
		return &Error{Message: fmt.Sprintf("HTTP %d", resp.StatusCode), StatusCode: resp.StatusCode}
	}
	apiErr.StatusCode = resp.StatusCode
	return &apiErr
}

// doJSON performs a JSON request and decodes the response into result.
func (c *Client) doJSON(ctx context.Context, method, path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := c.newRequest(ctx, method, path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return err
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// doCustom performs a request with a pre-built body and content type, decoding the JSON response.
func (c *Client) doCustom(ctx context.Context, method, path, contentType string, body io.Reader, result interface{}) error {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkResponse(resp); err != nil {
		return err
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// addStatsParams appends StatsParams as query string parameters to a path.
func addStatsParams(path string, params *StatsParams) string {
	if params == nil {
		return path
	}
	v := url.Values{}
	if params.Start != nil {
		v.Set("start", strconv.Itoa(*params.Start))
	}
	if params.End != nil {
		v.Set("end", strconv.Itoa(*params.End))
	}
	if params.Unit != "" {
		v.Set("unit", params.Unit)
	}
	if params.Direction != "" {
		v.Set("direction", params.Direction)
	}
	if params.Limit != nil {
		v.Set("limit", strconv.Itoa(*params.Limit))
	}
	if len(v) > 0 {
		return path + "?" + v.Encode()
	}
	return path
}
