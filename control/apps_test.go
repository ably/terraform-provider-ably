package control

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"io"
	"math/big"
	"mime"
	"mime/multipart"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"software.sslmate.com/src/go-pkcs12"
)

// ---------------------------------------------------------------------------
// ListApps
// ---------------------------------------------------------------------------

func TestListApps_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	want := []AppResponse{
		{ID: "app1", Name: "My App", AccountID: "acc123", Status: "enabled"},
		{ID: "app2", Name: "Other App", AccountID: "acc123", Status: "disabled"},
	}

	mux.HandleFunc("GET /accounts/acc123/apps", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		writeJSON(w, http.StatusOK, want)
	})

	got, err := client.ListApps(context.Background(), "acc123")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "app1", got[0].ID)
	assert.Equal(t, "My App", got[0].Name)
	assert.Equal(t, "app2", got[1].ID)
	assert.Equal(t, "Other App", got[1].Name)
}

func TestListApps_EmptyList(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("GET /accounts/acc123/apps", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		writeJSON(w, http.StatusOK, []AppResponse{})
	})

	got, err := client.ListApps(context.Background(), "acc123")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestListApps_Errors(t *testing.T) {
	t.Parallel()
	call := func(acct string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.ListApps(ctx, acct)
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "AuthFailure", pattern: "GET /accounts/acc123/apps", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("acc123")},
		{name: "NotFound", pattern: "GET /accounts/no-such-account/apps", status: 404, message: "Account not found", code: 40400, call: call("no-such-account")},
		{name: "ServerError", pattern: "GET /accounts/acc123/apps", status: 500, message: "Internal server error", code: 50000, call: call("acc123")},
	})
}

func TestListApps_ContextCancelled(t *testing.T) {
	testContextCanceled(t, "GET /accounts/acc123/apps", []AppResponse{}, func(ctx context.Context, c *Client) error {
		_, err := c.ListApps(ctx, "acc123")
		return err
	})
}

// ---------------------------------------------------------------------------
// CreateApp
// ---------------------------------------------------------------------------

func TestCreateApp_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("POST /accounts/acc123/apps", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}

		var body AppPost
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "New App", body.Name)

		writeJSON(w, http.StatusCreated, AppResponse{
			ID:        "app-new",
			AccountID: "acc123",
			Name:      body.Name,
			Status:    "enabled",
		})
	})

	got, err := client.CreateApp(context.Background(), "acc123", AppPost{Name: "New App"})
	require.NoError(t, err)
	assert.Equal(t, "app-new", got.ID)
	assert.Equal(t, "New App", got.Name)
	assert.Equal(t, "enabled", got.Status)
}

func TestCreateApp_Errors(t *testing.T) {
	t.Parallel()
	call := func(acct string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.CreateApp(ctx, acct, AppPost{Name: "X"})
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "AuthFailure", pattern: "POST /accounts/acc123/apps", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("acc123")},
		{name: "NotFound", pattern: "POST /accounts/no-such-account/apps", status: 404, message: "Account not found", code: 40400, call: call("no-such-account")},
		{name: "BadRequest", pattern: "POST /accounts/acc123/apps", status: 400, message: "Invalid JSON", code: 40000, call: call("acc123")},
		{name: "UnprocessableEntity", pattern: "POST /accounts/acc123/apps", status: 422, message: "Name already taken", code: 42200, call: call("acc123")},
		{name: "ServerError", pattern: "POST /accounts/acc123/apps", status: 500, message: "Internal server error", code: 50000, call: call("acc123")},
	})
}

// ---------------------------------------------------------------------------
// UpdateApp
// ---------------------------------------------------------------------------

func TestUpdateApp_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("PATCH /apps/app1", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}

		var body AppPatch
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "Renamed App", body.Name)

		writeJSON(w, http.StatusOK, AppResponse{
			ID:   "app1",
			Name: body.Name,
		})
	})

	got, err := client.UpdateApp(context.Background(), "app1", AppPatch{Name: "Renamed App"})
	require.NoError(t, err)
	assert.Equal(t, "app1", got.ID)
	assert.Equal(t, "Renamed App", got.Name)
}

func TestUpdateApp_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.UpdateApp(ctx, appID, AppPatch{Name: "X"})
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "AuthFailure", pattern: "PATCH /apps/app1", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app1")},
		{name: "NotFound", pattern: "PATCH /apps/no-such-app", status: 404, message: "App not found", code: 40400, call: call("no-such-app")},
		{name: "BadRequest", pattern: "PATCH /apps/app1", status: 400, message: "Invalid field value", code: 40000, call: call("app1")},
		{name: "UnprocessableEntity", pattern: "PATCH /apps/app1", status: 422, message: "Invalid resource", code: 42200, call: call("app1")},
		{name: "ServerError", pattern: "PATCH /apps/app1", status: 500, message: "Internal server error", code: 50000, call: call("app1")},
	})
}

// ---------------------------------------------------------------------------
// DeleteApp
// ---------------------------------------------------------------------------

