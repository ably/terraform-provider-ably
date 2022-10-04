package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleZapierType struct{}

// Get Rule Resource schema
func (r resourceRuleZapierType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
		},
		"The `ably_rule_zapier` resource allows you to create and manage an Ably integration rule for Zapier. Read more at https://ably.com/docs/general/webhooks/zapier",
	), nil
}

// New resource instance
func (r resourceRuleZapierType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceRuleZapier{
		p: *(p.(*provider)),
	}, nil
}

type resourceRuleZapier struct {
	p provider
}

func (r *resourceRuleZapier) Provider() *provider {
	return &r.p
}

func (r *resourceRuleZapier) Name() string {
	return "Zapier"
}

// Create a new resource
func (r resourceRuleZapier) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetZapier](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleZapier) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetZapier](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleZapier) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetZapier](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleZapier) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetZapier](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleZapier) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportRule(&r, ctx, req, resp)
}
