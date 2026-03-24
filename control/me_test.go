package control

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Me (GET /me)
// ---------------------------------------------------------------------------

func TestMe_Success(t *testing.T) {
	t.Parallel()
	mux, client := newTestMux(t)

	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		if !requireBearerToken(t, w, r, "test-token") {
			return
		}
		writeJSON(w, http.StatusOK, Me{
			Token: &MeToken{
				ID:           "tok-123",
				Name:         "My Token",
				Capabilities: []string{"read", "write"},
			},
			User: &MeUser{
				ID:    42,
				Email: "user@example.com",
			},
			Account: &MeAccount{
				ID:   "acc-abc",
				Name: "Test Account",
			},
		})
	})

	me, err := client.Me(context.Background())
	require.NoError(t, err)

	require.NotNil(t, me.Token)
	assert.Equal(t, "tok-123", me.Token.ID)
	assert.Equal(t, "My Token", me.Token.Name)
	assert.Equal(t, []string{"read", "write"}, me.Token.Capabilities)

	require.NotNil(t, me.User)
	assert.Equal(t, 42, me.User.ID)
	assert.Equal(t, "user@example.com", me.User.Email)

	require.NotNil(t, me.Account)
	assert.Equal(t, "acc-abc", me.Account.ID)
	assert.Equal(t, "Test Account", me.Account.Name)
}

func TestMe_Errors(t *testing.T) {
	t.Parallel()
	call := func(ctx context.Context, c *Client) error {
		_, err := c.Me(ctx)
		return err
	}
	runErrorTests(t, []errorTestCase{
		{name: "AuthFailure", pattern: "GET /me", status: 401, message: "Authentication failed", code: 40100, badToken: true, call: call},
		{name: "ServerError", pattern: "GET /me", status: 500, message: "Internal server error", code: 50000, call: call},
	})
}

func TestMe_ContextCancelled(t *testing.T) {
	testContextCanceled(t, "GET /me", Me{}, func(ctx context.Context, c *Client) error {
		_, err := c.Me(ctx)
		return err
	})
}

// ---------------------------------------------------------------------------
// Integration: Me
// ---------------------------------------------------------------------------

func TestIntegration_Me(t *testing.T) {
	client := integrationClient(t)

	me, err := client.Me(context.Background())
	require.NoError(t, err)
	require.NotNil(t, me.Token)
	assert.NotEmpty(t, me.Token.ID)
	require.NotNil(t, me.Account)
	assert.NotEmpty(t, me.Account.ID)
}

func TestIntegration_Me_BadToken(t *testing.T) {
	client := badTokenClient(t)

	_, err := client.Me(context.Background())
	assertAPIError(t, err, 401)
}
