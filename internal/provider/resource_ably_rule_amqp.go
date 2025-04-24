// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleAMQP struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleAMQP{}
var _ resource.ResourceWithImportState = &ResourceRuleAMQP{}

// Schema defines the schema for the resource.
func (r ResourceRuleAMQP) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

func (r ResourceRuleAMQP) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_amqp"
}

func (r *ResourceRuleAMQP) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleAMQP) Name() string {
	return "AMQP"
}

// Create creates a new resource.
func (r ResourceRuleAMQP) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetAMQP](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleAMQP) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetAMQP](&r, ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleAMQP) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetAMQP](&r, ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleAMQP) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetAMQP](&r, ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleAMQP) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