func TestDeleteApp_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("DELETE /apps/app1", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteApp(context.Background(), "app1")
	require.NoError(t, err)
}

func TestDeleteApp_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			return c.DeleteApp(ctx, appID)
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "AuthFailure", pattern: "DELETE /apps/app1", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app1")},
		{name: "NotFound", pattern: "DELETE /apps/no-such-app", status: 404, message: "App not found", code: 40400, call: call("no-such-app")},
		{name: "UnprocessableEntity", pattern: "DELETE /apps/app1", status: 422, message: "Cannot delete app with active resources", code: 42200, call: call("app1")},
		{name: "ServerError", pattern: "DELETE /apps/app1", status: 500, message: "Internal server error", code: 50000, call: call("app1")},
	})
}

// ---------------------------------------------------------------------------
// UpdateAppPKCS12
// ---------------------------------------------------------------------------

func TestUpdateAppPKCS12_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	p12Data := []byte("fake-p12-binary-data")
	p12Pass := "secret"

	mux.HandleFunc("POST /apps/app1/pkcs12", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}

		contentType := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(contentType)
		require.NoError(t, err)
		assert.Equal(t, "multipart/form-data", mediaType)

		reader := multipart.NewReader(r.Body, params["boundary"])

		var gotFile []byte
		var gotPass string

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			switch part.FormName() {
			case "p12File":
				gotFile, err = io.ReadAll(part)
				require.NoError(t, err)
			case "p12Pass":
				data, err := io.ReadAll(part)
				require.NoError(t, err)
				gotPass = string(data)
			}
			part.Close()
		}

		assert.Equal(t, string(p12Data), string(gotFile))
		assert.Equal(t, p12Pass, gotPass)

		writeJSON(w, http.StatusOK, AppResponse{
			ID:   "app1",
			Name: "My App",
		})
	})

	got, err := client.UpdateAppPKCS12(context.Background(), "app1", p12Data, p12Pass)
	require.NoError(t, err)
	assert.Equal(t, "app1", got.ID)
	assert.Equal(t, "My App", got.Name)
}

func TestUpdateAppPKCS12_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.UpdateAppPKCS12(ctx, appID, []byte("data"), "pass")
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "AuthFailure", pattern: "POST /apps/app1/pkcs12", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app1")},
		{name: "NotFound", pattern: "POST /apps/no-such-app/pkcs12", status: 404, message: "App not found", code: 40400, call: call("no-such-app")},
		{name: "BadRequest", pattern: "POST /apps/app1/pkcs12", status: 400, message: "Invalid PKCS12 data", code: 40000, call: call("app1")},
		{name: "ServerError", pattern: "POST /apps/app1/pkcs12", status: 500, message: "Internal server error", code: 50000, call: call("app1")},
	})
}

// ---------------------------------------------------------------------------
// GetAppStats
// ---------------------------------------------------------------------------

func TestGetAppStats_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	want := []StatsResponse{
		{
			IntervalID: "2024-01-01:00:00",
			Unit:       "hour",
			Schema:     "v1",
			AppID:      "app1",
		},
	}

	mux.HandleFunc("GET /apps/app1/stats", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}

		q := r.URL.Query()
		assert.Equal(t, "1704067200", q.Get("start"))
		assert.Equal(t, "1704153600", q.Get("end"))
		assert.Equal(t, "hour", q.Get("unit"))
		assert.Equal(t, "forwards", q.Get("direction"))
		assert.Equal(t, "10", q.Get("limit"))

		writeJSON(w, http.StatusOK, want)
	})

	got, err := client.GetAppStats(context.Background(), "app1", &StatsParams{
		Start:     ptr(1704067200),
		End:       ptr(1704153600),
		Unit:      "hour",
		Direction: "forwards",
		Limit:     ptr(10),
	})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "2024-01-01:00:00", got[0].IntervalID)
	assert.Equal(t, "hour", got[0].Unit)
	assert.Equal(t, "app1", got[0].AppID)
}

func TestGetAppStats_NilParams(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("GET /apps/app1/stats", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}

		assert.Empty(t, r.URL.RawQuery)
		writeJSON(w, http.StatusOK, []StatsResponse{})
	})

	got, err := client.GetAppStats(context.Background(), "app1", nil)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetAppStats_Errors(t *testing.T) {
	t.Parallel()
	call := func(appID string) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			_, err := c.GetAppStats(ctx, appID, nil)
			return err
		}
	}
	runErrorTests(t, []errorTestCase{
		{name: "AuthFailure", pattern: "GET /apps/app1/stats", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call("app1")},
		{name: "NotFound", pattern: "GET /apps/no-such-app/stats", status: 404, message: "App not found", code: 40400, call: call("no-such-app")},
		{name: "BadRequest", pattern: "GET /apps/app1/stats", status: 400, message: "Invalid unit parameter", code: 40000, call: call("app1")},
		{name: "ServerError", pattern: "GET /apps/app1/stats", status: 500, message: "Internal server error", code: 50000, call: call("app1")},
	})
}

