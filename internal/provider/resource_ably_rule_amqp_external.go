package ably_control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceRuleAmqpExternal struct {
	p *provider
}

// Get Rule Resource schema
func (r resourceRuleAmqpExternal) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return GetRuleSchema(
		map[string]tfsdk.Attribute{
			"url": {
				Type:        types.StringType,
				Required:    true,
				Description: "The webhook URL that Ably will POST events to",
			},
			"routing_key": {
				Type:        types.StringType,
				Required:    true,
				Description: "The Kafka partition key. This is used to determine which partition a message should be routed to, where a topic has been partitioned. routingKey should be in the format topic:key where topic is the topic to publish to, and key is the value to use as the message key",
			},
			"mandatory_route": {
				Type:        types.BoolType,
				Required:    true,
				Description: "Reject delivery of the message if the route does not exist, otherwise fail silently.",
			},
			"persistent_messages": {
				Type:        types.BoolType,
				Required:    true,
				Description: "Marks the message as persistent, instructing the broker to write it to disk if it is in a durable queue.",
			},
			"message_ttl": {
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				Description: "You can optionally override the default TTL on a queue and specify a TTL in minutes for messages to be persisted. It is unusual to change the default TTL, so if this field is left empty, the default TTL for the queue will be used.",
			},
			"headers":   GetHeaderSchema(),
			"enveloped": GetEnvelopedchema(),
			"format":    GetFormatSchema(),
		},
		"The `ably_rule_amqp_external` resource allows you to create and manage an Ably integration rule for Firehose. Read more at https://ably.com/docs/general/firehose",
	), nil
}

func (r resourceRuleAmqpExternal) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_rule_amqp_external"
}

func (r *resourceRuleAmqpExternal) Provider() *provider {
	return r.p
}

func (r *resourceRuleAmqpExternal) Name() string {
	return "AMQP External"
}

// Create a new resource
func (r resourceRuleAmqpExternal) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	CreateRule[AblyRuleTargetAmqpExternal](&r, ctx, req, resp)
}

// Read resource
func (r resourceRuleAmqpExternal) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	ReadRule[AblyRuleTargetAmqpExternal](&r, ctx, req, resp)
}

// // Update resource
func (r resourceRuleAmqpExternal) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	UpdateRule[AblyRuleTargetAmqpExternal](&r, ctx, req, resp)
}

// Delete resource
func (r resourceRuleAmqpExternal) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	DeleteRule[AblyRuleTargetAmqpExternal](&r, ctx, req, resp)
}

// Import resource
func (r resourceRuleAmqpExternal) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")

}
