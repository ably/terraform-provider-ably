---
page_title: "ably_rule_lambda Resource - terraform-provider-ably"
subcategory: ""
description: |-
  The ably_rule_lambda resource allows you to create and manage an Ably integration rule for AWS Lambda. Read more at https://ably.com/docs/general/webhooks/aws-lambda
---

# ably_rule_lambda (Resource)

The `ably_rule_lambda` resource allows you to create and manage an Ably integration rule for AWS Lambda. Read more at https://ably.com/docs/general/webhooks/aws-lambda


## Example Usage

```terraform
resource "ably_rule_lambda" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  target = {
    region        = "us-west-1",
    function_name = "rule0",
    enveloped     = false,
    format        = "json"
    authentication = {
      mode              = "credentials",
      access_key_id     = "hhhh"
      secret_access_key = "ffff"
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

- `enveloped` (Boolean) Delivered messages are wrapped in an Ably envelope by default that contains metadata about the message and its payload. The form of the envelope depends on whether it is part of a Webhook/Function or a Queue/Firehose rule. For everything besides Webhooks, you can ensure you only get the raw payload by unchecking "Enveloped" when setting up the rule.
- `function_name` (String)
- `region` (String)

<a id="nestedatt--target--authentication"></a>
### Nested Schema for `target.authentication`

Required:

- `mode` (String) Authentication method. Use 'credentials' or 'assumeRole'

Optional:

- `access_key_id` (String, Sensitive) The AWS key ID for the AWS IAM user
- `role_arn` (String) If you are using the 'ARN of an assumable role' authentication method, this is your Assume Role ARN
- `secret_access_key` (String, Sensitive) The AWS secret key for the AWS IAM user