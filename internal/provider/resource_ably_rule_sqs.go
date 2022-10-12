package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleSqsType struct{}

// Get Rule Resource schema
func (r resourceRuleSqsType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"region": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The region is which AWS SQS is hosted",
			},
			"aws_account_id": {
				Type:        types.StringType,
				Optional:    true,
				Description: "Your AWS account ID",
			},
			"queue_name": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The AWS SQS queue name",
			},
			"enveloped":      GetEnvelopedchema(),
			"format":         GetFormatSchema(),
			"authentication": GetAwsAuthSchema(),
		},
		"The `ably_rule_sqs` resource allows you to create and manage an Ably integration rule for AWS SQS. Read more at https://ably.com/docs/general/firehose/sqs-rule",
	), nil
}

// New resource instance
func (r resourceRuleSqsType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceRuleSqs{
		p: *(p.(*provider)),
	}, nil
}

type resourceRuleSqs struct {
	p provider
}

func (r *resourceRuleSqs) Provider() *provider {
	return &r.p
}

func (r *resourceRuleSqs) Name() string {
	return "AWS Sqs"
}

// Create a new resource
func (r resourceRuleSqs) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetSqs](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleSqs) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetSqs](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleSqs) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetSqs](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleSqs) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetSqs](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleSqs) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
