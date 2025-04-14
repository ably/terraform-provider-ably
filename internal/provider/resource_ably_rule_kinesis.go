package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleKinesis struct {
	p *AblyProvider
}

// Get Rule Resource schema
func (r resourceRuleKinesis) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"region": {
				Type:     types.StringType,
				Optional: true,
			},
			"stream_name": {
				Type:     types.StringType,
				Optional: true,
			},
			"partition_key": {
				Type:     types.StringType,
				Optional: true,
			},
			"enveloped":      GetEnvelopedSchema(),
			"format":         GetFormatSchema(),
			"authentication": GetAwsAuthSchema(),
		},
		"The `ably_rule_kinesis` resource allows you to create and manage an Ably integration rule for AWS Kinesis. Read more at https://ably.com/docs/general/firehose/kinesis-rule",
	), nil
}

func (r resourceRuleKinesis) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_kinesis"
}

func (r *resourceRuleKinesis) Provider() *AblyProvider {
	return r.p
}

func (r *resourceRuleKinesis) Name() string {
	return "AWS Kinesis"
}

// Create a new resource
func (r resourceRuleKinesis) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleKinesis) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleKinesis) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleKinesis) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleKinesis) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
