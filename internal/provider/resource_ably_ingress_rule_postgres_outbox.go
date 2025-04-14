package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceIngressRulePostgresOutbox struct {
	p *AblyProvider
}

// Get Rule Resource schema
func (r resourceIngressRulePostgresOutbox) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetIngressRuleSchema(
		map[string]tfsdk.Attribute{
			"url": {
				Type:        types.StringType,
				Required:    true,
				Description: "The URL for your Postgres database, for example postgres://user:password@example.com:5432/your-database-name. The associated user must have the correct privileges",
			},
			"outbox_table_schema": {
				Type:        types.StringType,
				Required:    true,
				Description: "Schema for the outbox table in your database, which allows for the reliable publication of an ordered sequence of change event messages over Ably.",
			},
			"outbox_table_name": {
				Type:        types.StringType,
				Required:    true,
				Description: "Name for the outbox table.",
			},
			"nodes_table_schema": {
				Type:        types.StringType,
				Required:    true,
				Description: "Schema for the nodes table in your database to allow for operation as a cluster to provide fault tolerance.",
			},
			"nodes_table_name": {
				Type:        types.StringType,
				Required:    true,
				Description: "Name for the nodes table.",
			},
			"ssl_mode": {
				Type:     types.StringType,
				Required: true,
				Description: `Determines the level of protection provided by the SSL connection. Options are:
  - prefer: Attempt SSL but allow non-SSL.
  - require: Always use SSL but don't verify certificates.
  - verify-ca: Verify server certificate is signed by a trusted CA.
  - verify-full: Verify server certificate and hostname.

Default: prefer.`,
			},
			"ssl_root_cert": {
				Type:        types.StringType,
				Optional:    true,
				Description: "Optional. Specifies the SSL certificate authority (CA) certificates. Required if SSL mode is verify-ca or verify-full.",
			},
			"primary_site": {
				Type:        types.StringType,
				Required:    true,
				Description: "The primary data center in which to run the integration rule.",
			},
		},
		"The `ably_ingress_rule_postgres_outbox` resource Use the Postgres database connector to distribute changes from your Postgres database to end users at scale. It enables you to distribute records using the outbox pattern to large numbers of subscribing clients, in realtime, as the changes occur."), nil
}

func (r resourceIngressRulePostgresOutbox) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_ingress_rule_postgres_outbox"
}

func (r *resourceIngressRulePostgresOutbox) Provider() *AblyProvider {
	return r.p
}

func (r *resourceIngressRulePostgresOutbox) Name() string {
	return "PostgresOutbox"
}

// Create a new resource
func (r resourceIngressRulePostgresOutbox) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateIngressRule[AblyIngressRuleTargetPostgresOutbox](&r, ctx, req, resp)
}

// Read resource
func (r resourceIngressRulePostgresOutbox) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadIngressRule[AblyIngressRuleTargetPostgresOutbox](&r, ctx, req, resp)
}

// Update resource
func (r resourceIngressRulePostgresOutbox) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateIngressRule[AblyIngressRuleTargetPostgresOutbox](&r, ctx, req, resp)
}

// Delete resource
func (r resourceIngressRulePostgresOutbox) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteIngressRule[AblyIngressRuleTargetPostgresOutbox](&r, ctx, req, resp)
}

// Import resource
func (r resourceIngressRulePostgresOutbox) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
