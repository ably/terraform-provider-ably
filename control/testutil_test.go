package control

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testSuffix returns a short random hex string for unique test resource names.
func testSuffix() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// newTestServer creates an httptest.Server that routes requests to the given handler.
// It returns the server and a Client pointed at it.
func newTestServer(t *testing.T, handler http.Handler) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient("test-token")
	c.BaseURL = srv.URL
	c.HTTPClient.RetryMax = 0 // no retries in unit tests
	return srv, c
}

// newTestServerRaw creates an httptest.Server without a pre-configured client.
// Use this when you need to construct a Client with custom options.
func newTestServerRaw(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// newTestMux creates an http.ServeMux, an httptest.Server, and a Client.
func newTestMux(t *testing.T) (*http.ServeMux, *Client) {
	t.Helper()
	mux := http.NewServeMux()
	_, c := newTestServer(t, mux)
	return mux, c
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes an API error response.
func writeError(w http.ResponseWriter, statusCode int, message string, code int) {
	writeJSON(w, statusCode, Error{
		Message:    message,
		Code:       code,
		StatusCode: statusCode,
		Href:       "https://help.ably.io/error/" + http.StatusText(statusCode),
	})
}

// requireBearerToken checks the Authorization header.
func requireBearerToken(t *testing.T, w http.ResponseWriter, r *http.Request, token string) bool {
	t.Helper()
	got := r.Header.Get("Authorization")
	want := "Bearer " + token
	if got != want {
		writeError(w, 401, "Authentication failed", 40100)
		return false
	}
	return true
}

// ptr returns a pointer to v. Useful for constructing request bodies with optional fields.
func ptr[T any](v T) *T { return &v }

// assertAPIError asserts that err is an *Error with the expected status code.
func assertAPIError(t *testing.T, err error, expectedStatus int) {
	t.Helper()
	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, expectedStatus, apiErr.StatusCode)
}

// errorTestCase defines a test case for API error responses.
type errorTestCase struct {
	name     string
	pattern  string
	status   int
	message  string
	code     int
	badToken bool
	call     func(context.Context, *Client) error
}

// runErrorTests runs a set of error test cases against API methods.
func runErrorTests(t *testing.T, tests []errorTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mux, client := newTestMux(t)
			mux.HandleFunc(tt.pattern, func(w http.ResponseWriter, r *http.Request) {
				if tt.badToken {
					if !requireBearerToken(t, w, r, "test-token") {
						return
					}
				}
				writeError(w, tt.status, tt.message, tt.code)
			})
			if tt.badToken {
				client.Token = "bad-token"
			}
			err := tt.call(context.Background(), client)
			assertAPIError(t, err, tt.status)
		})
	}
}

// testContextCanceled tests that a canceled context returns an error.
func testContextCanceled(t *testing.T, pattern string, okResp interface{}, call func(context.Context, *Client) error) {
	t.Helper()
	t.Parallel()
	mux, client := newTestMux(t)
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, okResp)
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := call(ctx, client)
	require.Error(t, err)
}

// integrationClient returns a Client configured for integration testing.
// It skips the test if SKIP_INTEGRATION is set, and fails if ABLY_ACCOUNT_TOKEN is not set.
func integrationClient(t *testing.T) *Client {
	t.Helper()
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("SKIP_INTEGRATION set, skipping integration test")
	}
	token := os.Getenv("ABLY_ACCOUNT_TOKEN")
	if token == "" {
		t.Skip("ABLY_ACCOUNT_TOKEN not set (set SKIP_INTEGRATION to skip integration tests)")
	}
	c := NewClient(token)
	if url := os.Getenv("ABLY_URL"); url != "" {
		c.BaseURL = url
	}
	return c
}

// badTokenClient returns an integration Client with an invalid token.
func badTokenClient(t *testing.T) *Client {
	t.Helper()
	c := integrationClient(t)
	c.Token = "invalid-token-that-should-not-work"
	return c
}

// createTestApp creates a temporary app for integration testing and registers cleanup.
func createTestApp(t *testing.T, client *Client, name string) string {
	t.Helper()
	me, err := client.Me(context.Background())
	require.NoError(t, err)
	app, err := client.CreateApp(context.Background(), me.Account.ID, AppPost{Name: name})
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := client.DeleteApp(ctx, app.ID); err != nil {
			t.Errorf("cleanup: failed to delete test app %s: %v", app.ID, err)
		}
	})
	return app.ID
}
