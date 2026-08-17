// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/resource_rule_hive_text"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// hiveTextRuleType is the Control API discriminator for Hive text-model-only
// moderation rules.
const hiveTextRuleType = "hive/text-model-only"

// AblyRuleHiveTextTarget mirrors control.HiveTextModelOnlyTarget.
type AblyRuleHiveTextTarget struct {
	ApiKey     types.String `tfsdk:"api_key"`
	ModelURL   types.String `tfsdk:"model_url"`
	Thresholds types.Map    `tfsdk:"thresholds"`
}

// AblyRuleHiveText is the tfsdk model for the Hive text-model-only moderation
// rule. Like the other moderation rules it has no `source` and no
// `request_mode`; see before_publish_rules.go.
type AblyRuleHiveText struct {
	ID                  types.String             `tfsdk:"id"`
	AppID               types.String             `tfsdk:"app_id"`
	Status              types.String             `tfsdk:"status"`
	InvocationMode      types.String             `tfsdk:"invocation_mode"`
	ChatRoomFilter      types.String             `tfsdk:"chat_room_filter"`
	BeforePublishConfig *AblyBeforePublishConfig `tfsdk:"before_publish_config"`
	Target              *AblyRuleHiveTextTarget  `tfsdk:"target"`
}

type ResourceRuleHiveText struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleHiveText{}
var _ resource.ResourceWithImportState = &ResourceRuleHiveText{}

// Schema defines the schema for the resource.
//
// PORTED ONTO GENERATED CODE (see DEVELOPMENT.md "Porting a resource onto
// generated code").
func (r ResourceRuleHiveText) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_rule_hive_text.RuleHiveTextResourceSchema(ctx)
	stripNestedCustomTypes(&s, "before_publish_config", "target")

	s.MarkdownDescription = "The `ably_rule_hive_text` resource allows you to create and manage an Ably integration rule for Hive AI text moderation, using the model only with no dashboard. This rule moderates messages before they are published. Read more at https://ably.com/docs/chat/moderation"

	resp.Schema = s
}

func (r ResourceRuleHiveText) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_hive_text"
}

func (r *ResourceRuleHiveText) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleHiveText) Name() string {
	return "Hive text moderation"
}

func (r ResourceRuleHiveText) crud() beforePublishCRUD[AblyRuleHiveText] {
	return beforePublishCRUD[AblyRuleHiveText]{
		p:        r.p,
		name:     r.Name(),
		appID:    func(m AblyRuleHiveText) string { return m.AppID.ValueString() },
		id:       func(m AblyRuleHiveText) string { return m.ID.ValueString() },
		post:     getPlanHiveTextPost,
		response: getHiveTextResponse,
	}
}

// getPlanHiveTextPost converts the plan model into the Control API create body.
func getPlanHiveTextPost(ctx context.Context, plan AblyRuleHiveText) (any, diag.Diagnostics) {
	thresholds, diags := thresholdsPost(ctx, plan.Target.Thresholds)
	if diags.HasError() {
		return nil, diags
	}

	return control.HiveTextModelOnlyRulePost{
		Status:              plan.Status.ValueString(),
		RuleType:            hiveTextRuleType,
		InvocationMode:      plan.InvocationMode.ValueString(),
		ChatRoomFilter:      plan.ChatRoomFilter.ValueString(),
		BeforePublishConfig: beforePublishConfigPost(plan.BeforePublishConfig),
		Target: control.HiveTextModelOnlyTarget{
			APIKey:     plan.Target.ApiKey.ValueString(),
			ModelURL:   plan.Target.ModelURL.ValueString(),
			Thresholds: thresholds,
		},
	}, diags
}

// getHiveTextResponse maps an API rule response back onto the tfsdk model,
// api_key included: the Control API returns the full target for moderation
// rules, so out-of-band changes surface as drift and import captures the
// complete resource.
func getHiveTextResponse(ctx context.Context, rule *control.RuleResponse, _ *AblyRuleHiveText) (AblyRuleHiveText, diag.Diagnostics) {
	diags := checkRuleType(rule, hiveTextRuleType)
	if diags.HasError() {
		return AblyRuleHiveText{}, diags
	}

	target, err := unmarshalTarget[control.HiveTextModelOnlyTarget](rule.Target)
	if err != nil {
		diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal %s target: %s", hiveTextRuleType, err.Error()))
		return AblyRuleHiveText{}, diags
	}

	thresholds, thresholdDiags := thresholdsResponse(ctx, target.Thresholds)
	diags.Append(thresholdDiags...)
	if diags.HasError() {
		return AblyRuleHiveText{}, diags
	}

	return AblyRuleHiveText{
		ID:                  types.StringValue(rule.ID),
		AppID:               types.StringValue(rule.AppID),
		Status:              types.StringValue(rule.Status),
		InvocationMode:      stringOrNull(rule.InvocationMode),
		ChatRoomFilter:      stringOrNull(rule.ChatRoomFilter),
		BeforePublishConfig: beforePublishConfigResponse(rule.BeforePublishConfig),
		Target: &AblyRuleHiveTextTarget{
			ApiKey:     stringOrNull(target.APIKey),
			ModelURL:   stringOrNull(target.ModelURL),
			Thresholds: thresholds,
		},
	}, diags
}

// Create creates a new resource.
func (r ResourceRuleHiveText) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.crud().create(ctx, req, resp)
}

// Read reads the resource.
func (r ResourceRuleHiveText) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.crud().read(ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleHiveText) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.crud().update(ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleHiveText) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.crud().delete(ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleHiveText) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
