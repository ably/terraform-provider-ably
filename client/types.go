package ably

import (
	"encoding/json"
	"fmt"
)

// Error represents an API error response.
type Error struct {
	Message    string          `json:"message"`
	Code       int             `json:"code"`
	StatusCode int             `json:"statusCode"`
	Href       string          `json:"href"`
	Details    json.RawMessage `json:"details,omitempty"`
}

func (e *Error) Error() string {
	if len(e.Details) > 0 && string(e.Details) != "null" {
		return fmt.Sprintf("%s (code: %d, details: %s)", e.Message, e.Code, string(e.Details))
	}
	if e.Code != 0 {
		return fmt.Sprintf("%s (code: %d)", e.Message, e.Code)
	}
	return e.Message
}

// AppPost is the request body for creating an app.
type AppPost struct {
	Name                   string  `json:"name"`
	Status                 string  `json:"status,omitempty"`
	TLSOnly                *bool   `json:"tlsOnly,omitempty"`
	FCMKey                 *string `json:"fcmKey,omitempty"`
	FCMServiceAccount      *string `json:"fcmServiceAccount,omitempty"`
	FCMProjectID           *string `json:"fcmProjectId,omitempty"`
	APNSCertificate        *string `json:"apnsCertificate,omitempty"`
	APNSPrivateKey         *string `json:"apnsPrivateKey,omitempty"`
	APNSUseSandboxEndpoint *bool   `json:"apnsUseSandboxEndpoint,omitempty"`
	APNSAuthType           *string `json:"apnsAuthType,omitempty"`
	APNSSigningKey         *string `json:"apnsSigningKey,omitempty"`
	APNSSigningKeyID       *string `json:"apnsSigningKeyId,omitempty"`
	APNSIssuerKey          *string `json:"apnsIssuerKey,omitempty"`
	APNSTopicHeader        *string `json:"apnsTopicHeader,omitempty"`
}

// AppPatch is the request body for updating an app.
type AppPatch struct {
	Name                   string  `json:"name,omitempty"`
	Status                 string  `json:"status,omitempty"`
	TLSOnly                *bool   `json:"tlsOnly,omitempty"`
	FCMKey                 *string `json:"fcmKey,omitempty"`
	FCMServiceAccount      *string `json:"fcmServiceAccount,omitempty"`
	FCMProjectID           *string `json:"fcmProjectId,omitempty"`
	APNSCertificate        *string `json:"apnsCertificate,omitempty"`
	APNSPrivateKey         *string `json:"apnsPrivateKey,omitempty"`
	APNSUseSandboxEndpoint *bool   `json:"apnsUseSandboxEndpoint,omitempty"`
	APNSAuthType           *string `json:"apnsAuthType,omitempty"`
	APNSSigningKey         *string `json:"apnsSigningKey,omitempty"`
	APNSSigningKeyID       *string `json:"apnsSigningKeyId,omitempty"`
	APNSIssuerKey          *string `json:"apnsIssuerKey,omitempty"`
	APNSTopicHeader        *string `json:"apnsTopicHeader,omitempty"`
}

// AppResponse is the response for app operations.
type AppResponse struct {
	AccountID                   string          `json:"accountId,omitempty"`
	ID                          string          `json:"id,omitempty"`
	Name                        string          `json:"name,omitempty"`
	Status                      string          `json:"status,omitempty"`
	TLSOnly                     *bool           `json:"tlsOnly,omitempty"`
	APNSUseSandboxEndpoint      *bool           `json:"apnsUseSandboxEndpoint,omitempty"`
	APNSAuthType                *string         `json:"apnsAuthType,omitempty"`
	APNSCertificateConfigured   *bool           `json:"apnsCertificateConfigured,omitempty"`
	APNSSigningKeyConfigured    *bool           `json:"apnsSigningKeyConfigured,omitempty"`
	APNSIssuerKey               *string         `json:"apnsIssuerKey,omitempty"`
	APNSSigningKeyID            *string         `json:"apnsSigningKeyId,omitempty"`
	APNSTopicHeader             *string         `json:"apnsTopicHeader,omitempty"`
	FCMProjectID                *string         `json:"fcmProjectId,omitempty"`
	FCMServiceAccountConfigured *bool           `json:"fcmServiceAccountConfigured,omitempty"`
	Created                     int64           `json:"created,omitempty"`
	Modified                    int64           `json:"modified,omitempty"`
	Links                       json.RawMessage `json:"_links,omitempty"`
}

