package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceIngressRuleMongo struct {
	p *provider
}

// Get Rule Resource schema
func (r resourceIngressRuleMongo) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetIngressRuleSchema(
		map[string]tfsdk.Attribute{
			"url": {
				Type:        types.StringType,
				Required:    true,
				Description: "The connection string of your MongoDB instance. (e.g. mongodb://user:pass@myhost.com)",
			},
			"database": {
				Type:        types.StringType,
				Required:    true,
				Description: "The MongoDB Database Name",
			},
			"collection": {
				Type:        types.StringType,
				Required:    true,
				Description: "What the connector should watch within the database. The connector only supports watching collections.",
			},
			"pipeline": {
				Type:        types.StringType,
				Required:    true,
				Description: "A MongoDB pipeline to pass to the Change Stream API. This field allows you to control which types of change events are published, and which channel the change event should be published to. The pipeline must set the _ablyChannel field on the root of the change event. It must also be a valid JSON array of pipeline operations.",
			},
			"full_document": {
				Type:        types.StringType,
				Required:    true,
				Description: "Controls whether the full document should be included in the published change events. Full Document is not available by default in all types of change event. Possible values are `updateLookup`, `whenAvailable`, `off`. The default is `off`.",
			},
			"full_document_before_change": {
				Type:        types.StringType,
				Required:    true,
				Description: "Controls whether the full document before the change should be included in the change event. Full Document before change is not available on all types of change event. Possible values are `whenAvailable` or `off`. The default is `off`.",
			},
			"primary_site": {
				Type:        types.StringType,
				Required:    true,
				Description: "The primary site that the connector will run in. You should choose a site that is close to your database.",
			},
		},
		"The `ably_ingress_rule_mongodb` resource sets up a MongoDB Integration Rule to stream document changes from a database collection over Ably."), nil
}

func (r resourceIngressRuleMongo) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_ingress_rule_mongodb"
}

func (r *resourceIngressRuleMongo) Provider() *provider {
	return r.p
}

func (r *resourceIngressRuleMongo) Name() string {
	return "MongoDB"
}

// Create a new resource
func (r resourceIngressRuleMongo) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateIngressRule[AblyIngressRuleTargetMongo](&r, ctx, req, resp)
}

// Read resource
func (r resourceIngressRuleMongo) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadIngressRule[AblyIngressRuleTargetMongo](&r, ctx, req, resp)
}

// Update resource
func (r resourceIngressRuleMongo) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateIngressRule[AblyIngressRuleTargetMongo](&r, ctx, req, resp)
}

// Delete resource
func (r resourceIngressRuleMongo) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteIngressRule[AblyIngressRuleTargetMongo](&r, ctx, req, resp)
}

// Import resource
func (r resourceIngressRuleMongo) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
