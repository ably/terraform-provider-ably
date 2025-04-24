// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleLambda struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleLambda{}
var _ resource.ResourceWithImportState = &ResourceRuleLambda{}

// Schema defines the schema for the resource.
func (r ResourceRuleLambda) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional: true,
			},
			"function_name": schema.StringAttribute{
				Optional: true,
			},
			"enveloped":      GetEnvelopedSchema(),
			"authentication": GetAwsAuthSchema(),
		},
		"The `ably_rule_lambda` resource allows you to create and manage an Ably integration rule for AWS Lambda. Read more at https://ably.com/docs/general/webhooks/aws-lambda",
	)
}

func (r ResourceRuleLambda) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_lambda"
}

func (r *ResourceRuleLambda) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleLambda) Name() string {
	return "AWS Lambda"
}

// Create creates a new resource.
func (r ResourceRuleLambda) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleLambda) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleLambda) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleLambda) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetLambda](&r, ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleLambda) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
