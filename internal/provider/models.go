package ably_control

import (
	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ably App
type AblyApp struct {
	AccountID              types.String `tfsdk:"account_id"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Status                 types.String `tfsdk:"status"`
	TLSOnly                types.Bool   `tfsdk:"tls_only"`
	FcmKey                 types.String `tfsdk:"fcm_key"`
	ApnsCertificate        types.String `tfsdk:"apns_certificate"`
	ApnsPrivateKey         types.String `tfsdk:"apns_private_key"`
	ApnsUseSandboxEndpoint types.Bool   `tfsdk:"apns_use_sandbox_endpoint"`
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

func emptyStringToNull(v *types.String) {
	if v.Value == "" {
		v.Null = true
	}
}

// Ably Rule
type AblyRuleSource struct {
	ChannelFilter types.String               `tfsdk:"channel_filter"`
	Type          ably_control_go.SourceType `tfsdk:"type"`
}

func (r *AblyRuleDecoder[_]) Rule() AblyRule {
	return AblyRule{
		ID:          r.ID,
		AppID:       r.AppID,
		Status:      r.Status,
		RequestMode: r.RequestMode,
		Source:      r.Source,
		Target:      r.Target,
	}
}

type AblyRuleDecoder[T any] struct {
	ID          types.String   `tfsdk:"id"`
	AppID       types.String   `tfsdk:"app_id"`
	Status      types.String   `tfsdk:"status"`
	RequestMode types.String   `tfsdk:"request_mode"`
	Source      AblyRuleSource `tfsdk:"source"`
	Target      T              `tfsdk:"target"`
}

type AblyRule AblyRuleDecoder[any]

type AblyRuleTargetKinesis struct {
	Region       string                 `tfsdk:"region"`
	StreamName   string                 `tfsdk:"stream_name"`
	PartitionKey string                 `tfsdk:"partition_key"`
	AwsAuth      AwsAuth                `tfsdk:"authentication"`
	Enveloped    bool                   `tfsdk:"enveloped"`
	Format       ably_control_go.Format `tfsdk:"format"`
}

type AwsAuth struct {
	AuthenticationMode types.String `tfsdk:"mode"`
	RoleArn            types.String `tfsdk:"role_arn"`
	AccessKeyId        types.String `tfsdk:"access_key_id"`
	SecretAccessKey    types.String `tfsdk:"secret_access_key"`
}

type AblyRuleTargetSqs struct {
	Region       string                 `tfsdk:"region"`
	AwsAccountID string                 `tfsdk:"aws_account_id"`
	QueueName    string                 `tfsdk:"queue_name"`
	AwsAuth      AwsAuth                `tfsdk:"authentication"`
	Enveloped    bool                   `tfsdk:"enveloped"`
	Format       ably_control_go.Format `tfsdk:"format"`
}

type AblyRuleTargetLambda struct {
	Region       string  `tfsdk:"region"`
	FunctionName string  `tfsdk:"function_name"`
	AwsAuth      AwsAuth `tfsdk:"authentication"`
	Enveloped    bool    `tfsdk:"enveloped"`
}

type AblyRuleTargetGoogleFunction struct {
	Region       string                 `tfsdk:"region"`
	ProjectID    string                 `tfsdk:"project_id"`
	FunctionName string                 `tfsdk:"function_name"`
	Headers      []AblyRuleHeaders      `tfsdk:"headers"`
	SigningKeyId string                 `tfsdk:"signing_key_id"`
	Enveloped    bool                   `tfsdk:"enveloped"`
	Format       ably_control_go.Format `tfsdk:"format"`
}

type AblyRuleTargetCloudflareWorker struct {
	Url          string            `tfsdk:"url"`
	Headers      []AblyRuleHeaders `tfsdk:"headers"`
	SigningKeyId string            `tfsdk:"signing_key_id"`
}

type AblyRuleTargetHTTP struct {
	Url          string                 `tfsdk:"url"`
	Headers      []AblyRuleHeaders      `tfsdk:"headers"`
	SigningKeyId string                 `tfsdk:"signing_key_id"`
	Format       ably_control_go.Format `tfsdk:"format"`
}

type AblyRuleTargetPulsar struct {
	RoutingKey     string                 `tfsdk:"routing_key"`
	Topic          string                 `tfsdk:"topic"`
	ServiceURL     string                 `tfsdk:"service_url"`
	TlsTrustCerts  []string               `tfsdk:"tls_trust_certs"`
	Authentication PulsarAuthentication   `tfsdk:"authentication"`
	Enveloped      bool                   `tfsdk:"enveloped"`
	Format         ably_control_go.Format `tfsdk:"format"`
}

type PulsarAuthentication struct {
	Mode  string `tfsdk:"mode"`
	Token string `tfsdk:"token"`
}

type AblyRuleTargetZapier struct {
	Url          string            `tfsdk:"url"`
	Headers      []AblyRuleHeaders `tfsdk:"headers"`
	SigningKeyId string            `tfsdk:"signing_key_id"`
}

type AblyRuleTargetIFTTT struct {
	WebhookKey string `tfsdk:"webhook_key"`
	EventName  string `tfsdk:"event_name"`
}

type AblyRuleTargetAzureFunction struct {
	AzureAppID        string                 `tfsdk:"azure_app_id"`
	AzureFunctionName string                 `tfsdk:"function_name"`
	Headers           []AblyRuleHeaders      `tfsdk:"headers"`
	SigningKeyID      string                 `tfsdk:"signing_key_id"`
	Format            ably_control_go.Format `tfsdk:"format"`
}

type AblyRuleHeaders struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type AblyRuleTargetKafka struct {
	RoutingKey          string                 `tfsdk:"routing_key"`
	Brokers             []string               `tfsdk:"brokers"`
	KafkaAuthentication KafkaAuthentication    `tfsdk:"auth"`
	Enveloped           bool                   `tfsdk:"enveloped"`
	Format              ably_control_go.Format `tfsdk:"format"`
}

type KafkaAuthentication struct {
	Sasl Sasl `tfsdk:"sasl"`
}

type Sasl struct {
	Mechanism string `tfsdk:"mechanism"`
	Username  string `tfsdk:"username"`
	Password  string `tfsdk:"password"`
}
