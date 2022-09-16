resource "ably_rule_zapier" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  request_mode = "batch"
  target = {
    url = "https://example.com/webhooks",
    headers = [
      {
        name : "User-Agent",
        value : "user-agent-string",
      },
      {
        name : "User-Agent-Extra",
        value : "user-agent-string",
      },
    ]
    signing_key_id = ably_api_key.api_key_1.id
  }
}
