// Package provider implements the Ably provider for Terraform
package provider

import (
	"encoding/json"
	"testing"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// samplePlan returns a fully-populated Bodyguard plan for use in tests.
func samplePlan() AblyRuleBodyguard {
	return AblyRuleBodyguard{
		AppID:          types.StringValue("app-123"),
		Status:         types.StringValue("enabled"),
		InvocationMode: types.StringValue("BEFORE_PUBLISH"),
		ChatRoomFilter: types.StringValue("/room-.*/"),
		BeforePublishConfig: &AblyRuleBodyguardBeforePublishConfig{
			RetryTimeout:          types.Int64Value(5000),
			MaxRetries:            types.Int64Value(3),
			FailedAction:          types.StringValue("PUBLISH"),
			TooManyRequestsAction: types.StringValue("RETRY"),
		},
		Target: &AblyRuleBodyguardTarget{
			ApiKey:          types.StringValue("secret-key"),
			ChannelID:       types.StringValue("my-channel"),
			ApiURL:          types.StringNull(),
			DefaultLanguage: types.StringValue("en"),
		},
	}
}

// TestGetPlanBodyguardPost_Discriminator verifies the create body uses the
// correct ruleType discriminator and carries the moderation-specific fields
// rather than the webhook source/request_mode fields.
func TestGetPlanBodyguardPost_Discriminator(t *testing.T) {
	t.Parallel()

	post := getPlanBodyguardPost(samplePlan())

	if post.RuleType != "bodyguard/text-moderation" {
		t.Fatalf("expected ruleType=bodyguard/text-moderation, got %q", post.RuleType)
	}
	if post.InvocationMode != "BEFORE_PUBLISH" {
		t.Fatalf("expected invocationMode=BEFORE_PUBLISH, got %q", post.InvocationMode)
	}
	if post.ChatRoomFilter != "/room-.*/" {
		t.Fatalf("expected chatRoomFilter=/room-.*/, got %q", post.ChatRoomFilter)
	}
	if post.Target.APIKey != "secret-key" {
		t.Fatalf("expected target apiKey=secret-key, got %q", post.Target.APIKey)
	}
	if post.BeforePublishConfig.RetryTimeout != 5000 || post.BeforePublishConfig.MaxRetries != 3 {
		t.Fatalf("unexpected before-publish config: %+v", post.BeforePublishConfig)
	}

	// Marshal the body and assert the exact wire tokens the Control API accepts
	// (per swagger.yaml: REJECT/PUBLISH, RETRY/FAIL, BEFORE_PUBLISH). This pins
	// the enum casing so the resource can't silently regress to invalid values,
	// which neither the echoing fake nor a key-presence check would catch. It
	// also guards that bodyguard rules never serialize source/requestMode (the
	// webhook shape).
	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if _, ok := raw["source"]; ok {
		t.Fatal("bodyguard rule body must not contain a source field")
	}
	if _, ok := raw["requestMode"]; ok {
		t.Fatal("bodyguard rule body must not contain a requestMode field")
	}
	if got := string(raw["invocationMode"]); got != `"BEFORE_PUBLISH"` {
		t.Fatalf("invocationMode wire value = %s, want \"BEFORE_PUBLISH\"", got)
	}
	var bpc struct {
		FailedAction          string `json:"failedAction"`
		TooManyRequestsAction string `json:"tooManyRequestsAction"`
	}
	if err := json.Unmarshal(raw["beforePublishConfig"], &bpc); err != nil {
		t.Fatalf("beforePublishConfig unmarshal error: %v", err)
	}
	if bpc.FailedAction != "PUBLISH" {
		t.Fatalf("failedAction wire value = %q, want PUBLISH", bpc.FailedAction)
	}
	if bpc.TooManyRequestsAction != "RETRY" {
		t.Fatalf("tooManyRequestsAction wire value = %q, want RETRY", bpc.TooManyRequestsAction)
	}
}

// sampleResponse returns a moderation rule response as the Control API shapes
// it: all moderation fields present, but no api_key in the target (it is
// write-only and never returned).
func sampleResponse() control.RuleResponse {
	return control.RuleResponse{
		ID:             "rule-1",
		AppID:          "app-123",
		Status:         "enabled",
		RuleType:       "bodyguard/text-moderation",
		InvocationMode: "BEFORE_PUBLISH",
		ChatRoomFilter: "/room-.*/",
		BeforePublishConfig: &control.BeforePublishConfig{
			RetryTimeout:          5000,
			MaxRetries:            3,
			FailedAction:          "PUBLISH",
			TooManyRequestsAction: "RETRY",
		},
		Target: map[string]any{
			"channelId":       "my-channel",
			"defaultLanguage": "en",
		},
	}
}

// TestGetBodyguardResponse_ReadsModerationFieldsBack verifies the moderation
// fields come from the response, not the plan, so out-of-band changes surface
// as drift. The write-only target api_key is the one exception: it is
// preserved from the plan because the API never returns it.
func TestGetBodyguardResponse_ReadsModerationFieldsBack(t *testing.T) {
	t.Parallel()

	plan := samplePlan()

	// The response disagrees with the plan on every moderation field: the
	// response must win everywhere except the write-only api_key.
	rule := sampleResponse()
	rule.ChatRoomFilter = "/other-.*/"
	rule.BeforePublishConfig.MaxRetries = 7

	got, diags := getBodyguardResponse(&rule, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}

	if got.Target.ApiKey.ValueString() != "secret-key" {
		t.Fatalf("expected write-only api_key to be preserved from plan, got %q", got.Target.ApiKey.ValueString())
	}
	if got.InvocationMode.ValueString() != "BEFORE_PUBLISH" {
		t.Fatalf("expected invocation_mode from response, got %q", got.InvocationMode.ValueString())
	}
	if got.ChatRoomFilter.ValueString() != "/other-.*/" {
		t.Fatalf("expected chat_room_filter from response, got %q", got.ChatRoomFilter.ValueString())
	}
	if got.BeforePublishConfig == nil {
		t.Fatal("expected before_publish_config from response, got nil")
	}
	if got.BeforePublishConfig.MaxRetries.ValueInt64() != 7 {
		t.Fatalf("expected before_publish_config.max_retries=7 from response, got %d", got.BeforePublishConfig.MaxRetries.ValueInt64())
	}
	if got.ID.ValueString() != "rule-1" {
		t.Fatalf("expected id from response, got %q", got.ID.ValueString())
	}
}

// TestGetBodyguardResponse_ClearedFilterIsNull verifies a response without
// chatRoomFilter maps to null even when the prior plan/state held a value:
// this is the read-back half of clearing the filter.
func TestGetBodyguardResponse_ClearedFilterIsNull(t *testing.T) {
	t.Parallel()

	plan := samplePlan()
	rule := sampleResponse()
	rule.ChatRoomFilter = ""

	got, diags := getBodyguardResponse(&rule, &plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	if !got.ChatRoomFilter.IsNull() {
		t.Fatalf("expected chat_room_filter null when absent from response, got %q", got.ChatRoomFilter.ValueString())
	}
}

// TestGetBodyguardResponse_ImportLeavesApiKeyNull verifies the import path
// (no plan): the write-only api_key must be null, not a known "", since the
// API never returns it and a known empty string would misrepresent it.
func TestGetBodyguardResponse_ImportLeavesApiKeyNull(t *testing.T) {
	t.Parallel()

	rule := sampleResponse()

	got, diags := getBodyguardResponse(&rule, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	if !got.Target.ApiKey.IsNull() {
		t.Fatalf("expected api_key null on import, got %q", got.Target.ApiKey.ValueString())
	}
	if got.ChatRoomFilter.ValueString() != "/room-.*/" {
		t.Fatalf("expected chat_room_filter from response on import, got %q", got.ChatRoomFilter.ValueString())
	}
	if got.BeforePublishConfig == nil {
		t.Fatal("expected before_publish_config from response on import, got nil")
	}
}

// TestGetBodyguardResponse_WrongRuleType ensures a mismatched discriminator in
// the response is surfaced as an error rather than silently mis-mapped.
func TestGetBodyguardResponse_WrongRuleType(t *testing.T) {
	t.Parallel()

	plan := samplePlan()
	rule := control.RuleResponse{
		ID:       "rule-1",
		RuleType: "http",
		Target:   map[string]any{},
	}

	_, diags := getBodyguardResponse(&rule, &plan)
	if !diags.HasError() {
		t.Fatal("expected an error for a non-bodyguard rule type, got none")
	}
}