// KeyPost is the request body for creating a key.
type KeyPost struct {
	Name            string              `json:"name"`
	RevocableTokens *bool               `json:"revocableTokens,omitempty"`
	Capability      map[string][]string `json:"capability"`
}

// KeyPatch is the request body for updating a key.
type KeyPatch struct {
	Name            string              `json:"name,omitempty"`
	RevocableTokens *bool               `json:"revocableTokens,omitempty"`
	Capability      map[string][]string `json:"capability,omitempty"`
}

// KeyResponse is the response for key operations.
type KeyResponse struct {
	AppID           string              `json:"appId,omitempty"`
	ID              string              `json:"id,omitempty"`
	Name            string              `json:"name,omitempty"`
	Status          int                 `json:"status,omitempty"`
	Key             string              `json:"key,omitempty"`
	RevocableTokens *bool               `json:"revocableTokens,omitempty"`
	Capability      map[string][]string `json:"capability,omitempty"`
	Created         int64               `json:"created,omitempty"`
	Modified        int64               `json:"modified,omitempty"`
}

// NamespacePost is the request body for creating a namespace.
type NamespacePost struct {
	ID                      string  `json:"id"`
	Authenticated           bool    `json:"authenticated"`
	Persisted               bool    `json:"persisted"`
	PersistLast             bool    `json:"persistLast"`
	PushEnabled             bool    `json:"pushEnabled"`
	TLSOnly                 bool    `json:"tlsOnly"`
	ExposeTimeserial        bool    `json:"exposeTimeserial"`
	MutableMessages         bool    `json:"mutableMessages"`
	PopulateChannelRegistry bool    `json:"populateChannelRegistry"`
	BatchingEnabled         *bool   `json:"batchingEnabled,omitempty"`
	BatchingInterval        *int    `json:"batchingInterval,omitempty"`
	ConflationEnabled       *bool   `json:"conflationEnabled,omitempty"`
	ConflationInterval      *int    `json:"conflationInterval,omitempty"`
	ConflationKey           *string `json:"conflationKey,omitempty"`
}

// NamespacePatch is the request body for updating a namespace.
type NamespacePatch struct {
	Authenticated           *bool   `json:"authenticated,omitempty"`
	Persisted               *bool   `json:"persisted,omitempty"`
	PersistLast             *bool   `json:"persistLast,omitempty"`
	PushEnabled             *bool   `json:"pushEnabled,omitempty"`
	TLSOnly                 *bool   `json:"tlsOnly,omitempty"`
	ExposeTimeserial        *bool   `json:"exposeTimeserial,omitempty"`
	MutableMessages         *bool   `json:"mutableMessages,omitempty"`
	PopulateChannelRegistry *bool   `json:"populateChannelRegistry,omitempty"`
	BatchingEnabled         *bool   `json:"batchingEnabled,omitempty"`
	BatchingInterval        *int    `json:"batchingInterval,omitempty"`
	ConflationEnabled       *bool   `json:"conflationEnabled,omitempty"`
	ConflationInterval      *int    `json:"conflationInterval,omitempty"`
	ConflationKey           *string `json:"conflationKey,omitempty"`
}

// NamespaceResponse is the response for namespace operations.
type NamespaceResponse struct {
	AppID                   string  `json:"appId,omitempty"`
	Authenticated           bool    `json:"authenticated"`
	Created                 int64   `json:"created,omitempty"`
	Modified                int64   `json:"modified,omitempty"`
	ID                      string  `json:"id,omitempty"`
	Persisted               bool    `json:"persisted"`
	PersistLast             bool    `json:"persistLast"`
	PushEnabled             bool    `json:"pushEnabled"`
	TLSOnly                 bool    `json:"tlsOnly"`
	ExposeTimeserial        bool    `json:"exposeTimeserial"`
	MutableMessages         bool    `json:"mutableMessages"`
	PopulateChannelRegistry bool    `json:"populateChannelRegistry"`
	BatchingEnabled         *bool   `json:"batchingEnabled,omitempty"`
	BatchingInterval        *int    `json:"batchingInterval,omitempty"`
	ConflationEnabled       *bool   `json:"conflationEnabled,omitempty"`
	ConflationInterval      *int    `json:"conflationInterval,omitempty"`
	ConflationKey           *string `json:"conflationKey,omitempty"`
}