// ---------------------------------------------------------------------------
// Integration: Apps CRUD
// ---------------------------------------------------------------------------

func TestIntegration_AppsCRUD(t *testing.T) {
	client := integrationClient(t)
	suffix := testSuffix()

	me, err := client.Me(context.Background())
	require.NoError(t, err)
	accountID := me.Account.ID

	var appID string
	appName := "integ-app-" + suffix

	t.Run("Create", func(t *testing.T) {
		app, err := client.CreateApp(context.Background(), accountID, AppPost{
			Name: appName,
		})
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)
		appID = app.ID
		assert.Equal(t, appName, app.Name)
	})

	t.Cleanup(func() {
		if appID != "" {
			if err := client.DeleteApp(context.Background(), appID); err != nil {
				t.Errorf("DeleteApp cleanup failed for app %q: %v", appID, err)
			}
		}
	})

	t.Run("List", func(t *testing.T) {
		apps, err := client.ListApps(context.Background(), accountID)
		require.NoError(t, err)
		assert.True(t, slices.ContainsFunc(apps, func(a AppResponse) bool { return a.ID == appID }),
			"created app %s not found in list", appID)
	})

	t.Run("Update", func(t *testing.T) {
		updated, err := client.UpdateApp(context.Background(), appID, AppPatch{
			Name: appName + "-updated",
		})
		require.NoError(t, err)
		assert.Equal(t, appName+"-updated", updated.Name)
	})

	t.Run("Delete", func(t *testing.T) {
		err := client.DeleteApp(context.Background(), appID)
		require.NoError(t, err)
		appID = "" // prevent double-delete in cleanup
	})
}

func TestIntegration_Apps_NotFound(t *testing.T) {
	client := integrationClient(t)

	_, err := client.UpdateApp(context.Background(), "nonexistent-app-id", AppPatch{Name: "X"})
	assertAPIError(t, err, 404)

	err = client.DeleteApp(context.Background(), "nonexistent-app-id")
	assertAPIError(t, err, 404)
}

func TestIntegration_Apps_BadToken(t *testing.T) {
	client := badTokenClient(t)

	_, err := client.ListApps(context.Background(), "any-account-id")
	assertAPIError(t, err, 401)

	_, err = client.CreateApp(context.Background(), "any-account-id", AppPost{Name: "X"})
	assertAPIError(t, err, 401)

	_, err = client.UpdateApp(context.Background(), "any-app-id", AppPatch{Name: "X"})
	assertAPIError(t, err, 401)

	err = client.DeleteApp(context.Background(), "any-app-id")
	assertAPIError(t, err, 401)
}

// ---------------------------------------------------------------------------
// Integration: PKCS12
// ---------------------------------------------------------------------------

func generateSelfSignedPKCS12(t *testing.T, password string) []byte {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "integration-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	p12Data, err := pkcs12.Modern.Encode(key, cert, nil, password)
	require.NoError(t, err)

	return p12Data
}

func TestIntegration_PKCS12(t *testing.T) {
	client := integrationClient(t)
	appID := createTestApp(t, client, "integ-pkcs12-"+testSuffix())

	password := "test-password"
	p12Data := generateSelfSignedPKCS12(t, password)

	got, err := client.UpdateAppPKCS12(context.Background(), appID, p12Data, password)
	require.NoError(t, err)
	assert.Equal(t, appID, got.ID)
}

func TestIntegration_PKCS12_NotFound(t *testing.T) {
	client := integrationClient(t)
	p12Data := generateSelfSignedPKCS12(t, "pass")
	_, err := client.UpdateAppPKCS12(context.Background(), "nonexistent-app-id", p12Data, "pass")
	assertAPIError(t, err, 404)
}

func TestIntegration_PKCS12_BadToken(t *testing.T) {
	client := badTokenClient(t)
	_, err := client.UpdateAppPKCS12(context.Background(), "any-app-id", []byte("data"), "pass")
	assertAPIError(t, err, 401)
}

// ---------------------------------------------------------------------------
// Integration: App Stats
// ---------------------------------------------------------------------------

func TestIntegration_AppStats(t *testing.T) {
	client := integrationClient(t)
	appID := createTestApp(t, client, "integ-stats-"+testSuffix())

	_, err := client.GetAppStats(context.Background(), appID, &StatsParams{Unit: "hour", Limit: ptr(1)})
	require.NoError(t, err)
}

func TestIntegration_AppStats_NotFound(t *testing.T) {
	client := integrationClient(t)
	_, err := client.GetAppStats(context.Background(), "nonexistent-app-id", nil)
	assertAPIError(t, err, 404)
}

func TestIntegration_AppStats_BadToken(t *testing.T) {
	client := badTokenClient(t)
	_, err := client.GetAppStats(context.Background(), "any-app-id", nil)
	assertAPIError(t, err, 401)
}
