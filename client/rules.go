package ably

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListRules lists all rules for an app.
func (c *Client) ListRules(ctx context.Context, appID string) ([]RuleResponse, error) {
	var result []RuleResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/rules", url.PathEscape(appID)), nil, &result)
	return result, err
}

// CreateRule creates a new rule for an app.
func (c *Client) CreateRule(ctx context.Context, appID string, body any) (RuleResponse, error) {
	var result RuleResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/rules", url.PathEscape(appID)), body, &result)
	return result, err
}

// GetRule retrieves a single rule.
func (c *Client) GetRule(ctx context.Context, appID string, ruleID string) (RuleResponse, error) {
	var result RuleResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/rules/%s", url.PathEscape(appID), url.PathEscape(ruleID)), nil, &result)
	return result, err
}

// UpdateRule updates an existing rule.
func (c *Client) UpdateRule(ctx context.Context, appID string, ruleID string, body any) (RuleResponse, error) {
	var result RuleResponse
	err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("apps/%s/rules/%s", url.PathEscape(appID), url.PathEscape(ruleID)), body, &result)
	return result, err
}

// DeleteRule deletes a rule.
func (c *Client) DeleteRule(ctx context.Context, appID string, ruleID string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("apps/%s/rules/%s", url.PathEscape(appID), url.PathEscape(ruleID)), nil, nil)
}
