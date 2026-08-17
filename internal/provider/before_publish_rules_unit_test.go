// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// sampleBeforePublishConfig returns a fully-populated before-publish config for
// use in the tests below.
func sampleBeforePublishConfig() *AblyBeforePublishConfig {
	return &AblyBeforePublishConfig{
		RetryTimeout:          types.Int64Value(5000),
		MaxRetries:            types.Int64Value(3),
		FailedAction:          types.StringValue("PUBLISH"),
		TooManyRequestsAction: types.StringValue("RETRY"),
	}
}

// sampleThresholds returns a thresholds map as Terraform holds it.
func sampleThresholds(t *testing.T) types.Map {
	t.Helper()
	m, diags := types.MapValueFrom(context.Background(), types.Int64Type, map[string]int64{"abuse": 2})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics building thresholds: %s", diags.Errors()[0].Detail())
	}
	return m
}

// wireBody marshals a create body and returns its top-level fields, so the tests
// can assert the exact tokens the Control API accepts. The echoing fake cannot
// catch a wrong enum or a stray webhook field; this can.
func wireBody(t *testing.T, body any) map[string]json.RawMessage {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	return raw
}

// assertNoWebhookFields guards the keystone footgun for this family: the
// dominant webhook pattern bakes in source/requestMode, which the moderation and
// before-publish APIs reject.
func assertNoWebhookFields(t *testing.T, raw map[string]json.RawMessage) {
	t.Helper()
	if _, ok := raw["requestMode"]; ok {
		t.Error("body must not contain a requestMode field")
	}
	if _, ok := raw["source"]; ok {
		t.Error("body must not contain a source field")
	}
}

