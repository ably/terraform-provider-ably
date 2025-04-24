package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleKinesis struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleKinesis{}
var _ resource.ResourceWithImportState = &ResourceRuleKinesis{}

// Get Rule Resource schema
func (r ResourceRuleKinesis) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional: true,
			},
			"stream_name": schema.StringAttribute{
				Optional: true,
			},
			"partition_key": schema.StringAttribute{
				Optional: true,
			},
			"enveloped":      GetEnvelopedSchema(),
			"format":         GetFormatSchema(),
			"authentication": GetAwsAuthSchema(),
		},
		"The `ably_rule_kinesis` resource allows you to create and manage an Ably integration rule for AWS Kinesis. Read more at https://ably.com/docs/general/firehose/kinesis-rule",
	)
}

func (r ResourceRuleKinesis) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_kinesis"
}

func (r *ResourceRuleKinesis) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleKinesis) Name() string {
	return "AWS Kinesis"
}

// Create a new resource
func (r ResourceRuleKinesis) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleKinesis) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// // Update resource
func (r ResourceRuleKinesis) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// Delete resource
func (r ResourceRuleKinesis) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// Import resource
func (r ResourceRuleKinesis) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
