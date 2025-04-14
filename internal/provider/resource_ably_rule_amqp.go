package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleAmqp struct {
	p *AblyProvider
}

// Get Rule Resource schema
func (r resourceRuleAmqp) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"queue_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The ID of your Ably queue",
			},
			"headers":   GetHeaderSchema(),
			"enveloped": GetEnvelopedSchema(),
			"format":    GetFormatSchema(),
		},
		"The `ably_rule_amqp` resource allows you to create and manage an Ably integration rule for AMQP. Read more at https://ably.com/docs/general/firehose/amqp-rule"), nil
}

func (r resourceRuleAmqp) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_amqp"
}

func (r *resourceRuleAmqp) Provider() *AblyProvider {
	return r.p
}

func (r *resourceRuleAmqp) Name() string {
	return "AMQP"
}

// Create a new resource
func (r resourceRuleAmqp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetAmqp](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleAmqp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetAmqp](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleAmqp) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetAmqp](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleAmqp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetAmqp](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleAmqp) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