// TestModerationRulePosts_Discriminators verifies each moderation rule sends its
// own ruleType discriminator with the moderation-shaped fields and none of the
// webhook ones.
func TestModerationRulePosts_Discriminators(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	thresholds := sampleThresholds(t)

	tisane, diags := getPlanTisanePost(ctx, AblyRuleTisane{
		Status:              types.StringValue("enabled"),
		InvocationMode:      types.StringValue("BEFORE_PUBLISH"),
		ChatRoomFilter:      types.StringValue("/room-.*/"),
		BeforePublishConfig: sampleBeforePublishConfig(),
		Target: &AblyRuleTisaneTarget{
			ApiKey:          types.StringValue("secret-key"),
			ModelURL:        types.StringNull(),
			Thresholds:      thresholds,
			DefaultLanguage: types.StringValue("en"),
		},
	})
	if diags.HasError() {
		t.Fatalf("tisane post diagnostics: %s", diags.Errors()[0].Detail())
	}

	azure, diags := getPlanAzureModerationPost(ctx, AblyRuleAzureModeration{
		Status:              types.StringValue("enabled"),
		InvocationMode:      types.StringValue("BEFORE_PUBLISH"),
		BeforePublishConfig: sampleBeforePublishConfig(),
		Target: &AblyRuleAzureModerationTarget{
			ApiKey:     types.StringValue("secret-key"),
			Endpoint:   types.StringValue("https://my-resource.cognitiveservices.azure.com"),
			Thresholds: thresholds,
		},
	})
	if diags.HasError() {
		t.Fatalf("azure post diagnostics: %s", diags.Errors()[0].Detail())
	}

	hiveText, diags := getPlanHiveTextPost(ctx, AblyRuleHiveText{
		Status:              types.StringValue("enabled"),
		InvocationMode:      types.StringValue("BEFORE_PUBLISH"),
		BeforePublishConfig: sampleBeforePublishConfig(),
		Target: &AblyRuleHiveTextTarget{
			ApiKey:     types.StringValue("secret-key"),
			ModelURL:   types.StringNull(),
			Thresholds: thresholds,
		},
	})
	if diags.HasError() {
		t.Fatalf("hive text post diagnostics: %s", diags.Errors()[0].Detail())
	}

	// Hive dashboard is the one rule in the family that runs after publish.
	hiveDashboard, diags := getPlanHiveDashboardPost(ctx, AblyRuleHiveDashboard{
		Status:         types.StringValue("enabled"),
		InvocationMode: types.StringValue("AFTER_PUBLISH"),
		Target: &AblyRuleHiveDashboardTarget{
			ApiKey:          types.StringValue("secret-key"),
			CheckWatchLists: types.BoolValue(true),
		},
	})
	if diags.HasError() {
		t.Fatalf("hive dashboard post diagnostics: %s", diags.Errors()[0].Detail())
	}

	webhook, diags := getPlanBeforePublishWebhookPost(ctx, AblyRuleBeforePublishWebhook{
		Status:              types.StringValue("enabled"),
		InvocationMode:      types.StringValue("BEFORE_PUBLISH"),
		BeforePublishConfig: sampleBeforePublishConfig(),
		Target: &AblyRuleBeforePublishWebhookTarget{
			Url: types.StringValue("https://example.com/moderate"),
			Headers: []AblyRuleHeaders{{
				Name:  types.StringValue("X-Custom-Header"),
				Value: types.StringValue("custom-header-value"),
			}},
		},
	})
	if diags.HasError() {
		t.Fatalf("before-publish webhook post diagnostics: %s", diags.Errors()[0].Detail())
	}

	for _, tc := range []struct {
		name           string
		body           any
		ruleType       string
		invocationMode string
	}{
		{"tisane", tisane, "tisane/text-moderation", "BEFORE_PUBLISH"},
		{"azure", azure, "azure/text-moderation", "BEFORE_PUBLISH"},
		{"hive text", hiveText, "hive/text-model-only", "BEFORE_PUBLISH"},
		{"hive dashboard", hiveDashboard, "hive/dashboard", "AFTER_PUBLISH"},
		{"before-publish webhook", webhook, "http/before-publish", "BEFORE_PUBLISH"},
	} {
		raw := wireBody(t, tc.body)
		if got := string(raw["ruleType"]); got != `"`+tc.ruleType+`"` {
			t.Errorf("%s ruleType wire value = %s, want %q", tc.name, got, tc.ruleType)
		}
		if got := string(raw["invocationMode"]); got != `"`+tc.invocationMode+`"` {
			t.Errorf("%s invocationMode wire value = %s, want %q", tc.name, got, tc.invocationMode)
		}
		assertNoWebhookFields(t, raw)
	}

	// Hive dashboard is the one rule in the family with no before-publish
	// config; sending one would be a copied-family bug.
	if _, ok := wireBody(t, hiveDashboard)["beforePublishConfig"]; ok {
		t.Error("hive/dashboard body must not contain beforePublishConfig")
	}
	// The other four must send it, with the exact enum casing.
	for name, body := range map[string]any{"tisane": tisane, "azure": azure, "hive text": hiveText, "webhook": webhook} {
		var bpc struct {
			FailedAction          string `json:"failedAction"`
			TooManyRequestsAction string `json:"tooManyRequestsAction"`
			RetryTimeout          int    `json:"retryTimeout"`
		}
		raw, ok := wireBody(t, body)["beforePublishConfig"]
		if !ok {
			t.Errorf("%s body is missing beforePublishConfig", name)
			continue
		}
		if err := json.Unmarshal(raw, &bpc); err != nil {
			t.Fatalf("%s beforePublishConfig unmarshal error: %v", name, err)
		}
		if bpc.FailedAction != "PUBLISH" || bpc.TooManyRequestsAction != "RETRY" || bpc.RetryTimeout != 5000 {
			t.Errorf("%s beforePublishConfig = %+v, want PUBLISH/RETRY/5000", name, bpc)
		}
	}
}

// TestThresholdsRoundTrip verifies the thresholds map survives the trip to the
// control type and back, and that an absent map is null rather than empty (an
// empty map in state against a null plan is a permanent diff).
func TestThresholdsRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	thresholds, diags := thresholdsPost(ctx, sampleThresholds(t))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	if thresholds["abuse"] != 2 {
		t.Fatalf("thresholds = %v, want abuse=2", thresholds)
	}

	back, diags := thresholdsResponse(ctx, thresholds)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	if len(back.Elements()) != 1 {
		t.Fatalf("mapped thresholds = %v, want one element", back)
	}

	nilThresholds, diags := thresholdsPost(ctx, types.MapNull(types.Int64Type))
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	if nilThresholds != nil {
		t.Fatalf("null thresholds must not serialize, got %v", nilThresholds)
	}

	nullBack, diags := thresholdsResponse(ctx, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	if !nullBack.IsNull() {
		t.Fatalf("absent thresholds must map to null, got %v", nullBack)
	}
}

