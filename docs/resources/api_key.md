---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ably_api_key Resource - terraform-provider-ably"
subcategory: ""
description: |-
  
---

# ably_api_key (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `app_id` (String) The Ably application ID which this key is associated with.
- `capabilities` (Map of List of String) The capabilities that this key has. More information on capabilities can be found in the Ably documentation.
- `name` (String) The name for your API key. This is a friendly name for your reference.

### Read-Only

- `created` (Number) Enforce TLS for all connections. This setting overrides any channel setting.
- `id` (String) The key ID.
- `key` (String) The complete API key including API secret.
- `modified` (Number) Unix timestamp representing the date and time of the last modification of the key.
- `status` (Number) The status of the key. 0 is enabled, 1 is revoked.

