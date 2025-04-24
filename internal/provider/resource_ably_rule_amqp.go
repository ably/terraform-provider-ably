package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleAmqp struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleAmqp{}
var _ resource.ResourceWithImportState = &ResourceRuleAmqp{}

// Schema defines the schema for the resource.
func (r ResourceRuleAmqp) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"queue_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of your Ably queue",
			},
			"headers":   GetHeaderSchema(),
			"enveloped": GetEnvelopedSchema(),
			"format":    GetFormatSchema(),
		},
		"The `ably_rule_amqp` resource allows you to create and manage an Ably integration rule for AMQP. Read more at https://ably.com/docs/general/firehose/amqp-rule")
}

func (r ResourceRuleAmqp) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_amqp"
}

func (r *ResourceRuleAmqp) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleAmqp) Name() string {
	return "AMQP"
}

// Create a new resource
func (r ResourceRuleAmqp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetAmqp](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleAmqp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetAmqp](&r, ctx, req, resp)
}

// // Update resource
func (r ResourceRuleAmqp) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetAmqp](&r, ctx, req, resp)
}

// Delete resource
func (r ResourceRuleAmqp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetAmqp](&r, ctx, req, resp)
}

// Import resource
func (r ResourceRuleAmqp) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