// TestGetTisaneResponse_AbsentOptionalsAreNull verifies a response without
// chatRoomFilter, modelUrl or thresholds maps them to null, matching a plan that
// never set them.
func TestGetTisaneResponse_AbsentOptionalsAreNull(t *testing.T) {
	t.Parallel()

	rule := control.RuleResponse{
		ID:             "rule-1",
		AppID:          "app-123",
		Status:         "enabled",
		RuleType:       "tisane/text-moderation",
		InvocationMode: "BEFORE_PUBLISH",
		BeforePublishConfig: &control.BeforePublishConfig{
			RetryTimeout: 5000, MaxRetries: 3, FailedAction: "PUBLISH", TooManyRequestsAction: "RETRY",
		},
		Target: map[string]any{"apiKey": "resp-key", "defaultLanguage": "en"},
	}

	got, diags := getTisaneResponse(context.Background(), &rule, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	if !got.ChatRoomFilter.IsNull() {
		t.Errorf("chat_room_filter = %q, want null", got.ChatRoomFilter.ValueString())
	}
	if !got.Target.ModelURL.IsNull() {
		t.Errorf("model_url = %q, want null", got.Target.ModelURL.ValueString())
	}
	if !got.Target.Thresholds.IsNull() {
		t.Errorf("thresholds = %v, want null", got.Target.Thresholds)
	}
	if got.Target.ApiKey.ValueString() != "resp-key" {
		t.Errorf("api_key = %q, want resp-key", got.Target.ApiKey.ValueString())
	}
	if got.BeforePublishConfig == nil || got.BeforePublishConfig.MaxRetries.ValueInt64() != 3 {
		t.Errorf("before_publish_config = %+v, want max_retries 3 from the response", got.BeforePublishConfig)
	}
}

// TestGetHiveTextResponse_WrongRuleType ensures a mismatched discriminator in the
// response is surfaced as an error rather than silently mis-mapped, so a rule of
// another family imported by ID does not land in this resource's state.
func TestGetHiveTextResponse_WrongRuleType(t *testing.T) {
	t.Parallel()

	rule := control.RuleResponse{ID: "rule-1", RuleType: "http", Target: map[string]any{}}

	if _, diags := getHiveTextResponse(context.Background(), &rule, nil); !diags.HasError() {
		t.Fatal("expected an error for a non-hive rule type, got none")
	}
}

// TestGetPlanBeforePublishLambdaPost_AuthModes verifies each AWS auth mode sends
// only its own fields: the API rejects credentials alongside an assume-role ARN.
func TestGetPlanBeforePublishLambdaPost_AuthModes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	credentials, diags := getPlanBeforePublishLambdaPost(ctx, sampleBeforePublishLambdaPlan())
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	raw := wireBody(t, credentials)
	if got := string(raw["ruleType"]); got != `"aws/lambda/before-publish"` {
		t.Errorf("ruleType wire value = %s, want \"aws/lambda/before-publish\"", got)
	}
	if _, ok := raw["requestMode"]; ok {
		t.Error("body must not contain a requestMode field")
	}
	// This is the one before-publish family that does take a source.
	if _, ok := raw["source"]; !ok {
		t.Error("body must carry the source when the plan sets one")
	}

	var target struct {
		Authentication control.AWSAuthentication `json:"authentication"`
	}
	if err := json.Unmarshal(raw["target"], &target); err != nil {
		t.Fatalf("target unmarshal error: %v", err)
	}
	if target.Authentication.AuthenticationMode != "credentials" {
		t.Errorf("authenticationMode = %q, want credentials", target.Authentication.AuthenticationMode)
	}
	if target.Authentication.SecretAccessKey != "secret" {
		t.Errorf("secretAccessKey = %q, want secret", target.Authentication.SecretAccessKey)
	}
	if target.Authentication.AssumeRoleArn != "" {
		t.Errorf("assumeRoleArn = %q, want empty in credentials mode", target.Authentication.AssumeRoleArn)
	}

	plan := sampleBeforePublishLambdaPlan()
	plan.Source = nil
	plan.Target.Authentication = &AblyRuleBeforePublishLambdaAuth{
		AuthenticationMode: types.StringValue("assumeRole"),
		AssumeRoleArn:      types.StringValue("arn:aws:iam::123456789012:role/ably-moderation"),
	}
	assumeRole, diags := getPlanBeforePublishLambdaPost(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	raw = wireBody(t, assumeRole)
	if _, ok := raw["source"]; ok {
		t.Error("body must omit the source when the plan has none")
	}
	// Reset before decoding: json.Unmarshal leaves fields absent from the body
	// untouched, which would carry the credentials over from the assertions above.
	target.Authentication = control.AWSAuthentication{}
	if err := json.Unmarshal(raw["target"], &target); err != nil {
		t.Fatalf("target unmarshal error: %v", err)
	}
	if target.Authentication.AuthenticationMode != "assumeRole" {
		t.Errorf("authenticationMode = %q, want assumeRole", target.Authentication.AuthenticationMode)
	}
	if target.Authentication.AccessKeyID != "" || target.Authentication.SecretAccessKey != "" {
		t.Errorf("credentials leaked into assumeRole body: %+v", target.Authentication)
	}
}

// TestGetBeforePublishLambdaResponse_PreservesSecret is the loud test for the
// write-only carve-out: the Control API never returns secretAccessKey, so it has
// to come from the plan (create/update) or prior state (read). Reading it back as
// null aborts the apply with "inconsistent result after apply" on a sensitive
// attribute, which is exactly the failure mode this guards.
func TestGetBeforePublishLambdaResponse_PreservesSecret(t *testing.T) {
	t.Parallel()

	prior := sampleBeforePublishLambdaPlan()
	rule := control.RuleResponse{
		ID:             "rule-1",
		AppID:          "app-123",
		Status:         "enabled",
		RuleType:       "aws/lambda/before-publish",
		InvocationMode: "BEFORE_PUBLISH",
		BeforePublishConfig: &control.BeforePublishConfig{
			RetryTimeout: 5000, MaxRetries: 3, FailedAction: "PUBLISH", TooManyRequestsAction: "RETRY",
		},
		Target: map[string]any{
			"region":       "us-west-1",
			"functionName": "my-moderation-function",
			// As the API returns it: the secret is absent.
			"authentication": map[string]any{
				"authenticationMode": "credentials",
				"accessKeyId":        "AKIAIOSFODNN7EXAMPLE",
			},
		},
	}

	got, diags := getBeforePublishLambdaResponse(context.Background(), &rule, &prior)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	if got.Target.Authentication.SecretAccessKey.ValueString() != "secret" {
		t.Fatalf("secret_access_key = %q, want the value preserved from the plan", got.Target.Authentication.SecretAccessKey.ValueString())
	}
	if got.Target.Authentication.AccessKeyID.ValueString() != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("access_key_id = %q, want it read back from the response", got.Target.Authentication.AccessKeyID.ValueString())
	}
	if !got.Target.Authentication.AssumeRoleArn.IsNull() {
		t.Errorf("assume_role_arn = %q, want null in credentials mode", got.Target.Authentication.AssumeRoleArn.ValueString())
	}

	// Switching to assumeRole must clear the credentials rather than leaving the
	// plan's values behind in state.
	rule.Target = map[string]any{
		"region":       "us-west-1",
		"functionName": "my-moderation-function",
		"authentication": map[string]any{
			"authenticationMode": "assumeRole",
			"assumeRoleArn":      "arn:aws:iam::123456789012:role/ably-moderation",
		},
	}
	got, diags = getBeforePublishLambdaResponse(context.Background(), &rule, &prior)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors()[0].Detail())
	}
	if !got.Target.Authentication.SecretAccessKey.IsNull() || !got.Target.Authentication.AccessKeyID.IsNull() {
		t.Fatalf("credentials survived a switch to assumeRole: %+v", got.Target.Authentication)
	}
	if got.Target.Authentication.AssumeRoleArn.ValueString() != "arn:aws:iam::123456789012:role/ably-moderation" {
		t.Errorf("assume_role_arn = %q, want it read back from the response", got.Target.Authentication.AssumeRoleArn.ValueString())
	}
}

// sampleBeforePublishLambdaPlan returns a credentials-mode Lambda plan.
func sampleBeforePublishLambdaPlan() AblyRuleBeforePublishLambda {
	return AblyRuleBeforePublishLambda{
		ID:                  types.StringValue("rule-1"),
		AppID:               types.StringValue("app-123"),
		Status:              types.StringValue("enabled"),
		InvocationMode:      types.StringValue("BEFORE_PUBLISH"),
		BeforePublishConfig: sampleBeforePublishConfig(),
		Source: &AblyRuleSource{
			ChannelFilter: types.StringValue("^room:"),
			Type:          types.StringValue("channel.message"),
		},
		Target: &AblyRuleBeforePublishLambdaTarget{
			Region:       types.StringValue("us-west-1"),
			FunctionName: types.StringValue("my-moderation-function"),
			Authentication: &AblyRuleBeforePublishLambdaAuth{
				AuthenticationMode: types.StringValue("credentials"),
				AccessKeyID:        types.StringValue("AKIAIOSFODNN7EXAMPLE"),
				SecretAccessKey:    types.StringValue("secret"),
			},
		},
	}
}
