---
page_title: "ably_rule_sqs Resource - terraform-provider-ably"
subcategory: ""
description: |-
  The ably_rule_sqs resource allows you to create and manage an Ably integration rule for AWS SQS. Read more at https://ably.com/docs/general/firehose/sqs-rule
---

# ably_rule_sqs (Resource)

The `ably_rule_sqs` resource allows you to create and manage an Ably integration rule for AWS SQS. Read more at https://ably.com/docs/general/firehose/sqs-rule


## Example Usage

```terraform
resource "ably_rule_sqs" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  target = {
    region         = "us-west-1",
    aws_account_id = "123456789012",
    queue_name     = "aaaaaa",
    enveloped      = false,
    format         = "json"
    authentication = {
      mode     = "assumeRole",
      role_arn = "cccc"
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_id` (String) The Ably application ID.
- `source` (Attributes) object (rule_source) (see [below for nested schema](#nestedatt--source))
- `target` (Attributes) object (rule_source) (see [below for nested schema](#nestedatt--target))

### Optional

- `request_mode` (String) This is Single Request mode or Batch Request mode. Single Request mode sends each event separately to the endpoint specified by the rule
- `status` (String) The status of the rule. Rules can be enabled or disabled.

### Read-Only

- `id` (String) The rule ID.

<a id="nestedatt--source"></a>
### Nested Schema for `source`

Required:

- `type` (String)

Optional:

- `channel_filter` (String)


<a id="nestedatt--target"></a>
### Nested Schema for `target`

Required:

- `authentication` (Attributes) object (rule_source) (see [below for nested schema](#nestedatt--target--authentication))

Optional:

- `aws_account_id` (String) Your AWS account ID
- `enveloped` (Boolean) Delivered messages are wrapped in an Ably envelope by default that contains metadata about the message and its payload. The form of the envelope depends on whether it is part of a Webhook/Function or a Queue/Firehose rule. For everything besides Webhooks, you can ensure you only get the raw payload by unchecking "Enveloped" when setting up the rule.
- `format` (String) JSON provides a text-based encoding, whereas MsgPack provides a more efficient binary encoding
- `queue_name` (String) The AWS SQS queue name
- `region` (String) The region is which AWS SQS is hosted

<a id="nestedatt--target--authentication"></a>
### Nested Schema for `target.authentication`

Required:

- `mode` (String) Authentication method. Use 'credentials' or 'assumeRole'

Optional:

- `access_key_id` (String, Sensitive) The AWS key ID for the AWS IAM user
- `role_arn` (String) If you are using the 'ARN of an assumable role' authentication method, this is your Assume Role ARN
- `secret_access_key` (String, Sensitive) The AWS secret key for the AWS IAM user