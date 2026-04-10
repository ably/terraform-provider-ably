# Changelog

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
