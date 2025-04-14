package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleIFTTT struct {
	p *AblyProvider
}

// Get Rule Resource schema
func (r resourceRuleIFTTT) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"webhook_key": {
				Type:        types.StringType,
				Required:    true,
				Description: "The key in the Webhook Service Documentation page of your IFTTT account",
			},
			"event_name": {
				Type:        types.StringType,
				Required:    true,
				Description: "The Event name is used to identify the IFTTT applet that will receive the Event, make sure the name matches the name of the IFTTT applet.",
			},
		},
		"The `ably_rule_ifttt` resource allows you to create and manage an Ably integration rule for IFTTT. Read more at https://ably.com/docs/general/webhooks/ifttt",
	), nil
}

func (r resourceRuleIFTTT) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_ifttt"
}

func (r *resourceRuleIFTTT) Provider() *AblyProvider {
	return r.p
}

func (r *resourceRuleIFTTT) Name() string {
	return "IFTTT"
}

// Create a new resource
func (r resourceRuleIFTTT) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetIFTTT](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleIFTTT) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetIFTTT](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleIFTTT) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetIFTTT](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleIFTTT) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetIFTTT](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleIFTTT) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
