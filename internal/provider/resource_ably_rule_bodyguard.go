// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/resource_rule_bodyguard"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// bodyguardRuleType is the Control API discriminator for Bodyguard text
// moderation rules.
const bodyguardRuleType = "bodyguard/text-moderation"

// AblyRuleBodyguardBeforePublishConfig mirrors control.BeforePublishConfig.
// Moderation rules run before a message is published, so this block controls
// retry/backoff and what happens when the moderation endpoint fails or rate
// limits.
type AblyRuleBodyguardBeforePublishConfig struct {
	RetryTimeout          types.Int64  `tfsdk:"retry_timeout"`
	MaxRetries            types.Int64  `tfsdk:"max_retries"`
	FailedAction          types.String `tfsdk:"failed_action"`
	TooManyRequestsAction types.String `tfsdk:"too_many_requests_action"`
}

// AblyRuleBodyguardTarget mirrors control.BodyguardTextModerationTarget.
type AblyRuleBodyguardTarget struct {
	ApiKey          types.String `tfsdk:"api_key"`
	ChannelID       types.String `tfsdk:"channel_id"`
	ApiURL          types.String `tfsdk:"api_url"`
	DefaultLanguage types.String `tfsdk:"default_language"`
}

// AblyRuleBodyguard is the tfsdk model for the Bodyguard text moderation rule.
//
// Note: unlike webhook/firehose rules (AblyRule), moderation rules have NO
// `source` and NO `request_mode`. They instead carry `invocation_mode`,
// `chat_room_filter` and a `before_publish_config` block, which is why this
// resource does not reuse the generic GetRuleSchema/CreateRule[T] plumbing.
type AblyRuleBodyguard struct {
	ID                  types.String                          `tfsdk:"id"`
	AppID               types.String                          `tfsdk:"app_id"`
	Status              types.String                          `tfsdk:"status"`
	InvocationMode      types.String                          `tfsdk:"invocation_mode"`
	ChatRoomFilter      types.String                          `tfsdk:"chat_room_filter"`
	BeforePublishConfig *AblyRuleBodyguardBeforePublishConfig `tfsdk:"before_publish_config"`
	Target              *AblyRuleBodyguardTarget              `tfsdk:"target"`
}

type ResourceRuleBodyguard struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleBodyguard{}
var _ resource.ResourceWithImportState = &ResourceRuleBodyguard{}

// Schema defines the schema for the resource.
//
// PORTED ONTO GENERATED CODE (see DEVELOPMENT.md "Porting a resource onto
// generated code"). The entire schema, attributes, types, nesting,
// sensitivity, descriptions, validators, defaults and plan modifiers, comes
// from the generated schema in internal/provider/codegen, produced by `make
// generate` from the in-repo control rule types plus the overrides table in
// codegen/ruletypesgen. The only hand-work left is stripping the generated
// CustomType from the nested blocks so the hand-written plain-struct model
// reflects cleanly, and setting the resource-level description.
func (r ResourceRuleBodyguard) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_rule_bodyguard.RuleBodyguardResourceSchema(ctx)

	bpc := s.Attributes["before_publish_config"].(schema.SingleNestedAttribute)
	bpc.CustomType = nil
	s.Attributes["before_publish_config"] = bpc

	tgt := s.Attributes["target"].(schema.SingleNestedAttribute)
	tgt.CustomType = nil
	s.Attributes["target"] = tgt

	s.MarkdownDescription = "The `ably_rule_bodyguard` resource allows you to create and manage an Ably integration rule for Bodyguard text moderation. This rule moderates messages before they are published. Read more at https://ably.com/docs/integrations/moderation"

	resp.Schema = s
}

func (r ResourceRuleBodyguard) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_bodyguard"
}

func (r *ResourceRuleBodyguard) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleBodyguard) Name() string {
	return "Bodyguard"
}

// getPlanBodyguardPost converts the plan model into the Control API create body.
func getPlanBodyguardPost(plan AblyRuleBodyguard) control.BodyguardTextModerationRulePost {
	return control.BodyguardTextModerationRulePost{
		Status:         plan.Status.ValueString(),
		RuleType:       bodyguardRuleType,
		InvocationMode: plan.InvocationMode.ValueString(),
		ChatRoomFilter: plan.ChatRoomFilter.ValueString(),
		BeforePublishConfig: control.BeforePublishConfig{
			RetryTimeout:          int(plan.BeforePublishConfig.RetryTimeout.ValueInt64()),
			MaxRetries:            int(plan.BeforePublishConfig.MaxRetries.ValueInt64()),
			FailedAction:          plan.BeforePublishConfig.FailedAction.ValueString(),
			TooManyRequestsAction: plan.BeforePublishConfig.TooManyRequestsAction.ValueString(),
		},
		Target: control.BodyguardTextModerationTarget{
			APIKey:          plan.Target.ApiKey.ValueString(),
			ChannelID:       plan.Target.ChannelID.ValueString(),
			APIURL:          plan.Target.ApiURL.ValueString(),
			DefaultLanguage: plan.Target.DefaultLanguage.ValueString(),
		},
	}
}

