resource "ably_rule" "amqp_external" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "amqp/external"
  request_mode = "single"
  source       = ably_rule_source.amqp_external
  target       = ably_rule_target_amqp_external.amqp_external
}

resource "ably_rule_source" "amqp_external" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_amqp_external" "amqp_external" {
  url                = "https://example.com/webhooks"
  routingKey         = "message name: #{message.name}, clientId: #{message.clientId}"
  mandatoryRoute     = true
  persistentMessages = true
  messageTtl         = 0
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
  enveloped = true
  format    = "json"
}
