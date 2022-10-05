package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleHTTPType struct{}

// Get Rule Resource schema
func (r resourceRuleHTTPType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"headers": GetHeaderSchema(),
			"url": {
				Type:        types.StringType,
				Required:    true,
				Description: "The webhook URL that Ably will POST events to",
			},
			"signing_key_id": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The signing key ID for use in batch mode. Ably will optionally sign the payload using an API key ensuring your servers can validate the payload using the private API key. See the [webhook security docs](https://ably.com/docs/general/webhooks#security) for more information",
			},
			"format": GetFormatSchema(),
		},
		"The `ably_rule_http` resource allows you to create and manage an Ably integration rule for HTTP. Read more at https://ably.com/docs/general/webhooks",
	), nil
}

// New resource instance
func (r resourceRuleHTTPType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceRuleHTTP{
		p: *(p.(*provider)),
	}, nil
}

type resourceRuleHTTP struct {
	p provider
}

func (r *resourceRuleHTTP) Provider() *provider {
	return &r.p
}

func (r *resourceRuleHTTP) Name() string {
	return "HTTP"
}

// Create a new resource
func (r resourceRuleHTTP) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetHTTP](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleHTTP) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetHTTP](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleHTTP) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetHTTP](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleHTTP) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetHTTP](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleHTTP) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
