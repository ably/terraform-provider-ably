package provider

import (
	"encoding/json"
	"testing"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestWebhookEnveloped_SingleMode(t *testing.T) {
	t.Parallel()

	got := webhookEnveloped(types.BoolValue(true), "single")
	if got == nil || *got != true {
		t.Fatalf("expected enveloped=true in single mode, got %v", got)
	}

	got = webhookEnveloped(types.BoolValue(false), "single")
	if got == nil || *got != false {
		t.Fatalf("expected enveloped=false in single mode, got %v", got)
	}
}

func TestWebhookEnveloped_BatchMode(t *testing.T) {
	t.Parallel()

	// Even when the user sets enveloped=true, batch mode forces false.
	got := webhookEnveloped(types.BoolValue(true), "batch")
	if got == nil || *got != false {
		t.Fatalf("expected enveloped=false in batch mode, got %v", got)
	}

	got = webhookEnveloped(types.BoolValue(false), "batch")
	if got == nil || *got != false {
		t.Fatalf("expected enveloped=false in batch mode, got %v", got)
	}
}

// TestGetPlanRule_HTTPWebhook_BatchOmitsEnveloped verifies that HTTP-type
// webhook rules send enveloped=false in batch mode, preventing API rejection
// when transitioning from single (enveloped=true) to batch.
func TestGetPlanRule_HTTPWebhook_BatchOmitsEnveloped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target any
	}{
		{
			name: "http",
			target: &AblyRuleTargetHTTP{
				Url:          types.StringValue("https://example.com"),
				Format:       types.StringValue("json"),
				Enveloped:    types.BoolValue(true),
				SigningKeyId: types.StringNull(),
			},
		},
		{
			name: "azure-function",
			target: &AblyRuleTargetAzureFunction{
				AzureAppID:        types.StringValue("demo"),
				AzureFunctionName: types.StringValue("func0"),
				Format:            types.StringValue("json"),
				Enveloped:         types.BoolValue(true),
				SigningKeyID:      types.StringNull(),
			},
		},
		{
			name: "google-cloud-function",
			target: &AblyRuleTargetGoogleFunction{
				Region:       types.StringValue("us-central1"),
				ProjectID:    types.StringValue("proj"),
				FunctionName: types.StringValue("func0"),
				Format:       types.StringValue("json"),
				Enveloped:    types.BoolValue(true),
				SigningKeyId: types.StringNull(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := AblyRule{
				Status:      types.StringValue("enabled"),
				RequestMode: types.StringValue("batch"),
				Source: &AblyRuleSource{
					ChannelFilter: types.StringValue("^test"),
					Type:          types.StringValue("channel.message"),
				},
				Target: tt.target,
			}

			result, diags := GetPlanRule(plan)
			if diags.HasError() {
				t.Fatalf("unexpected error: %s", diags.Errors()[0].Detail())
			}

			// Marshal to JSON and verify enveloped is false.
			data, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			var raw map[string]json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			var target map[string]json.RawMessage
			if err := json.Unmarshal(raw["target"], &target); err != nil {
				t.Fatalf("unmarshal target error: %v", err)
			}

			envelopedJSON, ok := target["enveloped"]
			if !ok {
				t.Fatal("enveloped field missing from target JSON")
			}
			if string(envelopedJSON) != "false" {
				t.Fatalf("expected enveloped=false in batch mode JSON, got %s", string(envelopedJSON))
			}
		})
	}
}

// TestGetPlanRule_HTTPWebhook_SinglePreservesEnveloped verifies that
// single mode preserves the user's enveloped setting.
func TestGetPlanRule_HTTPWebhook_SinglePreservesEnveloped(t *testing.T) {
	t.Parallel()

	plan := AblyRule{
		Status:      types.StringValue("enabled"),
		RequestMode: types.StringValue("single"),
		Source: &AblyRuleSource{
			ChannelFilter: types.StringValue("^test"),
			Type:          types.StringValue("channel.message"),
		},
		Target: &AblyRuleTargetHTTP{
			Url:          types.StringValue("https://example.com"),
			Format:       types.StringValue("json"),
			Enveloped:    types.BoolValue(true),
			SigningKeyId: types.StringNull(),
		},
	}

	result, diags := GetPlanRule(plan)
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags.Errors()[0].Detail())
	}

	post := result.(control.HTTPRulePost)
	if post.Target.Enveloped == nil || *post.Target.Enveloped != true {
		t.Fatalf("expected enveloped=true in single mode, got %v", post.Target.Enveloped)
	}
}

// TestGetPlanRule_NonWebhook_BatchKeepsEnveloped verifies that non-webhook
// rule types (AMQP, Kafka, etc.) are unaffected by the batch+enveloped logic.
func TestGetPlanRule_NonWebhook_BatchKeepsEnveloped(t *testing.T) {
	t.Parallel()

	plan := AblyRule{
		Status:      types.StringValue("enabled"),
		RequestMode: types.StringValue("batch"),
		Source: &AblyRuleSource{
			ChannelFilter: types.StringValue("^test"),
			Type:          types.StringValue("channel.message"),
		},
		Target: &AblyRuleTargetAMQP{
			QueueID:   types.StringValue("queue-123"),
			Format:    types.StringValue("json"),
			Enveloped: types.BoolValue(true),
		},
	}

	result, diags := GetPlanRule(plan)
	if diags.HasError() {
		t.Fatalf("unexpected error: %s", diags.Errors()[0].Detail())
	}

	post := result.(control.AMQPRulePost)
	if post.Target.Enveloped == nil || *post.Target.Enveloped != true {
		t.Fatalf("non-webhook rule should keep enveloped=true in batch mode, got %v", post.Target.Enveloped)
	}
}

// TestGetAwsAuth_ImportWithNilPlan verifies that GetAwsAuth produces a
// valid authentication mode even when the plan target is nil (during import).
func TestGetAwsAuth_ImportWithNilPlan(t *testing.T) {
	t.Parallel()

	auth := control.AWSAuthentication{
		AuthenticationMode: "credentials",
		AccessKeyID:        "AKID",
	}

	// Simulate import: plan target is a typed nil.
	plan := &AblyRule{
		Target: (*AblyRuleTargetLambda)(nil),
	}

	result := GetAwsAuth(auth, plan)

	if result.AuthenticationMode.ValueString() != "credentials" {
		t.Fatalf("expected mode=credentials, got %q", result.AuthenticationMode.ValueString())
	}
	if result.AccessKeyId.ValueString() != "AKID" {
		t.Fatalf("expected access_key_id=AKID, got %q", result.AccessKeyId.ValueString())
	}
}

func TestGetAwsAuth_AssumeRoleImport(t *testing.T) {
	t.Parallel()

	auth := control.AWSAuthentication{
		AuthenticationMode: "assumeRole",
		AssumeRoleArn:      "arn:aws:iam::123:role/test",
	}

	plan := &AblyRule{
		Target: (*AblyRuleTargetLambda)(nil),
	}

	result := GetAwsAuth(auth, plan)

	if result.AuthenticationMode.ValueString() != "assumeRole" {
		t.Fatalf("expected mode=assumeRole, got %q", result.AuthenticationMode.ValueString())
	}
	if result.RoleArn.ValueString() != "arn:aws:iam::123:role/test" {
		t.Fatalf("expected role_arn, got %q", result.RoleArn.ValueString())
	}
	if !result.AccessKeyId.IsNull() {
		t.Fatal("expected access_key_id to be null for assumeRole")
	}
}
