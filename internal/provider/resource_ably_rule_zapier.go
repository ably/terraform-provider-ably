// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleZapier struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleZapier{}
var _ resource.ResourceWithImportState = &ResourceRuleZapier{}

// Schema defines the schema for the resource.
func (r ResourceRuleZapier) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
		},
		"The `ably_rule_zapier` resource allows you to create and manage an Ably integration rule for Zapier. Read more at https://ably.com/docs/general/webhooks/zapier",
	)
}

func (r ResourceRuleZapier) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_zapier"
}

func (r *ResourceRuleZapier) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleZapier) Name() string {
	return "Zapier"
}

// Create creates a new resource.
func (r ResourceRuleZapier) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetZapier](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleZapier) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetZapier](&r, ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleZapier) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetZapier](&r, ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleZapier) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetZapier](&r, ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleZapier) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