// Queue is the request body for creating a queue.
type Queue struct {
	Name      string `json:"name"`
	TTL       int    `json:"ttl"`
	MaxLength int    `json:"maxLength"`
	Region    string `json:"region"`
}

// QueueMessages holds queue message counts.
type QueueMessages struct {
	Ready          *int `json:"ready,omitempty"`
	Unacknowledged *int `json:"unacknowledged,omitempty"`
	Total          *int `json:"total,omitempty"`
}

// QueueStats holds queue rate statistics.
type QueueStats struct {
	PublishRate         *float64 `json:"publishRate,omitempty"`
	DeliveryRate        *float64 `json:"deliveryRate,omitempty"`
	AcknowledgementRate *float64 `json:"acknowledgementRate,omitempty"`
}

// QueueAMQP holds AMQP connection details.
type QueueAMQP struct {
	URI       string `json:"uri,omitempty"`
	QueueName string `json:"queueName,omitempty"`
}

// QueueStomp holds STOMP connection details.
type QueueStomp struct {
	URI         string `json:"uri,omitempty"`
	Host        string `json:"host,omitempty"`
	Destination string `json:"destination,omitempty"`
}

// QueueResponse is the response for queue operations.
type QueueResponse struct {
	ID           string        `json:"id,omitempty"`
	AppID        string        `json:"appId,omitempty"`
	Name         string        `json:"name,omitempty"`
	Region       string        `json:"region,omitempty"`
	AMQP         QueueAMQP     `json:"amqp,omitempty"`
	Stomp        QueueStomp    `json:"stomp,omitempty"`
	State        string        `json:"state,omitempty"`
	Messages     QueueMessages `json:"messages,omitempty"`
	Stats        QueueStats    `json:"stats,omitempty"`
	TTL          int           `json:"ttl,omitempty"`
	MaxLength    int           `json:"maxLength,omitempty"`
	Deadletter   bool          `json:"deadletter,omitempty"`
	DeadletterID *string       `json:"deadletterId,omitempty"`
}

// RuleSource is the source configuration for a rule.
type RuleSource struct {
	ChannelFilter string `json:"channelFilter"`
	Type          string `json:"type"`
}

// HTTPRuleTarget is the target for HTTP rules.
type HTTPRuleTarget struct {
	URL          string       `json:"url"`
	Headers      []RuleHeader `json:"headers,omitempty"`
	SigningKeyID *string      `json:"signingKeyId,omitempty"`
	Enveloped    *bool        `json:"enveloped,omitempty"`
	Format       string       `json:"format"`
}

// RuleHeader is a header for rule targets.
type RuleHeader struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// HTTPRulePost is the request body for creating an HTTP rule.
type HTTPRulePost struct {
	Status      string         `json:"status,omitempty"`
	RuleType    string         `json:"ruleType"`
	RequestMode string         `json:"requestMode"`
	Source      RuleSource     `json:"source"`
	Target      HTTPRuleTarget `json:"target"`
}

// HTTPRuleTargetPatch is the patch-specific target for HTTP rules.
// Fields use pointer types so that omitted fields are not sent in the PATCH request.
type HTTPRuleTargetPatch struct {
	URL          *string      `json:"url,omitempty"`
	Headers      []RuleHeader `json:"headers,omitempty"`
	SigningKeyID *string      `json:"signingKeyId,omitempty"`
	Enveloped    *bool        `json:"enveloped,omitempty"`
	Format       *string      `json:"format,omitempty"`
}

// HTTPRulePatch is the request body for updating an HTTP rule.
type HTTPRulePatch struct {
	Status      string               `json:"status,omitempty"`
	RuleType    string               `json:"ruleType"`
	RequestMode string               `json:"requestMode,omitempty"`
	Source      *RuleSourcePatch     `json:"source,omitempty"`
	Target      *HTTPRuleTargetPatch `json:"target,omitempty"`
}

// AWSAuthMode is the authentication mode for AWS rules.
type AWSAuthMode string

