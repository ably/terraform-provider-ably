package ably

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Retry behavior
// ---------------------------------------------------------------------------

func TestRetry_5xxIsRetried(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n < 3 {
			writeError(w, http.StatusServiceUnavailable, "try again", 50300)
			return
		}
		writeJSON(w, http.StatusOK, Me{
			Token: &MeToken{ID: "tok-1", Name: "ok"},
		})
	})

	_, client := newTestServer(t, mux)
	client.HTTPClient.RetryMax = 4
	client.HTTPClient.RetryWaitMin = 0
	client.HTTPClient.RetryWaitMax = 0

	me, err := client.Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "tok-1", me.Token.ID)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestRetry_4xxIsNotRetried(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		writeError(w, http.StatusUnauthorized, "bad token", 40100)
	})

	_, client := newTestServer(t, mux)
	client.HTTPClient.RetryMax = 4
	client.HTTPClient.RetryWaitMin = 0
	client.HTTPClient.RetryWaitMax = 0

	_, err := client.Me(context.Background())
	assertAPIError(t, err, http.StatusUnauthorized)
	assert.Equal(t, int32(1), attempts.Load())
}

func TestRetry_ExhaustedReturnsLastError(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		writeError(w, http.StatusBadGateway, "upstream down", 50200)
	})

	_, client := newTestServer(t, mux)
	client.HTTPClient.RetryMax = 2
	client.HTTPClient.RetryWaitMin = 0
	client.HTTPClient.RetryWaitMax = 0

	_, err := client.Me(context.Background())
	assertAPIError(t, err, http.StatusBadGateway)
	assert.Equal(t, int32(3), attempts.Load()) // 1 initial + 2 retries
}

func TestRetry_DisabledWithZeroRetryMax(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		writeError(w, http.StatusInternalServerError, "error", 50000)
	})

	_, client := newTestServer(t, mux)
	client.HTTPClient.RetryMax = 0

	_, err := client.Me(context.Background())
	assertAPIError(t, err, http.StatusInternalServerError)
	assert.Equal(t, int32(1), attempts.Load())
}

func TestRetry_WithRetryMaxOption(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		writeError(w, http.StatusInternalServerError, "error", 50000)
	})

	srv := newTestServerRaw(t, mux)
	client := NewClient("test-token", WithRetryMax(1))
	client.BaseURL = srv.URL
	client.HTTPClient.RetryWaitMin = 0
	client.HTTPClient.RetryWaitMax = 0

	_, err := client.Me(context.Background())
	assertAPIError(t, err, http.StatusInternalServerError)
	assert.Equal(t, int32(2), attempts.Load()) // 1 initial + 1 retry
}

// ---------------------------------------------------------------------------
// User-Agent
// ---------------------------------------------------------------------------

func TestUserAgent_Default(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "ably-control-api/"+Version, r.Header.Get("User-Agent"))
		writeJSON(w, http.StatusOK, Me{})
	})

	_, client := newTestServer(t, mux)
	_, err := client.Me(context.Background())
	require.NoError(t, err)
}

func TestUserAgent_CustomAppend(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "ably-control-api/"+Version+" ably-terraform/2.0.0", r.Header.Get("User-Agent"))
		writeJSON(w, http.StatusOK, Me{})
	})

	_, client := newTestServer(t, mux)
	client.UserAgent += " ably-terraform/2.0.0"
	_, err := client.Me(context.Background())
	require.NoError(t, err)
}

func TestUserAgent_WithOption(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "custom-agent/1.0", r.Header.Get("User-Agent"))
		writeJSON(w, http.StatusOK, Me{})
	})

	srv := newTestServerRaw(t, mux)
	client := NewClient("test-token", WithUserAgent("custom-agent/1.0"))
	client.BaseURL = srv.URL
	client.HTTPClient.RetryMax = 0

	_, err := client.Me(context.Background())
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// WithHTTPClient option
// ---------------------------------------------------------------------------

func TestWithHTTPClient(t *testing.T) {
	t.Parallel()

	customTransport := &http.Transport{}
	customHTTPClient := &http.Client{Transport: customTransport}

	client := NewClient("test-token", WithHTTPClient(customHTTPClient))
	assert.Same(t, customHTTPClient, client.HTTPClient.HTTPClient)
}

