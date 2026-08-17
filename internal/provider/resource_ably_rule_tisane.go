// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/resource_rule_tisane"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// tisaneRuleType is the Control API discriminator for Tisane text moderation
// rules.
const tisaneRuleType = "tisane/text-moderation"

// AblyRuleTisaneTarget mirrors control.TisaneTextModerationTarget.
type AblyRuleTisaneTarget struct {
	ApiKey          types.String `tfsdk:"api_key"`
	ModelURL        types.String `tfsdk:"model_url"`
	Thresholds      types.Map    `tfsdk:"thresholds"`
	DefaultLanguage types.String `tfsdk:"default_language"`
}

// AblyRuleTisane is the tfsdk model for the Tisane text moderation rule. Like
// the other moderation rules it has no `source` and no `request_mode`; see
// before_publish_rules.go.
type AblyRuleTisane struct {
	ID                  types.String             `tfsdk:"id"`
	AppID               types.String             `tfsdk:"app_id"`
	Status              types.String             `tfsdk:"status"`
	InvocationMode      types.String             `tfsdk:"invocation_mode"`
	ChatRoomFilter      types.String             `tfsdk:"chat_room_filter"`
	BeforePublishConfig *AblyBeforePublishConfig `tfsdk:"before_publish_config"`
	Target              *AblyRuleTisaneTarget    `tfsdk:"target"`
}

type ResourceRuleTisane struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleTisane{}
var _ resource.ResourceWithImportState = &ResourceRuleTisane{}

// Schema defines the schema for the resource.
//
// PORTED ONTO GENERATED CODE (see DEVELOPMENT.md "Porting a resource onto
// generated code"): the attribute set, types, nesting, sensitivity,
// descriptions, validators, defaults and plan modifiers all come from
// internal/provider/codegen, produced by `make generate`.
func (r ResourceRuleTisane) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_rule_tisane.RuleTisaneResourceSchema(ctx)
	stripNestedCustomTypes(&s, "before_publish_config", "target")

	s.MarkdownDescription = "The `ably_rule_tisane` resource allows you to create and manage an Ably integration rule for Tisane text moderation. This rule moderates messages before they are published. Read more at https://ably.com/docs/chat/moderation"

	resp.Schema = s
}

func (r ResourceRuleTisane) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_tisane"
}

func (r *ResourceRuleTisane) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleTisane) Name() string {
	return "Tisane text moderation"
}

func (r ResourceRuleTisane) crud() beforePublishCRUD[AblyRuleTisane] {
	return beforePublishCRUD[AblyRuleTisane]{
		p:        r.p,
		name:     r.Name(),
		appID:    func(m AblyRuleTisane) string { return m.AppID.ValueString() },
		id:       func(m AblyRuleTisane) string { return m.ID.ValueString() },
		post:     getPlanTisanePost,
		response: getTisaneResponse,
	}
}

// getPlanTisanePost converts the plan model into the Control API create body.
func getPlanTisanePost(ctx context.Context, plan AblyRuleTisane) (any, diag.Diagnostics) {
	thresholds, diags := thresholdsPost(ctx, plan.Target.Thresholds)
	if diags.HasError() {
		return nil, diags
	}

	return control.TisaneTextModerationRulePost{
		Status:              plan.Status.ValueString(),
		RuleType:            tisaneRuleType,
		InvocationMode:      plan.InvocationMode.ValueString(),
		ChatRoomFilter:      plan.ChatRoomFilter.ValueString(),
		BeforePublishConfig: beforePublishConfigPost(plan.BeforePublishConfig),
		Target: control.TisaneTextModerationTarget{
			APIKey:          plan.Target.ApiKey.ValueString(),
			ModelURL:        plan.Target.ModelURL.ValueString(),
			Thresholds:      thresholds,
			DefaultLanguage: plan.Target.DefaultLanguage.ValueString(),
		},
	}, diags
}

// getTisaneResponse maps an API rule response back onto the tfsdk model. Every
// field is read back from the response, the sensitive target api_key included:
// the Control API returns the full target for moderation rules (verified for
// bodyguard against the live API, 2026-07-08), so out-of-band changes surface as
// drift and import captures the complete resource.
func getTisaneResponse(ctx context.Context, rule *control.RuleResponse, _ *AblyRuleTisane) (AblyRuleTisane, diag.Diagnostics) {
	diags := checkRuleType(rule, tisaneRuleType)
	if diags.HasError() {
		return AblyRuleTisane{}, diags
	}

	target, err := unmarshalTarget[control.TisaneTextModerationTarget](rule.Target)
	if err != nil {
		diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal %s target: %s", tisaneRuleType, err.Error()))
		return AblyRuleTisane{}, diags
	}

	thresholds, thresholdDiags := thresholdsResponse(ctx, target.Thresholds)
	diags.Append(thresholdDiags...)
	if diags.HasError() {
		return AblyRuleTisane{}, diags
	}

	return AblyRuleTisane{
		ID:                  types.StringValue(rule.ID),
		AppID:               types.StringValue(rule.AppID),
		Status:              types.StringValue(rule.Status),
		InvocationMode:      stringOrNull(rule.InvocationMode),
		ChatRoomFilter:      stringOrNull(rule.ChatRoomFilter),
		BeforePublishConfig: beforePublishConfigResponse(rule.BeforePublishConfig),
		Target: &AblyRuleTisaneTarget{
			ApiKey:          stringOrNull(target.APIKey),
			ModelURL:        stringOrNull(target.ModelURL),
			Thresholds:      thresholds,
			DefaultLanguage: stringOrNull(target.DefaultLanguage),
		},
	}, diags
}

// Create creates a new resource.
func (r ResourceRuleTisane) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.crud().create(ctx, req, resp)
}

// Read reads the resource.
func (r ResourceRuleTisane) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.crud().read(ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleTisane) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.crud().update(ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleTisane) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.crud().delete(ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleTisane) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
