package ably

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GetAccountStats retrieves stats for an account.
func (c *Client) GetAccountStats(ctx context.Context, accountID string, params *StatsParams) ([]StatsResponse, error) {
	path := addStatsParams(fmt.Sprintf("accounts/%s/stats", url.PathEscape(accountID)), params)
	var result []StatsResponse
	err := c.doJSON(ctx, http.MethodGet, path, nil, &result)
	return result, err
}
