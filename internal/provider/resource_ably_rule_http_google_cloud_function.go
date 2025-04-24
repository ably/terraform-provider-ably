package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleGoogleFunction struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleGoogleFunction{}
var _ resource.ResourceWithImportState = &ResourceRuleGoogleFunction{}

// Get Rule Resource schema
func (r ResourceRuleGoogleFunction) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Required:    true,
				Description: "The region in which your Google Cloud Function is hosted. See the Google documentation for more details.",
			},
			"function_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of your Google Cloud Function.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The project ID for your Google Cloud Project that was generated when you created your project.",
			},
			"headers": GetHeaderSchema(),
			"signing_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "The signing key ID for use in batch mode. Ably will optionally sign the payload using an API key ensuring your servers can validate the payload using the private API key. See the [webhook security docs](https://ably.com/docs/general/webhooks#security) for more information",
			},
			"enveloped": GetEnvelopedSchema(),
			"format":    GetFormatSchema(),
		},
		"The `ably_rule_google_cloud_function` resource allows you to create and manage an Ably integration rule for Google cloud functions. Read more at https://ably.com/docs/general/webhooks/google-functions",
	)
}

func (r ResourceRuleGoogleFunction) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_google_function"
}

func (r *ResourceRuleGoogleFunction) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleGoogleFunction) Name() string {
	return "Google Cloud Function"
}

// Create a new resource
func (r ResourceRuleGoogleFunction) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetGoogleFunction](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleGoogleFunction) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetGoogleFunction](&r, ctx, req, resp)
}

// // Update resource
func (r ResourceRuleGoogleFunction) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetGoogleFunction](&r, ctx, req, resp)
}

// Delete resource
func (r ResourceRuleGoogleFunction) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetGoogleFunction](&r, ctx, req, resp)
}

// Import resource
func (r ResourceRuleGoogleFunction) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
