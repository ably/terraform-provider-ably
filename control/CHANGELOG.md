# Changelog

## [v0.2.0](https://github.com/ably/terraform-provider-ably/tree/control/v0.2.0)

**Merged pull requests:**

- build(deps): bump software.sslmate.com/src/go-pkcs12 from 0.7.0 to 0.7.2 in /control [#250](https://github.com/ably/terraform-provider-ably/pull/250)
- \[INF-8060\] Generate remaining resources and add data sources [#249](https://github.com/ably/terraform-provider-ably/pull/249)
- Reduce concurrency in CI [#246](https://github.com/ably/terraform-provider-ably/pull/246)
- \[INF-7769\] Make retries more conservative [#244](https://github.com/ably/terraform-provider-ably/pull/244)
- Bump golang.org/x/crypto from 0.45.0 to 0.52.0 in /control [#242](https://github.com/ably/terraform-provider-ably/pull/242)
- \[INF-7589\] Switch from `authenticated` field to `identified` [#238](https://github.com/ably/terraform-provider-ably/pull/238)
- Control API code generation: strategy, hermetic tests, and first resources [#237](https://github.com/ably/terraform-provider-ably/pull/237)
- Bump golang.org/x/crypto from 0.11.0 to 0.45.0 in /control [#230](https://github.com/ably/terraform-provider-ably/pull/230)

**Breaking changes:**

- `NamespacePost.Authenticated` and `NamespacePost.Identified` are both `*bool`
  with `omitempty`, where `Authenticated` was previously `bool`. `Identified` is
  the canonical field; `Authenticated` is a legacy alias for the same setting.
  The server treats both being sent as a conflicting pair, so leave the unused
  one nil. Set `Identified`.
- `BeforePublishAWSLambdaRulePost.Source` and
  `BeforePublishAWSLambdaRulePatch.Source` are now `*ChatMessageRuleSource`
  rather than `*RuleSource`. Before-publish rules validate against a source
  schema that carries only a type, accepts `"chat.message"` and rejects a
  `channelFilter`. The API assigns the source when a rule is created without
  one, so leave it nil.

**Other changes:**

- Retry defaults are more conservative, giving a degraded API server room to
  recover. `DefaultRetryMax` drops from 4 to 2, and `DefaultRetryWaitMin` (2s)
  and `DefaultRetryWaitMax` (60s) are now exported alongside it.
  `WithRetryWaitMin` and `WithRetryWaitMax` join `WithRetryMax` as functional
  options.
- `NamespacePatch` and `NamespaceResponse` carry `Identified`, so the flag can
  be read and updated under its canonical name.
- `RuleResponse` decodes `InvocationMode`, `ChatRoomFilter` and
  `BeforePublishConfig`. The Control API returns these for moderation and
  before-publish rule types; they were previously discarded.
- `MeToken` carries `ExpiresAt` and `LastUsedAt`, both nil when the token never
  expires or has not been used since Ably began tracking it.
- Go 1.26, and `software.sslmate.com/src/go-pkcs12` and `golang.org/x/crypto`
  bumped.

## [v0.1.0](https://github.com/ably/terraform-provider-ably/tree/control/v0.1.0)

Initial release as an in-repo module under `control/`, replacing the
standalone [ably-control-go](https://github.com/ably/ably-control-go)
repository.

**Merged pull requests:**

- \[INF-6939\] Migrate client into this repo and switch provider to it [#229](https://github.com/ably/terraform-provider-ably/pull/229)

This release includes:
- Full rewrite of the client to match the current Ably Control API
  surface, fixing longstanding misalignment issues in the previous client
- Apps, keys, namespaces, queues, stats, and account info (me)
- Reactor/integration rules: HTTP, AMQP, AMQP external, Kafka, Kinesis,
  Lambda, SQS, Pulsar, IFTTT, Zapier, Cloudflare Workers, Azure
  Functions, Google Cloud Functions
- Ingress rules: Postgres Outbox, MongoDB
- Before-publish rules: webhook, AWS Lambda
- Moderation rules: Hive text-model-only, Hive dashboard, Bodyguard,
  Tisane, Azure Text Moderation
- Configurable retry (exponential backoff on 5xx, no retry on 4xx),
  custom User-Agent, and pluggable HTTP transport via functional options
- Comprehensive unit tests for all endpoints, plus integration test
  scaffolding gated behind `ABLY_ACCOUNT_TOKEN`
