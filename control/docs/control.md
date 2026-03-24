# Ably Control API Go Client

A Go client for the [Ably Control API](https://control.ably.net/v1),
providing programmatic management of Ably apps, keys, namespaces,
queues, and integration rules.

## Installation

```bash
go get github.com/ably/terraform-provider-ably/control
```

Requires Go 1.24 or later.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ably/terraform-provider-ably/control"
)

func main() {
	client := control.NewClient("your-control-api-token")
	ctx := context.Background()

	me, err := client.Me(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Account: %s (%s)\n", me.Account.Name, me.Account.ID)

	app, err := client.CreateApp(ctx, me.Account.ID, control.AppPost{
		Name: "my-new-app",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created app: %s\n", app.ID)
}
```

## Authentication

All requests require a Control API token, passed when creating the
client:

```go
client := control.NewClient("your-control-api-token")
```

Tokens can be generated in the
[Ably dashboard](https://ably.com/accounts) under **Account settings >
Control API tokens**. They are scoped to an account and can have
specific capabilities (read, write, etc.). The token is sent as a
`Bearer` token in the `Authorization` header.

## Client Configuration

### Defaults

```go
// Base URL:    https://control.ably.net/v1
// Retry:       up to 4 retries with exponential backoff on 5xx errors
// User-Agent:  ably-control-api/0.1.0
client := control.NewClient(token)
```

### Custom Retry Count

```go
client := control.NewClient(token, control.WithRetryMax(0)) // disable retries
client := control.NewClient(token, control.WithRetryMax(2)) // max 2 retries
```

### Custom User Agent

```go
// Override entirely
client := control.NewClient(token, control.WithUserAgent("my-tool/1.0"))

// Or append after creation
client := control.NewClient(token)
client.UserAgent += " my-tool/1.0"
```

### Custom HTTP Client

```go
httpClient := &http.Client{Timeout: 30 * time.Second}
client := control.NewClient(token, control.WithHTTPClient(httpClient))
```

### Combining Options

```go
client := control.NewClient(token,
	control.WithRetryMax(2),
	control.WithUserAgent("my-terraform-provider/1.0"),
	control.WithHTTPClient(customHTTPClient),
)
```

## Resource Operations

All methods take a `context.Context` as their first parameter.

### Me (Token Info)

```go
me, err := client.Me(ctx)
// me.Token   - token ID, name, capabilities
// me.User    - user ID, email
// me.Account - account ID, name
```

### Apps

```go
// List
apps, err := client.ListApps(ctx, accountID)

// Create
tlsOnly := true
app, err := client.CreateApp(ctx, accountID, control.AppPost{
	Name:    "production-app",
	TLSOnly: &tlsOnly,
})

// Update
app, err := client.UpdateApp(ctx, appID, control.AppPatch{
	Name: "renamed-app",
})

// Delete
err := client.DeleteApp(ctx, appID)

// Upload PKCS12 certificate
p12Data, _ := os.ReadFile("cert.p12")
app, err := client.UpdateAppPKCS12(ctx, appID, p12Data, "certificate-password")
```

### Keys

```go
// List
keys, err := client.ListKeys(ctx, appID)

// Create
key, err := client.CreateKey(ctx, appID, control.KeyPost{
	Name: "my-api-key",
	Capability: map[string][]string{
		"*":       {"publish", "subscribe", "presence"},
		"private": {"publish"},
	},
})

// Update
key, err := client.UpdateKey(ctx, appID, keyID, control.KeyPatch{
	Name: "renamed-key",
	Capability: map[string][]string{
		"public": {"subscribe"},
	},
})

// Revoke
err := client.RevokeKey(ctx, appID, keyID)
```

### Namespaces

```go
// List
namespaces, err := client.ListNamespaces(ctx, appID)

// Create
ns, err := client.CreateNamespace(ctx, appID, control.NamespacePost{
	ID:            "chat",
	Authenticated: true,
	Persisted:     true,
	PersistLast:   true,
	TLSOnly:       true,
})

// Update
persisted := false
ns, err := client.UpdateNamespace(ctx, appID, "chat", control.NamespacePatch{
	Persisted: &persisted,
})

// Delete
err := client.DeleteNamespace(ctx, appID, "chat")
```

### Queues

```go
// List
queues, err := client.ListQueues(ctx, appID)

// Create
queue, err := client.CreateQueue(ctx, appID, control.Queue{
	Name:      "my-queue",
	TTL:       60,
	MaxLength: 10000,
	Region:    "us-east-1-a",
})

// Delete
err := client.DeleteQueue(ctx, appID, queueID)
```

### Rules

`CreateRule` and `UpdateRule` accept `any` as the body, so you can pass
whichever rule-type struct matches your integration. A few common
examples follow; see `types.go` for the full set of rule types.

#### HTTP

```go
rule, err := client.CreateRule(ctx, appID, control.HTTPRulePost{
	RuleType:    "http",
	RequestMode: "single",
	Source: control.RuleSource{
		ChannelFilter: "^alerts",
		Type:          "channel.message",
	},
	Target: control.HTTPRuleTarget{
		URL:    "https://example.com/webhook",
		Format: "json",
		Headers: []control.RuleHeader{
			{Name: "X-Custom-Header", Value: "my-value"},
		},
	},
})
```

#### AWS Lambda

```go
rule, err := client.CreateRule(ctx, appID, control.AWSLambdaRulePost{
	RuleType:    "aws/lambda",
	RequestMode: "single",
	Source: control.RuleSource{
		ChannelFilter: "^events",
		Type:          "channel.message",
	},
	Target: control.AWSLambdaTarget{
		Region:       "us-east-1",
		FunctionName: "my-lambda-function",
		Authentication: control.AWSAuthentication{
			AuthenticationMode: string(control.AWSAuthModeCredentials),
			AccessKeyID:        "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	},
})
```

#### AWS Kinesis

```go
rule, err := client.CreateRule(ctx, appID, control.AWSKinesisRulePost{
	RuleType: "aws/kinesis",
	Source: control.RuleSource{
		ChannelFilter: "^stream",
		Type:          "channel.message",
	},
	Target: control.AWSKinesisTarget{
		Region:       "us-east-1",
		StreamName:   "my-stream",
		PartitionKey: "#{message.name}",
		Format:       "json",
		Authentication: control.AWSAuthentication{
			AuthenticationMode: string(control.AWSAuthModeAssumeRole),
			AssumeRoleArn:      "arn:aws:iam::123456789:role/my-role",
		},
	},
})
```

#### AWS SQS

```go
rule, err := client.CreateRule(ctx, appID, control.AWSSQSRulePost{
	RuleType: "aws/sqs",
	Source: control.RuleSource{
		ChannelFilter: "^tasks",
		Type:          "channel.message",
	},
	Target: control.AWSSQSTarget{
		Region:       "us-east-1",
		AWSAccountID: "123456789012",
		QueueName:    "my-queue",
		Format:       "json",
		Authentication: control.AWSAuthentication{
			AuthenticationMode: string(control.AWSAuthModeCredentials),
			AccessKeyID:        "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	},
})
```

#### AMQP (internal Ably queue)

```go
rule, err := client.CreateRule(ctx, appID, control.AMQPRulePost{
	RuleType:    "amqp",
	RequestMode: "single",
	Source: control.RuleSource{
		ChannelFilter: "^notifications",
		Type:          "channel.message",
	},
	Target: control.AMQPRuleTarget{
		QueueID: "my-queue-id",
		Format:  "json",
	},
})
```

#### AMQP (external broker)

```go
rule, err := client.CreateRule(ctx, appID, control.AMQPExternalRulePost{
	RuleType:    "amqp/external",
	RequestMode: "single",
	Source: control.RuleSource{
		ChannelFilter: "^data",
		Type:          "channel.message",
	},
	Target: control.AMQPExternalRuleTarget{
		URL:        "amqps://user:pass@broker.example.com:5671",
		RoutingKey: "my-routing-key",
		Exchange:   "my-exchange",
		Format:     "json",
	},
})
```

#### Kafka

```go
rule, err := client.CreateRule(ctx, appID, control.KafkaRulePost{
	RuleType:    "kafka",
	RequestMode: "single",
	Source: control.RuleSource{
		ChannelFilter: "^events",
		Type:          "channel.message",
	},
	Target: control.KafkaRuleTarget{
		RoutingKey: "my-topic",
		Brokers:    []string{"broker1.example.com:9092", "broker2.example.com:9092"},
		Format:     "json",
		Auth: &control.KafkaAuth{
			SASL: &control.KafkaSASL{
				Mechanism: "scram-sha-256",
				Username:  "kafka-user",
				Password:  "kafka-pass",
			},
		},
	},
})
```

#### Pulsar

```go
rule, err := client.CreateRule(ctx, appID, control.PulsarRulePost{
	RuleType:    "pulsar",
	RequestMode: "single",
	Source: control.RuleSource{
		ChannelFilter: "^telemetry",
		Type:          "channel.message",
	},
	Target: control.PulsarRuleTarget{
		RoutingKey: "persistent://tenant/ns/topic",
		Topic:      "my-topic",
		ServiceURL: "pulsar+ssl://pulsar.example.com:6651",
		Format:     "json",
		Authentication: &control.PulsarAuth{
			AuthenticationMode: "token",
			Token:              "pulsar-jwt-token",
		},
	},
})
```

#### Other rule types

The library also supports these rule types (see `types.go` for structs
and fields):

| Category | Types |
|----------|-------|
| HTTP variants | IFTTT, Zapier, Cloudflare Worker, Azure Function, Google Cloud Function |
| Before-publish | Webhook, AWS Lambda |
| Moderation | Hive (text-model-only, dashboard), Bodyguard, Tisane, Azure Text Moderation |
| Ingress | Postgres Outbox, MongoDB |

#### List, Get, Update, Delete

```go
// List all rules
rules, err := client.ListRules(ctx, appID)

// Get a single rule
rule, err := client.GetRule(ctx, appID, ruleID)

// Update (pass the same post struct type)
rule, err := client.UpdateRule(ctx, appID, ruleID, control.HTTPRulePost{
	RuleType:    "http",
	RequestMode: "single",
	Source: control.RuleSource{
		ChannelFilter: "^alerts:.*",
		Type:          "channel.message",
	},
	Target: control.HTTPRuleTarget{
		URL:    "https://example.com/webhook-v2",
		Format: "json",
	},
})

// Delete
err := client.DeleteRule(ctx, appID, ruleID)
```

### Statistics

```go
// App stats
start := 1700000000
stats, err := client.GetAppStats(ctx, appID, &control.StatsParams{
	Start:     &start,
	Unit:      "hour",
	Direction: "forwards",
})
for _, s := range stats {
	fmt.Printf("Interval: %s, Unit: %s\n", s.IntervalID, s.Unit)
}

// Account stats
stats, err := client.GetAccountStats(ctx, accountID, &control.StatsParams{
	Unit: "month",
})
```

`StatsParams` fields:

| Field       | Type     | Description                                            |
|-------------|----------|--------------------------------------------------------|
| `Start`     | `*int`   | Start of the query interval (Unix ms)                  |
| `End`       | `*int`   | End of the query interval (Unix ms)                    |
| `Unit`      | `string` | Granularity: `minute`, `hour`, `day`, `month`          |
| `Direction` | `string` | `forwards` or `backwards` (default)                    |
| `Limit`     | `*int`   | Maximum number of results                              |

Pass `nil` for defaults.

## Error Handling

API errors are returned as `*control.Error`:

```go
app, err := client.CreateApp(ctx, accountID, body)
if err != nil {
	var apiErr *control.Error
	if errors.As(err, &apiErr) {
		fmt.Printf("API error: %s (code: %d, status: %d)\n",
			apiErr.Message, apiErr.Code, apiErr.StatusCode)
	} else {
		fmt.Printf("Request failed: %v\n", err) // network error, timeout, etc.
	}
}
```

| Field        | Type               | Description                            |
|--------------|--------------------|----------------------------------------|
| `Message`    | `string`           | Human-readable error description       |
| `Code`       | `int`              | Ably error code                        |
| `StatusCode` | `int`              | HTTP status code                       |
| `Href`       | `string`           | Link to error documentation            |
| `Details`    | `json.RawMessage`  | Additional structured error details    |

| Status | Meaning                                               |
|--------|-------------------------------------------------------|
| 400    | Bad request                                           |
| 401    | Unauthorized (invalid or expired token)               |
| 403    | Forbidden (token lacks required capability)           |
| 404    | Not found                                             |
| 422    | Validation failed                                     |
| 429    | Rate limited                                          |
| 500+   | Server error (retried automatically up to `RetryMax`) |

