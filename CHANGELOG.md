# Change log

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


