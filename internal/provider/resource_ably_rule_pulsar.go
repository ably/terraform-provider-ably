// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceRulePulsar struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRulePulsar{}
var _ resource.ResourceWithImportState = &ResourceRulePulsar{}

// Schema defines the schema for the resource.
func (r ResourceRulePulsar) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"routing_key": schema.StringAttribute{
				Required:    true,
				Description: "The optional routing key (partition key) used to publish messages. Supports interpolation as described in the [Ably FAQs](https://faqs.ably.com/what-is-the-format-of-the-routingkey-for-an-amqp-or-kinesis-reactor-rule).",
			},
			"topic": schema.StringAttribute{
				Required:    true,
				Description: "A Pulsar topic. This is a named channel for transmission of messages between producers and consumers. The topic has the form: {persistent|non-persistent}://tenant/namespace/topic",
			},
			"service_url": schema.StringAttribute{
				Required:    true,
				Description: "The URL of the Pulsar cluster in the form pulsar://host:port or pulsar+ssl://host:port",
			},
			"tls_trust_certs": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Sensitive:   true,
				Description: "All connections to a Pulsar endpoint require TLS. The tls_trust_certs option allows you to configure different or additional trust anchors for those TLS connections. This enables server verification. You can specify an optional list of trusted CA certificates to use to verify the TLS certificate presented by the Pulsar cluster. Each certificate should be encoded in PEM format",
			},
			"enveloped": GetEnvelopedSchema(),
			"format":    GetFormatSchema(),
			"authentication": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Pulsar supports authenticating clients using security tokens that are based on JSON Web Tokens.",
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "Authentication mode, in this case JSON Web Token. Use `jwt`",
						Required:    true,
					},
					"token": schema.StringAttribute{
						Description: "The JWT string.`",
						Required:    true,
					},
				},
			},
		},
		"The `ably_rule_pulsar` resource allows you to create and manage an Ably integration rule for Pulsar. Read more at https://ably.com/docs/general/firehose/pulsar-rule",
	)
}

func (r ResourceRulePulsar) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_pulsar"
}

func (r *ResourceRulePulsar) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRulePulsar) Name() string {
	return "Pulsar"
}

// Create creates a new resource.
func (r ResourceRulePulsar) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetPulsar](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRulePulsar) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetPulsar](&r, ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRulePulsar) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetPulsar](&r, ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRulePulsar) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetPulsar](&r, ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRulePulsar) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