// AWS authentication modes.
const (
	AWSAuthModeCredentials AWSAuthMode = "credentials"
	AWSAuthModeAssumeRole  AWSAuthMode = "assumeRole"
)

// AWSLambdaTarget is the target for AWS Lambda rules.
type AWSLambdaTarget struct {
	Region         string            `json:"region"`
	FunctionName   string            `json:"functionName"`
	Enveloped      *bool             `json:"enveloped,omitempty"`
	Authentication AWSAuthentication `json:"authentication"`
}

// AWSLambdaRulePost creates an AWS Lambda rule.
type AWSLambdaRulePost struct {
	Status      string          `json:"status,omitempty"`
	RuleType    string          `json:"ruleType"`
	RequestMode string          `json:"requestMode"`
	Source      RuleSource      `json:"source"`
	Target      AWSLambdaTarget `json:"target"`
}

// AWSKinesisTarget is the target for AWS Kinesis rules.
type AWSKinesisTarget struct {
	Region         string            `json:"region"`
	StreamName     string            `json:"streamName"`
	PartitionKey   string            `json:"partitionKey,omitempty"`
	Enveloped      *bool             `json:"enveloped,omitempty"`
	Format         string            `json:"format"`
	Authentication AWSAuthentication `json:"authentication"`
}

// AWSKinesisRulePost creates an AWS Kinesis rule.
type AWSKinesisRulePost struct {
	Status      string           `json:"status,omitempty"`
	RuleType    string           `json:"ruleType"`
	RequestMode string           `json:"requestMode"`
	Source      RuleSource       `json:"source"`
	Target      AWSKinesisTarget `json:"target"`
}

// AWSSQSTarget is the target for AWS SQS rules.
type AWSSQSTarget struct {
	Region         string            `json:"region"`
	AWSAccountID   string            `json:"awsAccountId"`
	QueueName      string            `json:"queueName"`
	Enveloped      *bool             `json:"enveloped,omitempty"`
	Format         string            `json:"format"`
	Authentication AWSAuthentication `json:"authentication"`
}

// AWSSQSRulePost creates an AWS SQS rule.
type AWSSQSRulePost struct {
	Status      string       `json:"status,omitempty"`
	RuleType    string       `json:"ruleType"`
	RequestMode string       `json:"requestMode"`
	Source      RuleSource   `json:"source"`
	Target      AWSSQSTarget `json:"target"`
}

// BeforePublishConfig holds configuration for before-publish rules.
type BeforePublishConfig struct {
	RetryTimeout          int    `json:"retryTimeout"`
	MaxRetries            int    `json:"maxRetries"`
	FailedAction          string `json:"failedAction"`
	TooManyRequestsAction string `json:"tooManyRequestsAction"`
}

// AMQPRuleTarget is the target for AMQP rules.
type AMQPRuleTarget struct {
	QueueID   string       `json:"queueId"`
	Headers   []RuleHeader `json:"headers,omitempty"`
	Enveloped *bool        `json:"enveloped,omitempty"`
	Format    string       `json:"format"`
}

// AMQPRulePost is the request body for creating an AMQP rule.
type AMQPRulePost struct {
	Status      string         `json:"status,omitempty"`
	RuleType    string         `json:"ruleType"`
	RequestMode string         `json:"requestMode"`
	Source      RuleSource     `json:"source"`
	Target      AMQPRuleTarget `json:"target"`
}

// AMQPExternalRuleTarget is the target for external AMQP rules.
type AMQPExternalRuleTarget struct {
	URL                string       `json:"url"`
	RoutingKey         string       `json:"routingKey"`
	Exchange           string       `json:"exchange,omitempty"`
	MandatoryRoute     *bool        `json:"mandatoryRoute,omitempty"`
	PersistentMessages *bool        `json:"persistentMessages,omitempty"`
	MessageTTL         *int         `json:"messageTtl,omitempty"`
	Headers            []RuleHeader `json:"headers,omitempty"`
	Enveloped          *bool        `json:"enveloped,omitempty"`
	Format             string       `json:"format"`
}

// AMQPExternalRulePost is the request body for creating an external AMQP rule.
type AMQPExternalRulePost struct {
	Status      string                 `json:"status,omitempty"`
	RuleType    string                 `json:"ruleType"`
	RequestMode string                 `json:"requestMode"`
	Source      RuleSource             `json:"source"`
	Target      AMQPExternalRuleTarget `json:"target"`
}

