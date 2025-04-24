// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceRuleAMQPExternal struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceRuleAMQPExternal{}
var _ resource.ResourceWithImportState = &ResourceRuleAMQPExternal{}

// Schema defines the schema for the resource.
func (r ResourceRuleAMQPExternal) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetRuleSchema(
		map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The webhook URL that Ably will POST events to",
			},
			"routing_key": schema.StringAttribute{
				Required:    true,
				Description: "The Kafka partition key. This is used to determine which partition a message should be routed to, where a topic has been partitioned. routingKey should be in the format topic:key where topic is the topic to publish to, and key is the value to use as the message key",
			},
			"exchange": schema.StringAttribute{
				Required:    true,
				Description: "The RabbitMQ exchange, if needed, supports interpolation; see https://faqs.ably.com/what-is-the-format-of-the-routingkey-for-an-amqp-or-kinesis-reactor-rule for more info. If you don't use RabbitMQ exchanges, leave this blank.",
			},
			"mandatory_route": schema.BoolAttribute{
				Required:    true,
				Description: "Reject delivery of the message if the route does not exist, otherwise fail silently.",
			},
			"persistent_messages": schema.BoolAttribute{
				Required:    true,
				Description: "Marks the message as persistent, instructing the broker to write it to disk if it is in a durable queue.",
			},
			"message_ttl": schema.Int64Attribute{
				Optional:    true,
				Description: "You can optionally override the default TTL on a queue and specify a TTL in minutes for messages to be persisted. It is unusual to change the default TTL, so if this field is left empty, the default TTL for the queue will be used.",
			},
			"headers":   GetHeaderSchema(),
			"enveloped": GetEnvelopedSchema(),
			"format":    GetFormatSchema(),
		},
		"The `ably_rule_amqp_external` resource allows you to create and manage an Ably integration rule for Firehose. Read more at https://ably.com/docs/general/firehose")
}

func (r ResourceRuleAMQPExternal) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_rule_amqp_external"
}

func (r *ResourceRuleAMQPExternal) Provider() *AblyProvider {
	return r.p
}

func (r *ResourceRuleAMQPExternal) Name() string {
	return "AMQP External"
}

// Create creates a new resource.
func (r ResourceRuleAMQPExternal) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	CreateRule[AblyRuleTargetAMQPExternal](&r, ctx, req, resp)
}

// Read resource
func (r ResourceRuleAMQPExternal) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ReadRule[AblyRuleTargetAMQPExternal](&r, ctx, req, resp)
}

// Update updates an existing resource.
func (r ResourceRuleAMQPExternal) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetAMQPExternal](&r, ctx, req, resp)
}

// Delete deletes the resource.
func (r ResourceRuleAMQPExternal) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetAMQPExternal](&r, ctx, req, resp)
}

// ImportState handles the import state functionality.
func (r ResourceRuleAMQPExternal) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}
