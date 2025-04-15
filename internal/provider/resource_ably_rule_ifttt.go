package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleIFTTT struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleIFTTT{}
var _ resource.ResourceWithImportState = &ResourceRuleIFTTT{}

// Get Rule Resource schema
func (r ResourceRuleIFTTT) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"webhook_key": schema.StringAttribute{
				Required:    true,
				Description: "The key in the Webhook Service Documentation page of your IFTTT account",
			},
			"event_name": schema.StringAttribute{
				Required:    true,
				Description: "The Event name is used to identify the IFTTT applet that will receive the Event, make sure the name matches the name of the IFTTT applet.",
			},
		},
		"The `ably_rule_ifttt` resource allows you to create and manage an Ably integration rule for IFTTT. Read more at https://ably.com/docs/general/webhooks/ifttt",
	)
}

func (r ResourceRuleIFTTT) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_ifttt"
}

func (r *ResourceRuleIFTTT) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleIFTTT) Name() string {
	return "IFTTT"
}

// Create a new resource
func (r ResourceRuleIFTTT) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetIFTTT](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleIFTTT) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetIFTTT](&r, ctx, req, resp)
}

// // Update resource
func (r ResourceRuleIFTTT) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetIFTTT](&r, ctx, req, resp)
}

// Delete resource
func (r ResourceRuleIFTTT) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetIFTTT](&r, ctx, req, resp)
}

// Import resource
func (r ResourceRuleIFTTT) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
