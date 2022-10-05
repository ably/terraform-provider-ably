package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRulePulsarType struct{}

// Get Rule Resource schema
func (r resourceRulePulsarType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"routing_key": {
				Type:        types.StringType,
				Required:    true,
				Description: "The optional routing key (partition key) used to publish messages. Supports interpolation as described in the [Ably FAQs](https://faqs.ably.com/what-is-the-format-of-the-routingkey-for-an-amqp-or-kinesis-reactor-rule).",
			},
			"topic": {
				Type:        types.StringType,
				Required:    true,
				Description: "A Pulsar topic. This is a named channel for transmission of messages between producers and consumers. The topic has the form: {persistent|non-persistent}://tenant/namespace/topic",
			},
			"service_url": {
				Type:        types.StringType,
				Required:    true,
				Description: "The URL of the Pulsar cluster in the form pulsar://host:port or pulsar+ssl://host:port",
			},
			"tls_trust_certs": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional:    true,
				Sensitive:   true,
				Description: "All connections to a Pulsar endpoint require TLS. The tls_trust_certs option allows you to configure different or additional trust anchors for those TLS connections. This enables server verification. You can specify an optional list of trusted CA certificates to use to verify the TLS certificate presented by the Pulsar cluster. Each certificate should be encoded in PEM format",
			},
			"enveloped": GetEnvelopedchema(),
			"format":    GetFormatSchema(),
			"authentication": {
				Required:    true,
				Description: "Pulsar supports authenticating clients using security tokens that are based on JSON Web Tokens.",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"mode": {
						Description: "Authentication mode, in this case JSON Web Token. Use `jwt`",
						Type:        types.StringType,
						Required:    true,
					},
					"token": {
						Description: "The JWT string.`",
						Type:        types.StringType,
						Required:    true,
					},
				}),
			},
		},
		"The `ably_rule_pulsar` resource allows you to create and manage an Ably integration rule for Pulsar. Read more at https://ably.com/docs/general/firehose/pulsar-rule",
	), nil
}

// New resource instance
func (r resourceRulePulsarType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceRulePulsar{
		p: *(p.(*provider)),
	}, nil
}

type resourceRulePulsar struct {
	p provider
}

func (r *resourceRulePulsar) Provider() *provider {
	return &r.p
}

func (r *resourceRulePulsar) Name() string {
	return "Pulsar"
}

// Create a new resource
func (r resourceRulePulsar) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetPulsar](&r, ctx, req, resp)
}

// Read resource
func (r resourceRulePulsar) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetPulsar](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRulePulsar) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetPulsar](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRulePulsar) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetPulsar](&r, ctx, req, resp)
}

// Import resource
func (r resourceRulePulsar) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportRule(&r, ctx, req, resp)
}
