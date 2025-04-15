package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleAzureFunction struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleAzureFunction{}
var _ resource.ResourceWithImportState = &ResourceRuleAzureFunction{}

// Schema defines the schema for the resource.
func (r ResourceRuleAzureFunction) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"headers": GetHeaderSchema(),
			"azure_app_id": schema.StringAttribute{
				Required:    true,
				Description: "The Microsoft Azure Application ID. You can find your Microsoft Azure Application ID",
			},
			"function_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of your Microsoft Azure Function",
			},
			"signing_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "The signing key ID for use in batch mode. Ably will optionally sign the payload using an API key ensuring your servers can validate the payload using the private API key. See the [webhook security docs](https://ably.com/docs/general/webhooks#security) for more information",
			},
			"format": GetFormatSchema(),
		},
		"The `ably_rule_azure_function` resource allows you to create and manage an Ably integration rule for Microsoft Azure Functions. Read more at https://ably.com/docs/general/webhooks/azure")
}

func (r ResourceRuleAzureFunction) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_azure_function"
}

func (r *ResourceRuleAzureFunction) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleAzureFunction) Name() string {
	return "Azure Function"
}

// Create a new resource
func (r ResourceRuleAzureFunction) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleAzureFunction) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// // Update resource
func (r ResourceRuleAzureFunction) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// Delete resource
func (r ResourceRuleAzureFunction) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetAzureFunction](&r, ctx, req, resp)
}

// Import resource
func (r ResourceRuleAzureFunction) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
