package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleKinesisType struct{}

// Get Rule Resource schema
func (r resourceRuleKinesisType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"enveloped":      GetEnvelopedchema(),
			"format":         GetFormatSchema(),
			"authentication": GetAwsAuthSchema(),
		},
		"The `ably_rule_kinesis` resource allows you to create and manage an Ably integration rule for AWS Kinesis. Read more at https://ably.com/docs/general/firehose/kinesis-rule",
	), nil
}

// New resource instance
func (r resourceRuleKinesisType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceRuleKinesis{
		p: *(p.(*provider)),
	}, nil
}

type resourceRuleKinesis struct {
	p provider
}

func (r *resourceRuleKinesis) Provider() *provider {
	return &r.p
}

func (r *resourceRuleKinesis) Name() string {
	return "AWS Kinesis"
}

// Create a new resource
func (r resourceRuleKinesis) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleKinesis) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleKinesis) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleKinesis) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetKinesis](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleKinesis) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
