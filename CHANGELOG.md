# Changelog

## [0.6.1](https://github.com/ably/terraform-provider-ably/tree/v0.6.1)

[Full Changelog](https://github.com/ably/terraform-provider-ably/compare/v0.6.0...v0.6.1)

**Merged pull requests:**

- \[INF-3250\] - Update Contributing and Provider documentation [\#176](https://github.com/ably/terraform-provider-ably/pull/176) ([graham-russell](https://github.com/graham-russell))

## [0.6.0](https://github.com/ably/terraform-provider-ably/tree/v0.6.0)

[Full Changelog](https://github.com/ably/terraform-provider-ably/compare/v0.5.0...v0.6.0)

**Merged pull requests:**

- Bump google.golang.org/grpc from 1.53.0 to 1.56.3 [\#175](https://github.com/ably/terraform-provider-ably/pull/175) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump golang.org/x/crypto from 0.0.0-20220817201139-bc19a97f63c8 to 0.17.0 [\#174](https://github.com/ably/terraform-provider-ably/pull/174) ([dependabot[bot]](https://github.com/apps/dependabot))
- Bump golang.org/x/net from 0.5.0 to 0.17.0 [\#173](https://github.com/ably/terraform-provider-ably/pull/173) ([dependabot[bot]](https://github.com/apps/dependabot))
- \[INF-3250\] - Add `revocable_tokens` parameter to `ably_api_key` resource [\#171](https://github.com/ably/terraform-provider-ably/pull/171) ([graham-russell](https://github.com/graham-russell))
- Add `exchange` parameter to AMQP External Rule [\#170](https://github.com/ably/terraform-provider-ably/pull/170) ([graham-russell](https://github.com/graham-russell))
- docs: bump readme version [\#169](https://github.com/ably/terraform-provider-ably/pull/169) ([AndyTWF](https://github.com/AndyTWF))
- Bump google.golang.org/grpc from 1.51.0 to 1.53.0 [\#165](https://github.com/ably/terraform-provider-ably/pull/165) ([dependabot[bot]](https://github.com/apps/dependabot))

## [0.5.0](https://github.com/ably/terraform-provider-ably/tree/v0.5.0)

[Full Changelog](https://github.com/ably/terraform-provider-ably/compare/v0.4.3...v0.5.0)

**Merged pull requests:**

- Provide envelope for HTTP rules [\#167](https://github.com/ably/terraform-provider-ably/pull/167) ([AndyTWF](https://github.com/AndyTWF))

Bugfixes:

- The provider now honours "enveloped" settings for HTTP rules in single publish mode

## [0.4.3](https://github.com/ably/terraform-provider-ably/tree/v0.4.3)

[Full Changelog](https://github.com/ably/terraform-provider-ably/compare/v0.4.2...v0.4.3)

**Merged pull requests:**

- Append 'terraform-provider-ably/VERSION' to the Ably-Agent HTTP header [\#156](https://github.com/ably/terraform-provider-ably/pull/156) ([lmars](https://github.com/lmars))
- add credit to CHANGELOG for external contribution [\#155](https://github.com/ably/terraform-provider-ably/pull/155) ([owenpearson](https://github.com/owenpearson))

## [0.4.2](https://github.com/ably/terraform-provider-ably/tree/v0.4.2)

Bugfixes:
- Fix importing of rules
- Fix channel filter being required

## [0.4.1](https://github.com/ably/terraform-provider-ably/tree/v0.4.1)

Bugfixes:
- Fix rules not updating correctly
- Fix resources being recreated when anything changes in app
- Fix terraform plan saying unknown app.id and app.account_id when they are known
- Fix description for apns_use_sandbox_endpoint
- Fix error when TTL is null in amqp/external
- Fix error when setting multiple capabilities

## [0.4.0](https://github.com/ably/terraform-provider-ably/tree/v0.4.0)

Bugfixes:
- Fixes key not being read from the control API ([tete17](https://github.com/tete17))
- Fixes reads not regestering when a resource had been deleted outside of terraform
- Fixes deletes failing when a resource had been deleted outside of terraform

## [0.3.0](https://github.com/ably/terraform-provider-ably/tree/v0.3.0)

This release adds:
- Ably Zapier integration rule via `ably_rule_zapier` resource
- Ably AWS Lambda integration rule via `ably_rule_lambda` resource
- Ably Google Cloud Function integration rule via `ably_rule_google_function` resource
- Ably IFTTT integration rule via `ably_rule_ifttt` resource
- Ably Azure Functions integration rule via `ably_rule_azure_function` resource
- Ably HTTP integration rule via `ably_rule_http` resource
- Ably Kafka integration rule via `ably_rule_kafka` resource
- Ably Pulsar integration rule via `ably_rule_pulsar` resource
- Ably AMQP and external AMQP integration rules via `ably_rule_amqp` and `ably_rule_amqp_external` resources
- Updated documentation

Bugfixes:
- Fixes issues with certain fields (including API Key) being available only on the first apply
- Fixes issue with some optional fields not really being optional

The release also includes additional code quality improvements.

## [0.2.0](https://github.com/ably/terraform-provider-ably/tree/v0.2.0)

This release adds: 
- Ably SQS integration rule via `ably_rule_sqs` resource
- Ably Kinesis integration rule via `ably_rule_kinesis` resource
- Updated documentation

Bugfixes:
- Fixes issues with certain fields (including API Key) being available only on the first apply
- Fixes issue with some optional fields not really being optional

## [0.1.0-beta](https://github.com/ably/terraform-provider-ably/tree/v0.1.0-beta)

Initial release to Terraform Registry. 

This version includes the following resources: 
- `ably_app`
- `ably_key`
- `ably_namespace`
- `ably-queue`
