// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/resource_rule_hive_dashboard"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// hiveDashboardRuleType is the Control API discriminator for Hive dashboard
// moderation rules.
const hiveDashboardRuleType = "hive/dashboard"

// AblyRuleHiveDashboardTarget mirrors control.HiveDashboardTarget.
type AblyRuleHiveDashboardTarget struct {
	ApiKey          types.String `tfsdk:"api_key"`
	CheckWatchLists types.Bool   `tfsdk:"check_watch_lists"`
}

// AblyRuleHiveDashboard is the tfsdk model for the Hive dashboard moderation
// rule.
//
// Note: unlike the other moderation rules, Hive dashboard rules run after a
// message is published (invocation_mode AFTER_PUBLISH) and carry NO
// before_publish_config; moderation decisions are made in the Hive dashboard, so
// there is no endpoint to retry against before publishing.
type AblyRuleHiveDashboard struct {
	ID             types.String                 `tfsdk:"id"`
	AppID          types.String                 `tfsdk:"app_id"`
	Status         types.String                 `tfsdk:"status"`
	InvocationMode types.String                 `tfsdk:"invocation_mode"`
	ChatRoomFilter types.String                 `tfsdk:"chat_room_filter"`
	Target         *AblyRuleHiveDashboardTarget `tfsdk:"target"`
}

type ResourceRuleHiveDashboard struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleHiveDashboard{}
var _ resource.ResourceWithImportState = &ResourceRuleHiveDashboard{}

// Schema defines the schema for the resource.
//
// PORTED ONTO GENERATED CODE (see DEVELOPMENT.md "Porting a resource onto
// generated code").
func (r ResourceRuleHiveDashboard) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_rule_hive_dashboard.RuleHiveDashboardResourceSchema(ctx)
	stripNestedCustomTypes(&s, "target")

	s.MarkdownDescription = "The `ably_rule_hive_dashboard` resource allows you to create and manage an Ably integration rule for Hive AI moderation with the Hive dashboard. Unlike the other moderation rules this one runs *after* a message is published (`invocation_mode` is `AFTER_PUBLISH`), so it takes no `before_publish_config`. Read more at https://ably.com/docs/chat/moderation"

	resp.Schema = s
}

func (r ResourceRuleHiveDashboard) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_hive_dashboard"
}

func (r *ResourceRuleHiveDashboard) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleHiveDashboard) Name() string {
	return "Hive dashboard moderation"
}

func (r ResourceRuleHiveDashboard) crud() beforePublishCRUD[AblyRuleHiveDashboard] {
	return beforePublishCRUD[AblyRuleHiveDashboard]{
		p:        r.p,
		name:     r.Name(),
		appID:    func(m AblyRuleHiveDashboard) string { return m.AppID.ValueString() },
		id:       func(m AblyRuleHiveDashboard) string { return m.ID.ValueString() },
		post:     getPlanHiveDashboardPost,
		response: getHiveDashboardResponse,
	}
}

// getPlanHiveDashboardPost converts the plan model into the Control API create
// body.
func getPlanHiveDashboardPost(_ context.Context, plan AblyRuleHiveDashboard) (any, diag.Diagnostics) {
	return control.HiveDashboardRulePost{
		Status:         plan.Status.ValueString(),
		RuleType:       hiveDashboardRuleType,
		InvocationMode: plan.InvocationMode.ValueString(),
		ChatRoomFilter: plan.ChatRoomFilter.ValueString(),
		Target: control.HiveDashboardTarget{
			APIKey:          plan.Target.ApiKey.ValueString(),
			CheckWatchLists: optionalBoolPtr(plan.Target.CheckWatchLists),
		},
	}, nil
}

// getHiveDashboardResponse maps an API rule response back onto the tfsdk model,
// api_key included: the Control API returns the full target for moderation
// rules, so out-of-band changes surface as drift and import captures the
// complete resource.
func getHiveDashboardResponse(_ context.Context, rule *control.RuleResponse, _ *AblyRuleHiveDashboard) (AblyRuleHiveDashboard, diag.Diagnostics) {
	diags := checkRuleType(rule, hiveDashboardRuleType)
	if diags.HasError() {
		return AblyRuleHiveDashboard{}, diags
	}

	target, err := unmarshalTarget[control.HiveDashboardTarget](rule.Target)
	if err != nil {
		diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal %s target: %s", hiveDashboardRuleType, err.Error()))
		return AblyRuleHiveDashboard{}, diags
	}

	return AblyRuleHiveDashboard{
		ID:             types.StringValue(rule.ID),
		AppID:          types.StringValue(rule.AppID),
		Status:         types.StringValue(rule.Status),
		InvocationMode: stringOrNull(rule.InvocationMode),
		ChatRoomFilter: stringOrNull(rule.ChatRoomFilter),
		Target: &AblyRuleHiveDashboardTarget{
			ApiKey:          stringOrNull(target.APIKey),
			CheckWatchLists: optBoolValue(target.CheckWatchLists),
		},
	}, diags
}

// Create creates a new resource.
func (r ResourceRuleHiveDashboard) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.crud().create(ctx, req, resp)
}

// Read reads the resource.
func (r ResourceRuleHiveDashboard) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.crud().read(ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleHiveDashboard) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.crud().update(ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleHiveDashboard) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.crud().delete(ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleHiveDashboard) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
