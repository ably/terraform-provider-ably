package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleCloudflareWorker struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleCloudflareWorker{}
var _ resource.ResourceWithImportState = &ResourceRuleCloudflareWorker{}

// Schema defines the schema for the resource.
func (r ResourceRuleCloudflareWorker) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The webhook URL that Ably will POST events to",
			},
			"headers": GetHeaderSchema(),
			"signing_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "The signing key ID for use in batch mode. Ably will optionally sign the payload using an API key ensuring your servers can validate the payload using the private API key. See the [webhook security docs](https://ably.com/docs/general/webhooks#security) for more information",
			},
		},
		"The `ably_rule_cloudflare_worker` resource allows you to create and manage an Ably integration rule for Cloudflare workers. Read more at https://ably.com/docs/general/webhooks/cloudflare")
}

func (r ResourceRuleCloudflareWorker) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_cloudflare_worker"
}

func (r *ResourceRuleCloudflareWorker) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleCloudflareWorker) Name() string {
	return "Cloudflare Worker"
}

// Create a new resource
func (r ResourceRuleCloudflareWorker) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetCloudflareWorker](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleCloudflareWorker) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetCloudflareWorker](&r, ctx, req, resp)
}

// // Update resource
func (r ResourceRuleCloudflareWorker) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetCloudflareWorker](&r, ctx, req, resp)
}

// Delete resource
func (r ResourceRuleCloudflareWorker) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetCloudflareWorker](&r, ctx, req, resp)
}

// Import resource
func (r ResourceRuleCloudflareWorker) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
