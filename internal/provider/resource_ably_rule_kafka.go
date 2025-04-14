package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleKafka struct {
	p *AblyProvider
}

// Get Rule Resource schema
func (r resourceRuleKafka) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"routing_key": {
				Type:        types.StringType,
				Required:    true,
				Description: "The Kafka partition key. This is used to determine which partition a message should be routed to, where a topic has been partitioned. routingKey should be in the format topic:key where topic is the topic to publish to, and key is the value to use as the message key",
			},
			"enveloped": GetEnvelopedSchema(),
			"format":    GetFormatSchema(),
			"brokers": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Required:    true,
				Description: "This is a list of brokers that host your Kafka partitions. Each broker is specified using the format `host`, `host:port` or `ip:port`",
			},
			"auth": {
				Required:    true,
				Description: "The Kafka [authentication mechanism](https://docs.confluent.io/platform/current/kafka/overview-authentication-methods.html)",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"sasl": {
						Optional:    true,
						Description: "SASL(Simple Authentication Security Layer) / SCRAM (Salted Challenge Response Authentication Mechanism) uses usernames and passwords stored in ZooKeeper. Credentials are created during installation. See documentation on [configuring SCRAM](https://docs.confluent.io/platform/current/kafka/authentication_sasl/authentication_sasl_scram.html#kafka-sasl-auth-scram)",
						Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
							"mechanism": {
								Description: "`plain` `scram-sha-256` `scram-sha-512`. The hash type to use. SCRAM supports either SHA-256 or SHA-512 hash functions",
								Type:        types.StringType,
								Required:    true,
							},
							"username": {
								Description: "Kafka login credential",
								Type:        types.StringType,
								Required:    true,
								Sensitive:   true,
							},
							"password": {
								Description: "Kafka login credential",
								Type:        types.StringType,
								Required:    true,
								Sensitive:   true,
							},
						}),
					},
				}),
			},
		},
		"The `ably_rule_kafka` resource allows you to create and manage an Ably integration rule for Kafka. Read more at https://ably.com/docs/general/firehose/kafka-rule",
	), nil
}

func (r resourceRuleKafka) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_rule_kafka"
}

func (r *resourceRuleKafka) Provider() *AblyProvider {
	return r.p
}

func (r *resourceRuleKafka) Name() string {
	return "Kafka"
}

// Create a new resource
func (r resourceRuleKafka) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetKafka](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleKafka) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetKafka](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleKafka) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetKafka](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleKafka) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetKafka](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleKafka) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
