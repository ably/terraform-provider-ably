package ably

import (
	"context"
	"net/http"
)

// Me returns information about the current token, user, and account.
func (c *Client) Me(ctx context.Context) (Me, error) {
	var result Me
	err := c.doJSON(ctx, http.MethodGet, "me", nil, &result)
	return result, err
}
