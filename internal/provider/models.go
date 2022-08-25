package ably_control

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ably App
type AblyApp struct {
	AccountID types.String `tfsdk:"account_id"`
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Status    types.String `tfsdk:"status"`
	TLSOnly   types.Bool   `tfsdk:"tls_only"`
}

// Ably Namespace
type AblyNamespace struct {
	AppID            types.String `tfsdk:"app_id"`
	ID               types.String `tfsdk:"id"`
	Authenticated    types.Bool   `tfsdk:"authenticated"`
	Persisted        types.Bool   `tfsdk:"persisted"`
	PersistLast      types.Bool   `tfsdk:"persist_last"`
	PushEnabled      types.Bool   `tfsdk:"push_enabled"`
	TlsOnly          types.Bool   `tfsdk:"tls_only"`
	ExposeTimeserial types.Bool   `tfsdk:"expose_timeserial"`
}

// Ably Key
type AblyKey struct {
	ID         types.String        `tfsdk:"id"`
	AppID      types.String        `tfsdk:"app_id"`
	Name       types.String        `tfsdk:"name"`
	Capability map[string][]string `tfsdk:"capabilities"`
	Status     types.Int64         `tfsdk:"status"`
	Key        types.String        `tfsdk:"key"`
	Created    types.Int64         `tfsdk:"created"`
	Modified   types.Int64         `tfsdk:"modified"`
}

// Ably Queue
type AblyQueue struct {
	AppID     types.String `tfsdk:"app_id"`
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Ttl       types.Int64  `tfsdk:"ttl"`
	MaxLength types.Int64  `tfsdk:"max_length"`
	Region    types.String `tfsdk:"region"`

	AmqpUri                  types.String `tfsdk:"amqp_uri"`
	AmqpQueueName            types.String `tfsdk:"amqp_queue_name"`
	StompURI                 types.String `tfsdk:"stomp_uri"`
	StompHost                types.String `tfsdk:"stomp_host"`
	StompDestination         types.String `tfsdk:"stomp_destination"`
	State                    types.String `tfsdk:"state"`
	MessagesReady            types.Int64  `tfsdk:"messages_ready"`
	MessagesUnacknowledged   types.Int64  `tfsdk:"messages_unacknowledged"`
	MessagesTotal            types.Int64  `tfsdk:"messages_total"`
	StatsPublishRate         types.Int64  `tfsdk:"stats_publish_rate"`
	StatsDeliveryRate        types.Int64  `tfsdk:"stats_delivery_rate"`
	StatsAcknowledgementRate types.Int64  `tfsdk:"stats_acknowledgement_rate"`
	Deadletter               types.Bool   `tfsdk:"deadletter"`
	DeadletterID             types.String `tfsdk:"deadletter_id"`
}
