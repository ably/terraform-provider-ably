resource "ably_rule" "amqp" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "amqp"
  request_mode = "single"
  source       = ably_rule_source.amqp
  target       = ably_rule_target_amqp.amqp
}

resource "ably_rule_source" "amqp" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_amqp" "amqp" {
  queue_id = "string"
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
