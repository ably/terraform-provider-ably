package control

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListRules returns all integration rules for an app. The full list is
// returned in a single request (no pagination).
func (c *Client) ListRules(ctx context.Context, appID string) ([]RuleResponse, error) {
	var result []RuleResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/rules", url.PathEscape(appID)), nil, &result)
	return result, err
}

// CreateRule creates an integration rule. Pass any XxxRulePost struct
// as body (e.g. [HTTPRulePost], [AWSLambdaRulePost], [KafkaRulePost]).
// The body is serialized as JSON; set the RuleType field to match the
// struct type.
func (c *Client) CreateRule(ctx context.Context, appID string, body any) (RuleResponse, error) {
	var result RuleResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("apps/%s/rules", url.PathEscape(appID)), body, &result)
	return result, err
}

// GetRule retrieves a single rule by ID.
func (c *Client) GetRule(ctx context.Context, appID string, ruleID string) (RuleResponse, error) {
	var result RuleResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("apps/%s/rules/%s", url.PathEscape(appID), url.PathEscape(ruleID)), nil, &result)
	return result, err
}

// UpdateRule performs a partial update (PATCH) on a rule. Pass any
// XxxRulePost or XxxRulePatch struct as body. The RuleType field must
// be set and match the existing rule's type.
func (c *Client) UpdateRule(ctx context.Context, appID string, ruleID string, body any) (RuleResponse, error) {
	var result RuleResponse
	err := c.doJSON(ctx, http.MethodPatch, fmt.Sprintf("apps/%s/rules/%s", url.PathEscape(appID), url.PathEscape(ruleID)), body, &result)
	return result, err
}

// DeleteRule deletes a rule. Returns [*Error] with StatusCode 404 if
// the rule does not exist.
func (c *Client) DeleteRule(ctx context.Context, appID string, ruleID string) error {
	return c.doJSON(ctx, http.MethodDelete, fmt.Sprintf("apps/%s/rules/%s", url.PathEscape(appID), url.PathEscape(ruleID)), nil, nil)
}