// ---------------------------------------------------------------------------
// NewClient defaults
// ---------------------------------------------------------------------------

func TestNewClient_Defaults(t *testing.T) {
	t.Parallel()

	client := NewClient("my-token")
	assert.Equal(t, "my-token", client.Token)
	assert.Equal(t, "https://control.ably.net/v1", client.BaseURL)
	assert.Equal(t, "ably-control-api/"+Version, client.UserAgent)
	assert.Equal(t, 4, client.HTTPClient.RetryMax)
}

// ---------------------------------------------------------------------------
// Error type
// ---------------------------------------------------------------------------

func TestError_ErrorMethod(t *testing.T) {
	t.Parallel()

	err := &Error{Message: "something went wrong", Code: 40400, StatusCode: 404}
	assert.Equal(t, "something went wrong (code: 40400)", err.Error())
	// Verify it satisfies the error interface.
	var e error = err
	assert.Equal(t, "something went wrong (code: 40400)", e.Error())

	// When code is 0, only the message is returned.
	errNoCode := &Error{Message: "plain error"}
	assert.Equal(t, "plain error", errNoCode.Error())
}

func TestError_DetailsDeserialization(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"bad","code":40000,"statusCode":400,"details":{"field":"name","reason":"too long"}}`))
	})

	_, client := newTestServer(t, mux)
	_, err := client.Me(context.Background())
	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	assert.Equal(t, "bad", apiErr.Message)
	assert.NotNil(t, apiErr.Details)
	assert.Contains(t, string(apiErr.Details), "too long")
}

// ---------------------------------------------------------------------------
// checkResponse with malformed JSON
// ---------------------------------------------------------------------------

func TestCheckResponse_MalformedErrorBody(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("this is not json"))
	})

	_, client := newTestServer(t, mux)
	_, err := client.Me(context.Background())
	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadGateway, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "502")
}

// ---------------------------------------------------------------------------
// Content-Type header on JSON requests
// ---------------------------------------------------------------------------

func TestRequest_ContentTypeJSON(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /accounts/acc1/apps", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		writeJSON(w, http.StatusCreated, AppResponse{ID: "app1"})
	})

	_, client := newTestServer(t, mux)
	_, err := client.CreateApp(context.Background(), "acc1", AppPost{Name: "test"})
	require.NoError(t, err)
}

func TestRequest_NoContentTypeOnGET(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		// GET requests with no body should not have Content-Type set
		assert.Empty(t, r.Header.Get("Content-Type"))
		writeJSON(w, http.StatusOK, Me{})
	})

	_, client := newTestServer(t, mux)
	_, err := client.Me(context.Background())
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Context cancellation during retry
// ---------------------------------------------------------------------------

func TestRetry_ContextCancelledDuringRetry(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			cancel() // cancel after first attempt
		}
		writeError(w, http.StatusInternalServerError, "error", 50000)
	})

	_, client := newTestServer(t, mux)
	client.HTTPClient.RetryMax = 4
	client.HTTPClient.RetryWaitMin = 0
	client.HTTPClient.RetryWaitMax = 0

	_, err := client.Me(ctx)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// addStatsParams edge cases
// ---------------------------------------------------------------------------

func TestAddStatsParams_PartialParams(t *testing.T) {
	t.Parallel()

	t.Run("OnlyStart", func(t *testing.T) {
		got := addStatsParams("apps/x/stats", &StatsParams{Start: ptr(100)})
		assert.Contains(t, got, "start=100")
		assert.NotContains(t, got, "end=")
		assert.NotContains(t, got, "unit=")
	})

	t.Run("OnlyDirection", func(t *testing.T) {
		got := addStatsParams("apps/x/stats", &StatsParams{Direction: "backwards"})
		assert.Contains(t, got, "direction=backwards")
		assert.NotContains(t, got, "start=")
	})

	t.Run("EmptyParams", func(t *testing.T) {
		got := addStatsParams("apps/x/stats", &StatsParams{})
		assert.Equal(t, "apps/x/stats", got)
	})

	t.Run("NilParams", func(t *testing.T) {
		got := addStatsParams("apps/x/stats", nil)
		assert.Equal(t, "apps/x/stats", got)
	})
}
