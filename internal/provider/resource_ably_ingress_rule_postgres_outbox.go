package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceIngressRulePostgresOutbox struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceIngressRulePostgresOutbox{}
var _ resource.ResourceWithImportState = &ResourceIngressRulePostgresOutbox{}

// Schema defines the schema for the resource.
func (r ResourceIngressRulePostgresOutbox) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetIngressRuleSchema(
		map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The URL for your Postgres database, for example postgres://user:password@example.com:5432/your-database-name. The associated user must have the correct privileges",
			},
			"outbox_table_schema": schema.StringAttribute{
				Required:    true,
				Description: "Schema for the outbox table in your database, which allows for the reliable publication of an ordered sequence of change event messages over Ably.",
			},
			"outbox_table_name": schema.StringAttribute{
				Required:    true,
				Description: "Name for the outbox table.",
			},
			"nodes_table_schema": schema.StringAttribute{
				Required:    true,
				Description: "Schema for the nodes table in your database to allow for operation as a cluster to provide fault tolerance.",
			},
			"nodes_table_name": schema.StringAttribute{
				Required:    true,
				Description: "Name for the nodes table.",
			},
			"ssl_mode": schema.StringAttribute{
				Required: true,
				Description: `Determines the level of protection provided by the SSL connection. Options are:
  - prefer: Attempt SSL but allow non-SSL.
  - require: Always use SSL but don't verify certificates.
  - verify-ca: Verify server certificate is signed by a trusted CA.
  - verify-full: Verify server certificate and hostname.

Default: prefer.`,
			},
			"ssl_root_cert": schema.StringAttribute{
				Optional:    true,
				Description: "Optional. Specifies the SSL certificate authority (CA) certificates. Required if SSL mode is verify-ca or verify-full.",
			},
			"primary_site": schema.StringAttribute{
				Required:    true,
				Description: "The primary data center in which to run the integration rule.",
			},
		},
		"The `ably_ingress_rule_postgres_outbox` resource Use the Postgres database connector to distribute changes from your Postgres database to end users at scale. It enables you to distribute records using the outbox pattern to large numbers of subscribing clients, in realtime, as the changes occur.")
}

func (r ResourceIngressRulePostgresOutbox) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_ingress_rule_postgres_outbox"
}

func (r *ResourceIngressRulePostgresOutbox) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceIngressRulePostgresOutbox) Name() string {
	return "PostgresOutbox"
}

// Create a new resource
func (r ResourceIngressRulePostgresOutbox) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateIngressRule[AblyIngressRuleTargetPostgresOutbox](&r, ctx, req, resp)
}

// Read resource
func (r ResourceIngressRulePostgresOutbox) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadIngressRule[AblyIngressRuleTargetPostgresOutbox](&r, ctx, req, resp)
}

// Update resource
func (r ResourceIngressRulePostgresOutbox) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateIngressRule[AblyIngressRuleTargetPostgresOutbox](&r, ctx, req, resp)
}

// Delete resource
func (r ResourceIngressRulePostgresOutbox) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteIngressRule[AblyIngressRuleTargetPostgresOutbox](&r, ctx, req, resp)
}

// Import resource
func (r ResourceIngressRulePostgresOutbox) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
