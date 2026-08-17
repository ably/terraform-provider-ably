// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/resource_rule_azure_moderation"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// azureModerationRuleType is the Control API discriminator for Azure text
// moderation rules.
const azureModerationRuleType = "azure/text-moderation"

// AblyRuleAzureModerationTarget mirrors control.AzureTextModerationTarget.
type AblyRuleAzureModerationTarget struct {
	ApiKey     types.String `tfsdk:"api_key"`
	Endpoint   types.String `tfsdk:"endpoint"`
	Thresholds types.Map    `tfsdk:"thresholds"`
}

// AblyRuleAzureModeration is the tfsdk model for the Azure text moderation
// rule. Like the other moderation rules it has no `source` and no
// `request_mode`; see before_publish_rules.go.
type AblyRuleAzureModeration struct {
	ID                  types.String                   `tfsdk:"id"`
	AppID               types.String                   `tfsdk:"app_id"`
	Status              types.String                   `tfsdk:"status"`
	InvocationMode      types.String                   `tfsdk:"invocation_mode"`
	ChatRoomFilter      types.String                   `tfsdk:"chat_room_filter"`
	BeforePublishConfig *AblyBeforePublishConfig       `tfsdk:"before_publish_config"`
	Target              *AblyRuleAzureModerationTarget `tfsdk:"target"`
}

type ResourceRuleAzureModeration struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleAzureModeration{}
var _ resource.ResourceWithImportState = &ResourceRuleAzureModeration{}

// Schema defines the schema for the resource.
//
// PORTED ONTO GENERATED CODE (see DEVELOPMENT.md "Porting a resource onto
// generated code").
func (r ResourceRuleAzureModeration) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_rule_azure_moderation.RuleAzureModerationResourceSchema(ctx)
	stripNestedCustomTypes(&s, "before_publish_config", "target")

	s.MarkdownDescription = "The `ably_rule_azure_moderation` resource allows you to create and manage an Ably integration rule for Azure AI Content Safety text moderation. This rule moderates messages before they are published. Read more at https://ably.com/docs/chat/moderation"

	resp.Schema = s
}

func (r ResourceRuleAzureModeration) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_azure_moderation"
}

func (r *ResourceRuleAzureModeration) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleAzureModeration) Name() string {
	return "Azure text moderation"
}

func (r ResourceRuleAzureModeration) crud() beforePublishCRUD[AblyRuleAzureModeration] {
	return beforePublishCRUD[AblyRuleAzureModeration]{
		p:        r.p,
		name:     r.Name(),
		appID:    func(m AblyRuleAzureModeration) string { return m.AppID.ValueString() },
		id:       func(m AblyRuleAzureModeration) string { return m.ID.ValueString() },
		post:     getPlanAzureModerationPost,
		response: getAzureModerationResponse,
	}
}

// getPlanAzureModerationPost converts the plan model into the Control API
// create body.
func getPlanAzureModerationPost(ctx context.Context, plan AblyRuleAzureModeration) (any, diag.Diagnostics) {
	thresholds, diags := thresholdsPost(ctx, plan.Target.Thresholds)
	if diags.HasError() {
		return nil, diags
	}

	return control.AzureTextModerationRulePost{
		Status:              plan.Status.ValueString(),
		RuleType:            azureModerationRuleType,
		InvocationMode:      plan.InvocationMode.ValueString(),
		ChatRoomFilter:      plan.ChatRoomFilter.ValueString(),
		BeforePublishConfig: beforePublishConfigPost(plan.BeforePublishConfig),
		Target: control.AzureTextModerationTarget{
			APIKey:     plan.Target.ApiKey.ValueString(),
			Endpoint:   plan.Target.Endpoint.ValueString(),
			Thresholds: thresholds,
		},
	}, diags
}

// getAzureModerationResponse maps an API rule response back onto the tfsdk
// model, api_key included: the Control API returns the full target for
// moderation rules, so out-of-band changes surface as drift and import captures
// the complete resource.
func getAzureModerationResponse(ctx context.Context, rule *control.RuleResponse, _ *AblyRuleAzureModeration) (AblyRuleAzureModeration, diag.Diagnostics) {
	diags := checkRuleType(rule, azureModerationRuleType)
	if diags.HasError() {
		return AblyRuleAzureModeration{}, diags
	}

	target, err := unmarshalTarget[control.AzureTextModerationTarget](rule.Target)
	if err != nil {
		diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal %s target: %s", azureModerationRuleType, err.Error()))
		return AblyRuleAzureModeration{}, diags
	}

	thresholds, thresholdDiags := thresholdsResponse(ctx, target.Thresholds)
	diags.Append(thresholdDiags...)
	if diags.HasError() {
		return AblyRuleAzureModeration{}, diags
	}

	return AblyRuleAzureModeration{
		ID:                  types.StringValue(rule.ID),
		AppID:               types.StringValue(rule.AppID),
		Status:              types.StringValue(rule.Status),
		InvocationMode:      stringOrNull(rule.InvocationMode),
		ChatRoomFilter:      stringOrNull(rule.ChatRoomFilter),
		BeforePublishConfig: beforePublishConfigResponse(rule.BeforePublishConfig),
		Target: &AblyRuleAzureModerationTarget{
			ApiKey:     stringOrNull(target.APIKey),
			Endpoint:   stringOrNull(target.Endpoint),
			Thresholds: thresholds,
		},
	}, diags
}

// Create creates a new resource.
func (r ResourceRuleAzureModeration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.crud().create(ctx, req, resp)
}

// Read reads the resource.
func (r ResourceRuleAzureModeration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.crud().read(ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleAzureModeration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.crud().update(ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleAzureModeration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.crud().delete(ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleAzureModeration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
