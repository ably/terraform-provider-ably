// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AblyApp represents an Ably application.
type AblyApp struct {
	AccountID              types.String `tfsdk:"account_id"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Status                 types.String `tfsdk:"status"`
	TLSOnly                types.Bool   `tfsdk:"tls_only"`
	FcmKey                 types.String `tfsdk:"fcm_key"`
	FcmServiceAccount      types.String `tfsdk:"fcm_service_account"`
	FcmProjectId           types.String `tfsdk:"fcm_project_id"`
	ApnsCertificate        types.String `tfsdk:"apns_certificate"`
	ApnsPrivateKey         types.String `tfsdk:"apns_private_key"`
	ApnsUseSandboxEndpoint types.Bool   `tfsdk:"apns_use_sandbox_endpoint"`
}

// AblyNamespace represents an Ably namespace.
type AblyNamespace struct {
	AppID              types.String `tfsdk:"app_id"`
	ID                 types.String `tfsdk:"id"`
	Authenticated      types.Bool   `tfsdk:"authenticated"`
	Persisted          types.Bool   `tfsdk:"persisted"`
	PersistLast        types.Bool   `tfsdk:"persist_last"`
	PushEnabled        types.Bool   `tfsdk:"push_enabled"`
	TlsOnly            types.Bool   `tfsdk:"tls_only"`
	ExposeTimeserial   types.Bool   `tfsdk:"expose_timeserial"`
	BatchingEnabled    types.Bool   `tfsdk:"batching_enabled"`
	BatchingInterval   types.Int64  `tfsdk:"batching_interval"`
	ConflationEnabled  types.Bool   `tfsdk:"conflation_enabled"`
	ConflationInterval types.Int64  `tfsdk:"conflation_interval"`
	ConflationKey      types.String `tfsdk:"conflation_key"`
}

// AblyKey represents an Ably API key.
type AblyKey struct {
	ID              types.String         `tfsdk:"id"`
	AppID           types.String         `tfsdk:"app_id"`
	Name            types.String         `tfsdk:"name"`
	RevocableTokens types.Bool           `tfsdk:"revocable_tokens"`
	Capability      map[string]types.Set `tfsdk:"capabilities"`
	Status          types.Int64          `tfsdk:"status"`
	Key             types.String         `tfsdk:"key"`
	Created         types.Int64          `tfsdk:"created"`
	Modified        types.Int64          `tfsdk:"modified"`
}

// AblyQueue represents an Ably queue.
type AblyQueue struct {
	AppID     types.String `tfsdk:"app_id"`
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Ttl       types.Int64  `tfsdk:"ttl"`
	MaxLength types.Int64  `tfsdk:"max_length"`
	Region    types.String `tfsdk:"region"`

	AmqpUri                  types.String  `tfsdk:"amqp_uri"`
	AmqpQueueName            types.String  `tfsdk:"amqp_queue_name"`
	StompURI                 types.String  `tfsdk:"stomp_uri"`
	StompHost                types.String  `tfsdk:"stomp_host"`
	StompDestination         types.String  `tfsdk:"stomp_destination"`
	State                    types.String  `tfsdk:"state"`
	MessagesReady            types.Int64   `tfsdk:"messages_ready"`
	MessagesUnacknowledged   types.Int64   `tfsdk:"messages_unacknowledged"`
	MessagesTotal            types.Int64   `tfsdk:"messages_total"`
	StatsPublishRate         types.Float64 `tfsdk:"stats_publish_rate"`
	StatsDeliveryRate        types.Float64 `tfsdk:"stats_delivery_rate"`
	StatsAcknowledgementRate types.Float64 `tfsdk:"stats_acknowledgement_rate"`
	Deadletter               types.Bool    `tfsdk:"deadletter"`
	DeadletterID             types.String  `tfsdk:"deadletter_id"`
}

func emptyStringToNull(v *types.String) {
	if v.ValueString() == "" {
		*v = types.StringNull()
	}
}

func sliceString(v []types.String) []string {
	s := make([]string, len(v))
	for i, v := range v {
		s[i] = v.ValueString()
	}
	return s
}

// mapFromStringSlice converts a map of strings to string slices (Terraform types) to a map of Go strings to Go string slices
func mapFromStringSlice(m map[string][]types.String) map[string][]string {
	result := make(map[string][]string, len(m))
	for k, v := range m {
		result[k] = sliceString(v)
	}
	return result
}

// mapFromSet converts a map of string sets (Terraform types) to a map of Go strings to Go string slices
func mapFromSet(ctx context.Context, m map[string]types.Set) map[string][]string {
	result := make(map[string][]string, len(m))
	for k, v := range m {
		var slice []string
		if !v.IsNull() && !v.IsUnknown() {
			var elems []types.String
			v.ElementsAs(ctx, &elems, false)
			slice = sliceString(elems)
		}
		result[k] = slice
	}
	return result
}

// mapToTypedStringSlice converts a map of Go strings to Go string slices to a map of Terraform types
func mapToTypedStringSlice(m map[string][]string) map[string][]types.String {
	result := make(map[string][]types.String, len(m))
	for k, v := range m {
		typedSlice := make([]types.String, len(v))
		for i, s := range v {
			typedSlice[i] = types.StringValue(s)
		}
		result[k] = typedSlice
	}
	return result
}

// mapToTypedSet converts a map of Go strings to Go string slices to a map of Terraform types
func mapToTypedSet(m map[string][]string) map[string]types.Set {
	result := make(map[string]types.Set, len(m))
	for k, v := range m {
		typedSlice := toTypedStringSlice(v)
		attrValues := make([]attr.Value, len(typedSlice))
		for i, v := range typedSlice {
			attrValues[i] = v
		}
		result[k] = types.SetValueMust(types.StringType, attrValues)
	}
	return result
}

// toTypedStringSlice converts a slice of Go strings to a slice of Terraform types.String
func toTypedStringSlice(slice []string) []types.String {
	typedSlice := make([]types.String, len(slice))
	for i, s := range slice {
		typedSlice[i] = types.StringValue(s)
	}
	return typedSlice
}

// IngressRule returns the ingress rule from the decoder.
func (r *AblyIngressRuleDecoder[_]) IngressRule() AblyIngressRule {
	return AblyIngressRule{
		ID:     r.ID,
		AppID:  r.AppID,
		Status: r.Status,
		Target: r.Target,
	}
}

type AblyIngressRuleDecoder[T any] struct {
	ID     types.String `tfsdk:"id"`
	AppID  types.String `tfsdk:"app_id"`
	Status types.String `tfsdk:"status"`
	Target T            `tfsdk:"target"`
}

type AblyIngressRule AblyIngressRuleDecoder[any]

type AblyIngressRuleTargetMongo struct {
	Url                      types.String `tfsdk:"url"`
	Database                 types.String `tfsdk:"database"`
	Collection               types.String `tfsdk:"collection"`
	Pipeline                 types.String `tfsdk:"pipeline"`
	FullDocument             types.String `tfsdk:"full_document"`
	FullDocumentBeforeChange types.String `tfsdk:"full_document_before_change"`
	PrimarySite              types.String `tfsdk:"primary_site"`
}

type AblyIngressRuleTargetPostgresOutbox struct {
	Url               types.String `tfsdk:"url"`
	OutboxTableSchema types.String `tfsdk:"outbox_table_schema"`
	OutboxTableName   types.String `tfsdk:"outbox_table_name"`
	NodesTableSchema  types.String `tfsdk:"nodes_table_schema"`
	NodesTableName    types.String `tfsdk:"nodes_table_name"`
	SslMode           types.String `tfsdk:"ssl_mode"`
	SslRootCert       types.String `tfsdk:"ssl_root_cert"`
	PrimarySite       types.String `tfsdk:"primary_site"`
}

// AblyRuleSource represents a source for Ably rules.
type AblyRuleSource struct {
	ChannelFilter types.String `tfsdk:"channel_filter"`
	Type          types.String `tfsdk:"type"`
}

// Rule returns the rule from the decoder.
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
	ID          types.String    `tfsdk:"id"`
	AppID       types.String    `tfsdk:"app_id"`
	Status      types.String    `tfsdk:"status"`
	RequestMode types.String    `tfsdk:"request_mode"`
	Source      *AblyRuleSource `tfsdk:"source"`
	Target      T               `tfsdk:"target"`
}

type AblyRule AblyRuleDecoder[any]

type AblyRuleTargetKinesis struct {
	Region       types.String `tfsdk:"region"`
	StreamName   types.String `tfsdk:"stream_name"`
	PartitionKey types.String `tfsdk:"partition_key"`
	AwsAuth      AwsAuth      `tfsdk:"authentication"`
	Enveloped    types.Bool   `tfsdk:"enveloped"`
	Format       types.String `tfsdk:"format"`
}

type AwsAuth struct {
	AuthenticationMode types.String `tfsdk:"mode"`
	RoleArn            types.String `tfsdk:"role_arn"`
	AccessKeyId        types.String `tfsdk:"access_key_id"`
	SecretAccessKey    types.String `tfsdk:"secret_access_key"`
}

type AblyRuleTargetSqs struct {
	Region       types.String `tfsdk:"region"`
	AwsAccountID types.String `tfsdk:"aws_account_id"`
	QueueName    types.String `tfsdk:"queue_name"`
	AwsAuth      AwsAuth      `tfsdk:"authentication"`
	Enveloped    types.Bool   `tfsdk:"enveloped"`
	Format       types.String `tfsdk:"format"`
}

type AblyRuleTargetLambda struct {
	Region       types.String `tfsdk:"region"`
	FunctionName types.String `tfsdk:"function_name"`
	AwsAuth      AwsAuth      `tfsdk:"authentication"`
	Enveloped    types.Bool   `tfsdk:"enveloped"`
}

type AblyRuleTargetGoogleFunction struct {
	Region       types.String      `tfsdk:"region"`
	ProjectID    types.String      `tfsdk:"project_id"`
	FunctionName types.String      `tfsdk:"function_name"`
	Headers      []AblyRuleHeaders `tfsdk:"headers"`
	SigningKeyId types.String      `tfsdk:"signing_key_id"`
	Enveloped    types.Bool        `tfsdk:"enveloped"`
	Format       types.String      `tfsdk:"format"`
}

type AblyRuleTargetCloudflareWorker struct {
	Url          types.String      `tfsdk:"url"`
	Headers      []AblyRuleHeaders `tfsdk:"headers"`
	SigningKeyId types.String      `tfsdk:"signing_key_id"`
}

type AblyRuleTargetHTTP struct {
	Url          types.String      `tfsdk:"url"`
	Headers      []AblyRuleHeaders `tfsdk:"headers"`
	SigningKeyId types.String      `tfsdk:"signing_key_id"`
	Format       types.String      `tfsdk:"format"`
	Enveloped    types.Bool        `tfsdk:"enveloped"`
}

type AblyRuleTargetPulsar struct {
	RoutingKey     types.String         `tfsdk:"routing_key"`
	Topic          types.String         `tfsdk:"topic"`
	ServiceURL     types.String         `tfsdk:"service_url"`
	TlsTrustCerts  []types.String       `tfsdk:"tls_trust_certs"`
	Authentication PulsarAuthentication `tfsdk:"authentication"`
	Enveloped      types.Bool           `tfsdk:"enveloped"`
	Format         types.String         `tfsdk:"format"`
}

type PulsarAuthentication struct {
	Mode  types.String `tfsdk:"mode"`
	Token types.String `tfsdk:"token"`
}

type AblyRuleTargetZapier struct {
	Url          types.String      `tfsdk:"url"`
	Headers      []AblyRuleHeaders `tfsdk:"headers"`
	SigningKeyId types.String      `tfsdk:"signing_key_id"`
}

type AblyRuleTargetIFTTT struct {
	WebhookKey types.String `tfsdk:"webhook_key"`
	EventName  types.String `tfsdk:"event_name"`
}

type AblyRuleTargetAzureFunction struct {
	AzureAppID        types.String      `tfsdk:"azure_app_id"`
	AzureFunctionName types.String      `tfsdk:"function_name"`
	Headers           []AblyRuleHeaders `tfsdk:"headers"`
	SigningKeyID      types.String      `tfsdk:"signing_key_id"`
	Format            types.String      `tfsdk:"format"`
}

type AblyRuleHeaders struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type AblyRuleTargetKafka struct {
	RoutingKey          types.String        `tfsdk:"routing_key"`
	Brokers             []types.String      `tfsdk:"brokers"`
	KafkaAuthentication KafkaAuthentication `tfsdk:"auth"`
	Enveloped           types.Bool          `tfsdk:"enveloped"`
	Format              types.String        `tfsdk:"format"`
}

type AblyRuleTargetAMQP struct {
	QueueID   types.String      `tfsdk:"queue_id"`
	Headers   []AblyRuleHeaders `tfsdk:"headers"`
	Enveloped types.Bool        `tfsdk:"enveloped"`
	Format    types.String      `tfsdk:"format"`
}

type AblyRuleTargetAMQPExternal struct {
	Url                types.String      `tfsdk:"url"`
	RoutingKey         types.String      `tfsdk:"routing_key"`
	Exchange           types.String      `tfsdk:"exchange"`
	MandatoryRoute     types.Bool        `tfsdk:"mandatory_route"`
	PersistentMessages types.Bool        `tfsdk:"persistent_messages"`
	MessageTtl         types.Int64       `tfsdk:"message_ttl"`
	Headers            []AblyRuleHeaders `tfsdk:"headers"`
	Enveloped          types.Bool        `tfsdk:"enveloped"`
	Format             types.String      `tfsdk:"format"`
}

type KafkaAuthentication struct {
	Sasl Sasl `tfsdk:"sasl"`
}

type Sasl struct {
	Mechanism types.String `tfsdk:"mechanism"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
}
