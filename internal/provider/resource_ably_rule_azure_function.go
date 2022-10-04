package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleAzureFunctionType struct{}

// Get Rule Resource schema
func (r resourceRuleAzureFunctionType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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

// New resource instance
func (r resourceRuleAzureFunctionType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceRuleAzureFunction{
		p: *(p.(*provider)),
	}, nil
}

type resourceRuleAzureFunction struct {
	p provider
}

func (r *resourceRuleAzureFunction) Provider() *provider {
	return &r.p
}

func (r *resourceRuleAzureFunction) Name() string {
	return "Azure Function"
}

// Create a new resource
func (r resourceRuleAzureFunction) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleAzureFunction) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleAzureFunction) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleAzureFunction) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleAzureFunction) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportRule(&r, ctx, req, resp)
}
