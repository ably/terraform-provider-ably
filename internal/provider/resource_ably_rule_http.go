package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleHTTP struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleHTTP{}
var _ resource.ResourceWithImportState = &ResourceRuleHTTP{}

// Schema defines the schema for the resource.
func (r ResourceRuleHTTP) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"headers": GetHeaderSchema(),
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The webhook URL that Ably will POST events to",
			},
			"signing_key_id": schema.StringAttribute{
				Optional:    true,
				Description: "The signing key ID for use in batch mode. Ably will optionally sign the payload using an API key ensuring your servers can validate the payload using the private API key. See the [webhook security docs](https://ably.com/docs/general/webhooks#security) for more information",
			},
			"format":    GetFormatSchema(),
			"enveloped": GetEnvelopedSchema(),
		},
		"The `ably_rule_http` resource allows you to create and manage an Ably integration rule for HTTP. Read more at https://ably.com/docs/general/webhooks")
}

func (r ResourceRuleHTTP) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_http"
}

func (r *ResourceRuleHTTP) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleHTTP) Name() string {
	return "HTTP"
}

// Create a new resource
func (r ResourceRuleHTTP) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetHTTP](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleHTTP) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetHTTP](&r, ctx, req, resp)
}

// // Update resource
func (r ResourceRuleHTTP) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetHTTP](&r, ctx, req, resp)
}

// Delete resource
func (r ResourceRuleHTTP) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetHTTP](&r, ctx, req, resp)
}

// Import resource
func (r ResourceRuleHTTP) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
