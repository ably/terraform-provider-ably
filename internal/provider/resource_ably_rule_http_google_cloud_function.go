package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleGoogleFunction struct {
	p *provider
}

// Get Rule Resource schema
func (r resourceRuleGoogleFunction) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"region": {
				Type:        types.StringType,
				Required:    true,
				Description: "The region in which your Google Cloud Function is hosted. See the Google documentation for more details.",
			},
			"function_name": {
				Type:        types.StringType,
				Required:    true,
				Description: "The name of your Google Cloud Function.",
			},
			"project_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The project ID for your Google Cloud Project that was generated when you created your project.",
			},
			"headers": GetHeaderSchema(),
			"signing_key_id": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The signing key ID for use in batch mode. Ably will optionally sign the payload using an API key ensuring your servers can validate the payload using the private API key. See the [webhook security docs](https://ably.com/docs/general/webhooks#security) for more information",
			},
			"enveloped": GetEnvelopedchema(),
			"format":    GetFormatSchema(),
		},
		"The `ably_rule_google_cloud_function` resource allows you to create and manage an Ably integration rule for Google cloud functions. Read more at https://ably.com/docs/general/webhooks/google-functions",
	), nil
}

func (r resourceRuleGoogleFunction) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_rule_google_function"
}

func (r *resourceRuleGoogleFunction) Provider() *provider {
	return r.p
}

func (r *resourceRuleGoogleFunction) Name() string {
	return "Google Cloud Function"
}

// Create a new resource
func (r resourceRuleGoogleFunction) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetGoogleFunction](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleGoogleFunction) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetGoogleFunction](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleGoogleFunction) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetGoogleFunction](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleGoogleFunction) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetGoogleFunction](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleGoogleFunction) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
