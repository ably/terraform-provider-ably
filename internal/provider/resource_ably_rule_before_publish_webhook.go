// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/ably/terraform-provider-ably/internal/provider/codegen/resource_rule_before_publish_webhook"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// beforePublishWebhookRuleType is the Control API discriminator for
// before-publish webhook rules.
const beforePublishWebhookRuleType = "http/before-publish"

// AblyRuleBeforePublishWebhookTarget mirrors
// control.BeforePublishWebhookTarget.
type AblyRuleBeforePublishWebhookTarget struct {
	Url     types.String      `tfsdk:"url"`
	Headers []AblyRuleHeaders `tfsdk:"headers"`
}

// AblyRuleBeforePublishWebhook is the tfsdk model for the before-publish
// webhook rule.
//
// Despite being a webhook this is NOT the AblyRule shape: before-publish rules
// have no `source` and no `request_mode`, and carry `invocation_mode`,
// `chat_room_filter` and `before_publish_config` instead. See
// before_publish_rules.go.
type AblyRuleBeforePublishWebhook struct {
	ID                  types.String                        `tfsdk:"id"`
	AppID               types.String                        `tfsdk:"app_id"`
	Status              types.String                        `tfsdk:"status"`
	InvocationMode      types.String                        `tfsdk:"invocation_mode"`
	ChatRoomFilter      types.String                        `tfsdk:"chat_room_filter"`
	BeforePublishConfig *AblyBeforePublishConfig            `tfsdk:"before_publish_config"`
	Target              *AblyRuleBeforePublishWebhookTarget `tfsdk:"target"`
}

type ResourceRuleBeforePublishWebhook struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleBeforePublishWebhook{}
var _ resource.ResourceWithImportState = &ResourceRuleBeforePublishWebhook{}

// Schema defines the schema for the resource.
//
// PORTED ONTO GENERATED CODE (see DEVELOPMENT.md "Porting a resource onto
// generated code").
func (r ResourceRuleBeforePublishWebhook) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_rule_before_publish_webhook.RuleBeforePublishWebhookResourceSchema(ctx)
	stripNestedCustomTypes(&s, "before_publish_config", "target")
	stripListNestedCustomType(&s, "target", "headers")

	s.MarkdownDescription = "The `ably_rule_before_publish_webhook` resource allows you to create and manage an Ably integration rule that calls your own HTTP endpoint before a message is published, so the endpoint can allow, reject or amend it. Read more at https://ably.com/docs/chat/moderation"

	resp.Schema = s
}

func (r ResourceRuleBeforePublishWebhook) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_before_publish_webhook"
}

func (r *ResourceRuleBeforePublishWebhook) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleBeforePublishWebhook) Name() string {
	return "Before-publish webhook"
}

func (r ResourceRuleBeforePublishWebhook) crud() beforePublishCRUD[AblyRuleBeforePublishWebhook] {
	return beforePublishCRUD[AblyRuleBeforePublishWebhook]{
		p:        r.p,
		name:     r.Name(),
		appID:    func(m AblyRuleBeforePublishWebhook) string { return m.AppID.ValueString() },
		id:       func(m AblyRuleBeforePublishWebhook) string { return m.ID.ValueString() },
		post:     getPlanBeforePublishWebhookPost,
		response: getBeforePublishWebhookResponse,
	}
}

// getPlanBeforePublishWebhookPost converts the plan model into the Control API
// create body.
func getPlanBeforePublishWebhookPost(_ context.Context, plan AblyRuleBeforePublishWebhook) (any, diag.Diagnostics) {
	return control.BeforePublishWebhookRulePost{
		Status:              plan.Status.ValueString(),
		RuleType:            beforePublishWebhookRuleType,
		InvocationMode:      plan.InvocationMode.ValueString(),
		ChatRoomFilter:      plan.ChatRoomFilter.ValueString(),
		BeforePublishConfig: beforePublishConfigPost(plan.BeforePublishConfig),
		Target: control.BeforePublishWebhookTarget{
			URL:     plan.Target.Url.ValueString(),
			Headers: GetHeaders(plan.Target.Headers),
		},
	}, nil
}

// getBeforePublishWebhookResponse maps an API rule response back onto the tfsdk
// model. Headers are returned by the API, so they come from the response rather
// than the plan.
func getBeforePublishWebhookResponse(_ context.Context, rule *control.RuleResponse, _ *AblyRuleBeforePublishWebhook) (AblyRuleBeforePublishWebhook, diag.Diagnostics) {
	diags := checkRuleType(rule, beforePublishWebhookRuleType)
	if diags.HasError() {
		return AblyRuleBeforePublishWebhook{}, diags
	}

	target, err := unmarshalTarget[control.BeforePublishWebhookTarget](rule.Target)
	if err != nil {
		diags.AddError("Error unmarshalling rule target", fmt.Sprintf("Could not unmarshal %s target: %s", beforePublishWebhookRuleType, err.Error()))
		return AblyRuleBeforePublishWebhook{}, diags
	}

	return AblyRuleBeforePublishWebhook{
		ID:                  types.StringValue(rule.ID),
		AppID:               types.StringValue(rule.AppID),
		Status:              types.StringValue(rule.Status),
		InvocationMode:      stringOrNull(rule.InvocationMode),
		ChatRoomFilter:      stringOrNull(rule.ChatRoomFilter),
		BeforePublishConfig: beforePublishConfigResponse(rule.BeforePublishConfig),
		Target: &AblyRuleBeforePublishWebhookTarget{
			Url:     stringOrNull(target.URL),
			Headers: ToHeaders(target.Headers),
		},
	}, diags
}

// Create creates a new resource.
func (r ResourceRuleBeforePublishWebhook) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.crud().create(ctx, req, resp)
}

// Read reads the resource.
func (r ResourceRuleBeforePublishWebhook) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.crud().read(ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleBeforePublishWebhook) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.crud().update(ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleBeforePublishWebhook) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.crud().delete(ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleBeforePublishWebhook) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
