// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
func (r ResourceRuleBodyguard) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `ably_rule_bodyguard` resource allows you to create and manage an Ably integration rule for Bodyguard text moderation. This rule moderates messages before they are published. Read more at https://ably.com/docs/integrations/moderation",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The rule ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				Required:    true,
				Description: "The Ably application ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The status of the rule. Rules can be enabled or disabled.",
				Default:     stringdefault.StaticString("enabled"),
				Validators: []validator.String{
					stringvalidator.OneOf("enabled", "disabled"),
				},
			},
			"invocation_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How the moderation endpoint is invoked. Only `BEFORE_PUBLISH` is supported, which moderates messages before they are published.",
				Default:     stringdefault.StaticString("BEFORE_PUBLISH"),
				Validators: []validator.String{
					stringvalidator.OneOf("BEFORE_PUBLISH"),
				},
			},
			"chat_room_filter": schema.StringAttribute{
				Optional:    true,
				Description: "An optional filter limiting the rule to matching chat rooms, given as a slash-delimited regular expression, e.g. `/room-.*/`.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^/.*/$`), "must be a slash-delimited regular expression, e.g. /room-.*/"),
				},
			},
			"before_publish_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Configuration controlling retry and failure behaviour when moderating messages before publish.",
				Attributes: map[string]schema.Attribute{
					"retry_timeout": schema.Int64Attribute{
						Required:    true,
						Description: "The timeout, in milliseconds, after which a moderation request is retried.",
					},
					"max_retries": schema.Int64Attribute{
						Required:    true,
						Description: "The maximum number of times a moderation request is retried before the failed action is taken.",
					},
					"failed_action": schema.StringAttribute{
						Required:    true,
						Description: "The action to take when moderation fails after exhausting retries. One of `REJECT` (do not publish the message) or `PUBLISH` (publish it without moderation).",
						Validators: []validator.String{
							stringvalidator.OneOf("REJECT", "PUBLISH"),
						},
					},
					"too_many_requests_action": schema.StringAttribute{
						Required:    true,
						Description: "The action to take when the moderation endpoint rate limits the request (HTTP 429). One of `RETRY` (retry the moderation request) or `FAIL` (treat it as a failure and apply `failed_action`).",
						Validators: []validator.String{
							stringvalidator.OneOf("RETRY", "FAIL"),
						},
					},
				},
			},
			"target": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The Bodyguard target for the rule.",
				Attributes: map[string]schema.Attribute{
					"api_key": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "The Bodyguard API key used to authenticate moderation requests.",
					},
					"channel_id": schema.StringAttribute{
						Optional:    true,
						Description: "The Bodyguard channel ID to associate moderated content with.",
					},
					"api_url": schema.StringAttribute{
						Optional:    true,
						Description: "An optional override for the Bodyguard API URL.",
					},
					"default_language": schema.StringAttribute{
						Optional:    true,
						Description: "The default language code used when moderating messages, e.g. `en`.",
					},
				},
			},
		},
	}
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
// The in-repo control.RuleResponse only decodes id/appId/status/ruleType and
// the (untyped) target. The Control API does return beforePublishConfig,
// invocationMode and chatRoomFilter for moderation rules, but the shared
// RuleResponse type does not yet have fields for them, so the client discards
// them; the target's api_key is genuinely write-only (never returned). We
// therefore preserve those values from the plan/prior state to avoid Terraform
// "inconsistent result after apply" errors and the opaque "inconsistent values
// for sensitive attribute" diff (the target block holds the sensitive api_key,
// so any field mismatch trips it).
//
// Known limitation: because we carry these forward rather than reading them
// back, out-of-band changes to invocation_mode/chat_room_filter/
// before_publish_config are not detected as drift. Decoding them properly
// belongs with the moderation/before-publish rule family work in
// CODEGEN_STRATEGY.md, where control.RuleResponse should learn these fields.
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
	// does not flip to empty on read.
	apiKey := types.StringValue(target.APIKey)
	if plan != nil && plan.Target != nil && !plan.Target.ApiKey.IsNull() {
		apiKey = plan.Target.ApiKey
	}

	respTarget := &AblyRuleBodyguardTarget{
		ApiKey:          apiKey,
		ChannelID:       stringOrNull(target.ChannelID),
		ApiURL:          stringOrNull(target.APIURL),
		DefaultLanguage: stringOrNull(target.DefaultLanguage),
	}

	// invocation_mode, chat_room_filter and before_publish_config are not
	// decoded by the in-repo control client (see the function doc), so carry
	// them forward from the plan/state.
	invocationMode := types.StringNull()
	chatRoomFilter := types.StringNull()
	var beforePublish *AblyRuleBodyguardBeforePublishConfig
	if plan != nil {
		invocationMode = plan.InvocationMode
		chatRoomFilter = plan.ChatRoomFilter
		beforePublish = plan.BeforePublishConfig
	}

	respRule := AblyRuleBodyguard{
		ID:                  types.StringValue(rule.ID),
		AppID:               types.StringValue(rule.AppID),
		Status:              types.StringValue(rule.Status),
		InvocationMode:      invocationMode,
		ChatRoomFilter:      chatRoomFilter,
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
