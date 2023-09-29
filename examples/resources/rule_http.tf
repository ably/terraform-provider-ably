resource "ably_rule_http" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  target = {
    url = "https://example.com/webhooks"
    headers = [
      {
        name  = "User-Agent"
        value = "user-agent-string"
      },
      {
        name  = "headerName"
        value = "headerValue"
      }
    ]
    signing_key_id = ably_api_key.api_key_0.id
    enveloped      = true
    format         = "json"
    # Note, "enveloped" can only be set to true for "single" request_mode.
    # "batch" request_mode is automatically enveloped.
    enveloped      = false
  }
}