// KafkaRuleTarget is the target for Kafka rules.
type KafkaRuleTarget struct {
	RoutingKey string     `json:"routingKey"`
	Brokers    []string   `json:"brokers"`
	Enveloped  *bool      `json:"enveloped,omitempty"`
	Format     string     `json:"format"`
	Auth       *KafkaAuth `json:"auth,omitempty"`
}

// KafkaAuth holds Kafka authentication details.
type KafkaAuth struct {
	SASL *KafkaSASL `json:"sasl,omitempty"`
}

// KafkaSASL holds Kafka SASL authentication details.
type KafkaSASL struct {
	Mechanism string `json:"mechanism,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"` //nolint:gosec // G117: not a hardcoded credential
}

// KafkaRulePost is the request body for creating a Kafka rule.
type KafkaRulePost struct {
	Status      string          `json:"status,omitempty"`
	RuleType    string          `json:"ruleType"`
	RequestMode string          `json:"requestMode"`
	Source      RuleSource      `json:"source"`
	Target      KafkaRuleTarget `json:"target"`
}

// PulsarRuleTarget is the target for Pulsar rules.
type PulsarRuleTarget struct {
	RoutingKey     string      `json:"routingKey"`
	Topic          string      `json:"topic"`
	ServiceURL     string      `json:"serviceUrl"`
	TLSTrustCerts  []string    `json:"tlsTrustCerts,omitempty"`
	Authentication *PulsarAuth `json:"authentication,omitempty"`
	Enveloped      *bool       `json:"enveloped,omitempty"`
	Format         string      `json:"format"`
}

// PulsarAuth holds Pulsar authentication details.
// The API uses authenticationMode: "token" (not "jwt").
type PulsarAuth struct {
	AuthenticationMode string `json:"authenticationMode"`
	Token              string `json:"token,omitempty"`
}

// PulsarRulePost is the request body for creating a Pulsar rule.
type PulsarRulePost struct {
	Status      string           `json:"status,omitempty"`
	RuleType    string           `json:"ruleType"`
	RequestMode string           `json:"requestMode"`
	Source      RuleSource       `json:"source"`
	Target      PulsarRuleTarget `json:"target"`
}

// RuleResponse is a generic rule response (used when you don't know the type).
// For type-specific handling, use the typed response variants.
type RuleResponse struct {
	ID          string          `json:"id,omitempty"`
	AppID       string          `json:"appId,omitempty"`
	Version     string          `json:"version,omitempty"`
	Status      string          `json:"status,omitempty"`
	Created     float64         `json:"created,omitempty"`
	Modified    float64         `json:"modified,omitempty"`
	Links       json.RawMessage `json:"_links,omitempty"`
	RuleType    string          `json:"ruleType,omitempty"`
	RequestMode string          `json:"requestMode,omitempty"`
	Source      *RuleSource     `json:"source,omitempty"`
	Target      interface{}     `json:"target,omitempty"`
}

// StatsResponse is the response for account and app stats endpoints.
type StatsResponse struct {
	IntervalID string                 `json:"intervalId"`
	Unit       string                 `json:"unit"`
	Schema     string                 `json:"schema"`
	Entries    map[string]interface{} `json:"entries"`
	InProgress string                 `json:"inProgress,omitempty"`
	AccountID  string                 `json:"accountId,omitempty"`
	AppID      string                 `json:"appId,omitempty"`
}

// Me is the response for GET /me.
type Me struct {
	Token   *MeToken   `json:"token,omitempty"`
	User    *MeUser    `json:"user,omitempty"`
	Account *MeAccount `json:"account,omitempty"`
}

// MeToken holds token information from GET /me.
type MeToken struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
}

// MeUser holds user information from GET /me.
type MeUser struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

// MeAccount holds account information from GET /me.
type MeAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// StatsParams holds query parameters for stats endpoints.
type StatsParams struct {
	Start     *int   `json:"start,omitempty"`
	End       *int   `json:"end,omitempty"`
	Unit      string `json:"unit,omitempty"`
	Direction string `json:"direction,omitempty"`
	Limit     *int   `json:"limit,omitempty"`
}
