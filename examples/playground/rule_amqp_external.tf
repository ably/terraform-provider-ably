resource "ably_rule_amqp_external" "rule0" {
  app_id       = ably_app.app0.id
  status       = "enabled"
  request_mode = "single"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  target = {
    url                 = "amqps://test.com"
    routing_key         = "new:key"
    exchange            = "testexchange"
    mandatory_route     = true
    persistent_messages = true
    message_ttl         = 55
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
    enveloped      = false
    format         = "json"
  }
}
