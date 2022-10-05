package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleLambdaType struct{}

// Get Rule Resource schema
func (r resourceRuleLambdaType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"region": {
				Type:     types.StringType,
				Optional: true,
			},
			"function_name": {
				Type:     types.StringType,
				Optional: true,
			},
			"enveloped":      GetEnvelopedchema(),
			"authentication": GetAwsAuthSchema(),
		},
		"The `ably_rule_lambda` resource allows you to create and manage an Ably integration rule for AWS Lambda. Read more at https://ably.com/docs/general/webhooks/aws-lambda",
	), nil
}

// New resource instance
func (r resourceRuleLambdaType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceRuleLambda{
		p: *(p.(*provider)),
	}, nil
}

type resourceRuleLambda struct {
	p provider
}

func (r *resourceRuleLambda) Provider() *provider {
	return &r.p
}

func (r *resourceRuleLambda) Name() string {
	return "AWS Lambda"
}

// Create a new resource
func (r resourceRuleLambda) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleLambda) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleLambda) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleLambda) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleLambda) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
