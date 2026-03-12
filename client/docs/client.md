# Ably Control API Go Client

A Go client library for the [Ably Control API](https://control.ably.net/v1), enabling programmatic management of Ably resources including apps, keys, namespaces, queues, and integration rules.

## Installation

```bash
go get github.com/ably/terraform-provider-ably/client
```

Requires Go 1.24 or later.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ably/terraform-provider-ably/client"
)

func main() {
	client := ably.NewClient("your-control-api-token")
	ctx := context.Background()

	// Get current token/user/account info
	me, err := client.Me(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Account: %s (%s)\n", me.Account.Name, me.Account.ID)

	// Create an app
	app, err := client.CreateApp(ctx, me.Account.ID, ably.AppPost{
		Name: "my-new-app",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created app: %s\n", app.ID)
}
```

## Authentication

All requests require a Control API token, passed when creating the client:

```go
client := ably.NewClient("your-control-api-token")
```

You can generate tokens in the [Ably dashboard](https://ably.com/accounts) under **Account settings > Control API tokens**. Tokens are scoped to an account and can have specific capabilities (read, write, etc.).

The token is sent as a `Bearer` token in the `Authorization` header of every request.

## Client Configuration

### Default Configuration

```go
// Creates a client with default settings:
// - Base URL: https://control.ably.net/v1
// - Retry: up to 4 retries with exponential backoff on 5xx errors
// - User-Agent: ably-control-api/0.1.0
client := ably.NewClient(token)
```

### Custom Retry Count

```go
// Disable retries
client := ably.NewClient(token, ably.WithRetryMax(0))

// Set max retries to 2
client := ably.NewClient(token, ably.WithRetryMax(2))
```

### Custom User Agent

```go
// Override the default user agent entirely
client := ably.NewClient(token, ably.WithUserAgent("my-tool/1.0"))

// Or append to the default after creation
client := ably.NewClient(token)
client.UserAgent += " my-tool/1.0"
```

### Custom HTTP Client

```go
import "net/http"

httpClient := &http.Client{
	Timeout: 30 * time.Second,
}
client := ably.NewClient(token, ably.WithHTTPClient(httpClient))
```

### Combining Options

```go
client := ably.NewClient(token,
	ably.WithRetryMax(2),
	ably.WithUserAgent("my-terraform-provider/1.0"),
	ably.WithHTTPClient(customHTTPClient),
)
```

## Resource Operations

All methods accept a `context.Context` as the first parameter for cancellation and deadline support.

### Me (Token Info)

Retrieve information about the current token, user, and account.

```go
me, err := client.Me(ctx)
// me.Token  - token ID, name, capabilities
// me.User   - user ID, email
// me.Account - account ID, name
```

### Apps

#### List Apps

```go
apps, err := client.ListApps(ctx, accountID)
for _, app := range apps {
	fmt.Printf("%s: %s (status: %s)\n", app.ID, app.Name, app.Status)
}
```

#### Create App

```go
tlsOnly := true
app, err := client.CreateApp(ctx, accountID, ably.AppPost{
	Name:    "production-app",
	TLSOnly: &tlsOnly,
})
```

#### Update App

```go
app, err := client.UpdateApp(ctx, appID, ably.AppPatch{
	Name: "renamed-app",
})
```

#### Delete App

```go
err := client.DeleteApp(ctx, appID)
```

#### Upload PKCS12 Certificate

```go
p12Data, _ := os.ReadFile("cert.p12")
app, err := client.UpdateAppPKCS12(ctx, appID, p12Data, "certificate-password")
```

#### Get App Stats

```go
start := 1700000000
limit := 100
stats, err := client.GetAppStats(ctx, appID, &ably.StatsParams{
	Start: &start,
	Unit:  "hour",
	Limit: &limit,
})
```

### Keys

#### List Keys

```go
keys, err := client.ListKeys(ctx, appID)
for _, key := range keys {
	fmt.Printf("%s: %s\n", key.ID, key.Name)
}
```

#### Create Key

```go
key, err := client.CreateKey(ctx, appID, ably.KeyPost{
	Name: "my-api-key",
	Capability: map[string][]string{
		"*":       {"publish", "subscribe", "presence"},
		"private": {"publish"},
	},
})
fmt.Println("Key:", key.Key)
```

#### Update Key

```go
key, err := client.UpdateKey(ctx, appID, keyID, ably.KeyPatch{
	Name: "renamed-key",
	Capability: map[string][]string{
		"public": {"subscribe"},
	},
})
```

#### Revoke Key

```go
err := client.RevokeKey(ctx, appID, keyID)
```

### Namespaces

#### List Namespaces

```go
namespaces, err := client.ListNamespaces(ctx, appID)
```

#### Create Namespace

```go
ns, err := client.CreateNamespace(ctx, appID, ably.NamespacePost{
	ID:            "chat",
	Authenticated: true,
	Persisted:     true,
	PersistLast:   true,
	TLSOnly:       true,
})
```

#### Update Namespace

```go
persisted := false
ns, err := client.UpdateNamespace(ctx, appID, "chat", ably.NamespacePatch{
	Persisted: &persisted,
})
```

#### Delete Namespace

```go
err := client.DeleteNamespace(ctx, appID, "chat")
```

### Queues

#### List Queues

```go
queues, err := client.ListQueues(ctx, appID)
for _, q := range queues {
	fmt.Printf("%s: %s (state: %s)\n", q.ID, q.Name, q.State)
}
```

#### Create Queue

```go
queue, err := client.CreateQueue(ctx, appID, ably.Queue{
	Name:      "my-queue",
	TTL:       60,
	MaxLength: 10000,
	Region:    "us-east-1-a",
})
```

#### Delete Queue

```go
err := client.DeleteQueue(ctx, appID, queueID)
```

### Rules

Rules connect Ably channels to external services. The `CreateRule` and `UpdateRule` methods accept `interface{}` so you can pass any rule type.

#### HTTP Rule

```go
rule, err := client.CreateRule(ctx, appID, ably.HTTPRulePost{
	RuleType:    "http",
	RequestMode: "single",
	Source: ably.RuleSource{
		ChannelFilter: "^alerts",
		Type:          "channel.message",
	},
	Target: ably.HTTPRuleTarget{
		URL:    "https://example.com/webhook",
		Format: "json",
		Headers: []ably.RuleHeader{
			{Name: "X-Custom-Header", Value: "my-value"},
		},
	},
})
```

#### AWS Lambda Rule

```go
rule, err := client.CreateRule(ctx, appID, ably.AWSLambdaRulePost{
	RuleType:    "aws/lambda",
	RequestMode: "single",
	Source: ably.RuleSource{
		ChannelFilter: "^events",
		Type:          "channel.message",
	},
	Target: ably.AWSLambdaTarget{
		Region:       "us-east-1",
		FunctionName: "my-lambda-function",
		Authentication: ably.AWSAuthentication{
			AuthenticationMode: string(ably.AWSAuthModeCredentials),
			AccessKeyID:        "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	},
})
```

#### AWS Kinesis Rule

```go
rule, err := client.CreateRule(ctx, appID, ably.AWSKinesisRulePost{
	RuleType: "aws/kinesis",
	Source: ably.RuleSource{
		ChannelFilter: "^stream",
		Type:          "channel.message",
	},
	Target: ably.AWSKinesisTarget{
		Region:       "us-east-1",
		StreamName:   "my-stream",
		PartitionKey: "#{message.name}",
		Format:       "json",
		Authentication: ably.AWSAuthentication{
			AuthenticationMode: string(ably.AWSAuthModeAssumeRole),
			AssumeRoleArn:      "arn:aws:iam::123456789:role/my-role",
		},
	},
})
```

#### AWS SQS Rule

```go
rule, err := client.CreateRule(ctx, appID, ably.AWSSQSRulePost{
	RuleType: "aws/sqs",
	Source: ably.RuleSource{
		ChannelFilter: "^tasks",
		Type:          "channel.message",
	},
	Target: ably.AWSSQSTarget{
		Region:       "us-east-1",
		AWSAccountID: "123456789012",
		QueueName:    "my-queue",
		Format:       "json",
		Authentication: ably.AWSAuthentication{
			AuthenticationMode: string(ably.AWSAuthModeCredentials),
			AccessKeyID:        "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	},
})
```

#### AMQP Rule (Internal)

Routes messages to an Ably-managed queue.

```go
rule, err := client.CreateRule(ctx, appID, ably.AMQPRulePost{
	RuleType:    "amqp",
	RequestMode: "single",
	Source: ably.RuleSource{
		ChannelFilter: "^notifications",
		Type:          "channel.message",
	},
	Target: ably.AMQPRuleTarget{
		QueueID: "my-queue-id",
		Format:  "json",
	},
})
```

#### AMQP Rule (External)

Routes messages to an external AMQP broker.

```go
rule, err := client.CreateRule(ctx, appID, ably.AMQPExternalRulePost{
	RuleType:    "amqp/external",
	RequestMode: "single",
	Source: ably.RuleSource{
		ChannelFilter: "^data",
		Type:          "channel.message",
	},
	Target: ably.AMQPExternalRuleTarget{
		URL:        "amqps://user:pass@broker.example.com:5671",
		RoutingKey: "my-routing-key",
		Exchange:   "my-exchange",
		Format:     "json",
	},
})
```

#### Kafka Rule

```go
rule, err := client.CreateRule(ctx, appID, ably.KafkaRulePost{
	RuleType:    "kafka",
	RequestMode: "single",
	Source: ably.RuleSource{
		ChannelFilter: "^events",
		Type:          "channel.message",
	},
	Target: ably.KafkaRuleTarget{
		RoutingKey: "my-topic",
		Brokers:    []string{"broker1.example.com:9092", "broker2.example.com:9092"},
		Format:     "json",
		Auth: &ably.KafkaAuth{
			SASL: &ably.KafkaSASL{
				Mechanism: "scram-sha-256",
				Username:  "kafka-user",
				Password:  "kafka-pass",
			},
		},
	},
})
```

#### Pulsar Rule

```go
rule, err := client.CreateRule(ctx, appID, ably.PulsarRulePost{
	RuleType:    "pulsar",
	RequestMode: "single",
	Source: ably.RuleSource{
		ChannelFilter: "^telemetry",
		Type:          "channel.message",
	},
	Target: ably.PulsarRuleTarget{
		RoutingKey: "persistent://tenant/ns/topic",
		Topic:      "my-topic",
		ServiceURL: "pulsar+ssl://pulsar.example.com:6651",
		Format:     "json",
		Authentication: &ably.PulsarAuth{
			AuthenticationMode: "token",
			Token:              "pulsar-jwt-token",
		},
	},
})
```

#### List Rules

```go
rules, err := client.ListRules(ctx, appID)
for _, r := range rules {
	fmt.Printf("%s: type=%s status=%s\n", r.ID, r.RuleType, r.Status)
}
```

#### Get a Single Rule

```go
rule, err := client.GetRule(ctx, appID, ruleID)
```

#### Update a Rule

Pass any rule post type to update. The fields you include will be patched.

```go
rule, err := client.UpdateRule(ctx, appID, ruleID, ably.HTTPRulePost{
	RuleType:    "http",
	RequestMode: "single",
	Source: ably.RuleSource{
		ChannelFilter: "^alerts:.*",
		Type:          "channel.message",
	},
	Target: ably.HTTPRuleTarget{
		URL:    "https://example.com/webhook-v2",
		Format: "json",
	},
})
```

#### Delete a Rule

```go
err := client.DeleteRule(ctx, appID, ruleID)
```

### Statistics

#### App Stats

```go
start := 1700000000
stats, err := client.GetAppStats(ctx, appID, &ably.StatsParams{
	Start:     &start,
	Unit:      "hour",
	Direction: "forwards",
})
for _, s := range stats {
	fmt.Printf("Interval: %s, Unit: %s\n", s.IntervalID, s.Unit)
}
```

#### Account Stats

```go
stats, err := client.GetAccountStats(ctx, accountID, &ably.StatsParams{
	Unit: "month",
})
```

#### Stats Parameters

The `StatsParams` struct supports the following fields:

| Field       | Type     | Description                                    |
|-------------|----------|------------------------------------------------|
| `Start`     | `*int`   | Start of the query interval (Unix ms)          |
| `End`       | `*int`   | End of the query interval (Unix ms)            |
| `Unit`      | `string` | Interval granularity: `minute`, `hour`, `day`, `month` |
| `Direction` | `string` | `forwards` or `backwards` (default)            |
| `Limit`     | `*int`   | Maximum number of results to return            |

Pass `nil` for `StatsParams` to use all defaults.

## Error Handling

All methods return errors. API errors are returned as `*ably.Error`, which implements the `error` interface.

### Checking for API Errors

```go
import "errors"

app, err := client.CreateApp(ctx, accountID, body)
if err != nil {
	var apiErr *ably.Error
	if errors.As(err, &apiErr) {
		fmt.Printf("API error: %s (code: %d, status: %d)\n",
			apiErr.Message, apiErr.Code, apiErr.StatusCode)
		if apiErr.Href != "" {
			fmt.Printf("See: %s\n", apiErr.Href)
		}
		if len(apiErr.Details) > 0 {
			fmt.Printf("Details: %s\n", string(apiErr.Details))
		}
	} else {
		// Network error, timeout, etc.
		fmt.Printf("Request failed: %v\n", err)
	}
}
```

### Error Fields

| Field        | Type               | Description                                  |
|--------------|--------------------|----------------------------------------------|
| `Message`    | `string`           | Human-readable error description             |
| `Code`       | `int`              | Ably-specific error code                     |
| `StatusCode` | `int`              | HTTP status code (e.g., 400, 401, 404, 422)  |
| `Href`       | `string`           | Link to documentation about this error       |
| `Details`    | `json.RawMessage`  | Additional error details (JSON)              |

### Common Status Codes

| Status | Meaning                                              |
|--------|------------------------------------------------------|
| 400    | Bad request - invalid parameters                     |
| 401    | Unauthorized - invalid or expired token              |
| 403    | Forbidden - token lacks required capability          |
| 404    | Not found - resource does not exist                  |
| 422    | Unprocessable entity - validation failed             |
| 429    | Rate limited - too many requests                     |
| 500+   | Server error - automatically retried (up to RetryMax)|

## API Reference Summary

| Method | Signature | Description |
|--------|-----------|-------------|
| `NewClient` | `NewClient(token string, opts ...ClientOption) *Client` | Create a new client |
| `WithRetryMax` | `WithRetryMax(n int) ClientOption` | Set max retries (default 4) |
| `WithUserAgent` | `WithUserAgent(ua string) ClientOption` | Override user agent |
| `WithHTTPClient` | `WithHTTPClient(hc *http.Client) ClientOption` | Set custom HTTP client |
| `Me` | `Me(ctx) (Me, error)` | Get token/user/account info |
| `ListApps` | `ListApps(ctx, accountID) ([]AppResponse, error)` | List all apps |
| `CreateApp` | `CreateApp(ctx, accountID, AppPost) (AppResponse, error)` | Create an app |
| `UpdateApp` | `UpdateApp(ctx, appID, AppPatch) (AppResponse, error)` | Update an app |
| `DeleteApp` | `DeleteApp(ctx, appID) error` | Delete an app |
| `UpdateAppPKCS12` | `UpdateAppPKCS12(ctx, appID, []byte, string) (AppResponse, error)` | Upload PKCS12 cert |
| `GetAppStats` | `GetAppStats(ctx, appID, *StatsParams) ([]AppStatsResponse, error)` | Get app statistics |
| `ListKeys` | `ListKeys(ctx, appID) ([]KeyResponse, error)` | List all keys |
| `CreateKey` | `CreateKey(ctx, appID, KeyPost) (KeyResponse, error)` | Create a key |
| `UpdateKey` | `UpdateKey(ctx, appID, keyID, KeyPatch) (KeyResponse, error)` | Update a key |
| `RevokeKey` | `RevokeKey(ctx, appID, keyID) error` | Revoke a key |
| `ListNamespaces` | `ListNamespaces(ctx, appID) ([]NamespaceResponse, error)` | List namespaces |
| `CreateNamespace` | `CreateNamespace(ctx, appID, NamespacePost) (NamespaceResponse, error)` | Create a namespace |
| `UpdateNamespace` | `UpdateNamespace(ctx, appID, nsID, NamespacePatch) (NamespaceResponse, error)` | Update a namespace |
| `DeleteNamespace` | `DeleteNamespace(ctx, appID, nsID) error` | Delete a namespace |
| `ListQueues` | `ListQueues(ctx, appID) ([]QueueResponse, error)` | List queues |
| `CreateQueue` | `CreateQueue(ctx, appID, Queue) (QueueResponse, error)` | Create a queue |
| `DeleteQueue` | `DeleteQueue(ctx, appID, queueID) error` | Delete a queue |
| `ListRules` | `ListRules(ctx, appID) ([]RuleResponse, error)` | List rules |
| `CreateRule` | `CreateRule(ctx, appID, interface{}) (RuleResponse, error)` | Create a rule |
| `GetRule` | `GetRule(ctx, appID, ruleID) (RuleResponse, error)` | Get a rule |
| `UpdateRule` | `UpdateRule(ctx, appID, ruleID, interface{}) (RuleResponse, error)` | Update a rule |
| `DeleteRule` | `DeleteRule(ctx, appID, ruleID) error` | Delete a rule |
| `GetAccountStats` | `GetAccountStats(ctx, accountID, *StatsParams) ([]AccountStatsResponse, error)` | Get account stats |
