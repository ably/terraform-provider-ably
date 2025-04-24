// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceIngressRuleMongo struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceIngressRuleMongo{}
var _ resource.ResourceWithImportState = &ResourceIngressRuleMongo{}

// Schema defines the schema for the resource.
func (r ResourceIngressRuleMongo) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetIngressRuleSchema(
		map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The connection string of your MongoDB instance. (e.g. mongodb://user:pass@myhost.com)",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "The MongoDB Database Name",
			},
			"collection": schema.StringAttribute{
				Required:    true,
				Description: "What the connector should watch within the database. The connector only supports watching collections.",
			},
			"pipeline": schema.StringAttribute{
				Required:    true,
				Description: "A MongoDB pipeline to pass to the Change Stream API. This field allows you to control which types of change events are published, and which channel the change event should be published to. The pipeline must set the _ablyChannel field on the root of the change event. It must also be a valid JSON array of pipeline operations.",
			},
			"full_document": schema.StringAttribute{
				Required:    true,
				Description: "Controls whether the full document should be included in the published change events. Full Document is not available by default in all types of change event. Possible values are `updateLookup`, `whenAvailable`, `off`. The default is `off`.",
			},
			"full_document_before_change": schema.StringAttribute{
				Required:    true,
				Description: "Controls whether the full document before the change should be included in the change event. Full Document before change is not available on all types of change event. Possible values are `whenAvailable` or `off`. The default is `off`.",
			},
			"primary_site": schema.StringAttribute{
				Required:    true,
				Description: "The primary site that the connector will run in. You should choose a site that is close to your database.",
			},
		},
		"The `ably_ingress_rule_mongodb` resource sets up a MongoDB Integration Rule to stream document changes from a database collection over Ably.")
}

func (r ResourceIngressRuleMongo) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_ingress_rule_mongodb"
}

func (r *ResourceIngressRuleMongo) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceIngressRuleMongo) Name() string {
	return "MongoDB"
}

// Create creates a new resource.
func (r ResourceIngressRuleMongo) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateIngressRule[AblyIngressRuleTargetMongo](&r, ctx, req, resp)
}

// Read reads the resource.
func (r ResourceIngressRuleMongo) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadIngressRule[AblyIngressRuleTargetMongo](&r, ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceIngressRuleMongo) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateIngressRule[AblyIngressRuleTargetMongo](&r, ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceIngressRuleMongo) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteIngressRule[AblyIngressRuleTargetMongo](&r, ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceIngressRuleMongo) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
