package control

import (
	"context"
	"net/http"
)

// Me returns information about the authenticated token, including the
// associated user and account. Useful for discovering the account ID
// needed by other methods.
func (c *Client) Me(ctx context.Context) (Me, error) {
	var result Me
	err := c.doJSON(ctx, http.MethodGet, "me", nil, &result)
	return result, err
}
