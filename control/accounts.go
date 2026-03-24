package control

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// GetAccountStats returns usage statistics for an account. Pass nil for
// params to use API defaults.
func (c *Client) GetAccountStats(ctx context.Context, accountID string, params *StatsParams) ([]StatsResponse, error) {
	path := addStatsParams(fmt.Sprintf("accounts/%s/stats", url.PathEscape(accountID)), params)
	var result []StatsResponse
	err := c.doJSON(ctx, http.MethodGet, path, nil, &result)
	return result, err
}
