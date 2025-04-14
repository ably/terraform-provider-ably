package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleLambda struct {
	p *AblyProvider
}

// Get Rule Resource schema
func (r resourceRuleLambda) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"enveloped":      GetEnvelopedSchema(),
			"authentication": GetAwsAuthSchema(),
		},
		"The `ably_rule_lambda` resource allows you to create and manage an Ably integration rule for AWS Lambda. Read more at https://ably.com/docs/general/webhooks/aws-lambda",
	), nil
}

func (r resourceRuleLambda) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_lambda"
}

func (r *resourceRuleLambda) Provider() *AblyProvider {
	return r.p
}

func (r *resourceRuleLambda) Name() string {
	return "AWS Lambda"
}

// Create a new resource
func (r resourceRuleLambda) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleLambda) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleLambda) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleLambda) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleLambda) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
