# Change log

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
- Fixes key not being read from the control API
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


