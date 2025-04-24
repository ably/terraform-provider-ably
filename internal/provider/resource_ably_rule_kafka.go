// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceRuleKafka struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleKafka{}
var _ resource.ResourceWithImportState = &ResourceRuleKafka{}

// Schema defines the schema for the resource.
func (r ResourceRuleKafka) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"routing_key": schema.StringAttribute{
				Required:    true,
				Description: "The Kafka partition key. This is used to determine which partition a message should be routed to, where a topic has been partitioned. routingKey should be in the format topic:key where topic is the topic to publish to, and key is the value to use as the message key",
			},
			"enveloped": GetEnvelopedSchema(),
			"format":    GetFormatSchema(),
			"brokers": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "This is a list of brokers that host your Kafka partitions. Each broker is specified using the format `host`, `host:port` or `ip:port`",
			},
			"auth": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The Kafka [authentication mechanism](https://docs.confluent.io/platform/current/kafka/overview-authentication-methods.html)",
				Attributes: map[string]schema.Attribute{
					"sasl": schema.SingleNestedAttribute{
						Optional:    true,
						Description: "SASL(Simple Authentication Security Layer) / SCRAM (Salted Challenge Response Authentication Mechanism) uses usernames and passwords stored in ZooKeeper. Credentials are created during installation. See documentation on [configuring SCRAM](https://docs.confluent.io/platform/current/kafka/authentication_sasl/authentication_sasl_scram.html#kafka-sasl-auth-scram)",
						Attributes: map[string]schema.Attribute{
							"mechanism": schema.StringAttribute{
								Description: "`plain` `scram-sha-256` `scram-sha-512`. The hash type to use. SCRAM supports either SHA-256 or SHA-512 hash functions",
								Required:    true,
							},
							"username": schema.StringAttribute{
								Description: "Kafka login credential",
								Required:    true,
								Sensitive:   true,
							},
							"password": schema.StringAttribute{
								Description: "Kafka login credential",
								Required:    true,
								Sensitive:   true,
							},
						},
					},
				},
			},
		},
		"The `ably_rule_kafka` resource allows you to create and manage an Ably integration rule for Kafka. Read more at https://ably.com/docs/general/firehose/kafka-rule",
	)
}

func (r ResourceRuleKafka) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_kafka"
}

func (r *ResourceRuleKafka) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleKafka) Name() string {
	return "Kafka"
}

// Create creates a new resource.
func (r ResourceRuleKafka) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetKafka](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleKafka) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetKafka](&r, ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleKafka) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetKafka](&r, ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleKafka) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetKafka](&r, ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleKafka) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