// getBodyguardResponse maps an API rule response back onto the tfsdk model.
//
// All moderation fields (invocation_mode, chat_room_filter,
// before_publish_config) are read back from the response, so out-of-band
// changes surface as drift. The only exception is the target's api_key, which
// is genuinely write-only (the API never returns it): the configured value is
// preserved from the plan/prior state, and left null when there is none (e.g.
// on import), since inventing a known "" would misrepresent what the API
// holds.
func getBodyguardResponse(rule *control.RuleResponse, plan *AblyRuleBodyguard) (AblyRuleBodyguard, diag.Diagnostics) {
	var diags diag.Diagnostics

	if rule.RuleType != bodyguardRuleType {
		diags.AddError(
			"Unexpected rule type in response",
			fmt.Sprintf("Expected rule type %q but received %q", bodyguardRuleType, rule.RuleType),
		)
		return AblyRuleBodyguard{}, diags
	}

	target, err := unmarshalTarget[control.BodyguardTextModerationTarget](rule.Target)
	if err != nil {
		diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal %s target: %s", bodyguardRuleType, err.Error()))
		return AblyRuleBodyguard{}, diags
	}

	// api_key is write-only: preserve the configured value from the plan so it
	// does not flip to empty on read, and stay null when there is no plan.
	apiKey := stringOrNull(target.APIKey)
	if plan != nil && plan.Target != nil && !plan.Target.ApiKey.IsNull() {
		apiKey = plan.Target.ApiKey
	}

	respTarget := &AblyRuleBodyguardTarget{
		ApiKey:          apiKey,
		ChannelID:       stringOrNull(target.ChannelID),
		ApiURL:          stringOrNull(target.APIURL),
		DefaultLanguage: stringOrNull(target.DefaultLanguage),
	}

	var beforePublish *AblyRuleBodyguardBeforePublishConfig
	if rule.BeforePublishConfig != nil {
		beforePublish = &AblyRuleBodyguardBeforePublishConfig{
			RetryTimeout:          types.Int64Value(int64(rule.BeforePublishConfig.RetryTimeout)),
			MaxRetries:            types.Int64Value(int64(rule.BeforePublishConfig.MaxRetries)),
			FailedAction:          types.StringValue(rule.BeforePublishConfig.FailedAction),
			TooManyRequestsAction: types.StringValue(rule.BeforePublishConfig.TooManyRequestsAction),
		}
	}

	respRule := AblyRuleBodyguard{
		ID:                  types.StringValue(rule.ID),
		AppID:               types.StringValue(rule.AppID),
		Status:              types.StringValue(rule.Status),
		InvocationMode:      stringOrNull(rule.InvocationMode),
		ChatRoomFilter:      stringOrNull(rule.ChatRoomFilter),
		BeforePublishConfig: beforePublish,
		Target:              respTarget,
	}

	return respRule, diags
}

// Create creates a new resource.
func (r ResourceRuleBodyguard) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.Provider().ensureConfigured(&resp.Diagnostics) {
		return
	}

	var plan AblyRuleBodyguard
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.Provider().client.CreateRule(ctx, plan.AppID.ValueString(), getPlanBodyguardPost(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error creating resource %s", r.Name()),
			fmt.Sprintf("Could not create resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues, respDiags := getBodyguardResponse(&rule, &plan)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, responseValues)
	resp.Diagnostics.Append(diags...)
}

// Read reads the resource.
func (r ResourceRuleBodyguard) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AblyRuleBodyguard
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.Provider().client.GetRule(ctx, state.AppID.ValueString(), state.ID.ValueString())
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error reading resource %s", r.Name()),
			fmt.Sprintf("Could not read resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues, respDiags := getBodyguardResponse(&rule, &state)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseValues)
	resp.Diagnostics.Append(diags...)
}

// Update updates an existing resource.
func (r ResourceRuleBodyguard) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AblyRuleBodyguard
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.Provider().client.UpdateRule(ctx, plan.AppID.ValueString(), plan.ID.ValueString(), getPlanBodyguardPost(plan))
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error updating resource %s", r.Name()),
			fmt.Sprintf("Could not update resource %s, unexpected error: %s", r.Name(), err.Error()),
		)
		return
	}

	responseValues, respDiags := getBodyguardResponse(&rule, &plan)
	resp.Diagnostics.Append(respDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &responseValues)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource.
func (r ResourceRuleBodyguard) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AblyRuleBodyguard
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.Provider().client.DeleteRule(ctx, state.AppID.ValueString(), state.ID.ValueString())
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Resource %s does not exist", r.Name()),
				fmt.Sprintf("Resource %s does not exist, it may have already been deleted: %s", r.Name(), err.Error()),
			)
		} else {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error deleting resource %s", r.Name()),
				fmt.Sprintf("Could not delete resource %s, unexpected error: %s", r.Name(), err.Error()),
			)
			return
		}
	}

	resp.State.RemoveResource(ctx)
}

// ImportState handles the import state functionality.
func (r ResourceRuleBodyguard) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
