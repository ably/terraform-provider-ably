package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleCloudflareWorker struct {
	p *provider
}

// Get Rule Resource schema
func (r resourceRuleCloudflareWorker) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"url": {
				Type:        types.StringType,
				Required:    true,
				Description: "The webhook URL that Ably will POST events to",
			},
			"headers": GetHeaderSchema(),
			"signing_key_id": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The signing key ID for use in batch mode. Ably will optionally sign the payload using an API key ensuring your servers can validate the payload using the private API key. See the [webhook security docs](https://ably.com/docs/general/webhooks#security) for more information",
			},
		},
		"The `ably_rule_cloudflare_worker` resource allows you to create and manage an Ably integration rule for Cloudflare workers. Read more at https://ably.com/docs/general/webhooks/cloudflare"), nil
}

func (r resourceRuleCloudflareWorker) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_rule_cloudflare_worker"
}

func (r *resourceRuleCloudflareWorker) Provider() *provider {
	return r.p
}

func (r *resourceRuleCloudflareWorker) Name() string {
	return "Cloudflare Worker"
}

// Create a new resource
func (r resourceRuleCloudflareWorker) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetCloudflareWorker](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleCloudflareWorker) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetCloudflareWorker](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleCloudflareWorker) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetCloudflareWorker](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleCloudflareWorker) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetCloudflareWorker](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleCloudflareWorker) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
