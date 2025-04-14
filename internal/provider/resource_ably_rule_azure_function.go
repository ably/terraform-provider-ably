package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleAzureFunction struct {
	p *AblyProvider
}

// Get Rule Resource schema
func (r resourceRuleAzureFunction) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"headers": GetHeaderSchema(),
			"azure_app_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The Microsoft Azure Application ID. You can find your Microsoft Azure Application ID",
			},
			"function_name": {
				Type:        types.StringType,
				Required:    true,
				Description: "The name of your Microsoft Azure Function",
			},
			"signing_key_id": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The signing key ID for use in batch mode. Ably will optionally sign the payload using an API key ensuring your servers can validate the payload using the private API key. See the [webhook security docs](https://ably.com/docs/general/webhooks#security) for more information",
			},
			"format": GetFormatSchema(),
		},
		"The `ably_rule_azure_function` resource allows you to create and manage an Ably integration rule for Microsoft Azure Functions. Read more at https://ably.com/docs/general/webhooks/azure",
	), nil
}

func (r resourceRuleAzureFunction) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_azure_function"
}

func (r *resourceRuleAzureFunction) Provider() *AblyProvider {
	return r.p
}

func (r *resourceRuleAzureFunction) Name() string {
	return "Azure Function"
}

// Create a new resource
func (r resourceRuleAzureFunction) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleAzureFunction) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleAzureFunction) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleAzureFunction) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleAzureFunction) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
