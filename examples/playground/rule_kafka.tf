resource "ably_rule_kafka" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  target = {
    routing_key = "topic:key",
    brokers     = ["kafka.ci.ably.io:19092", "kafka.ci.ably.io:19093"]
    auth = {
      sasl = {
        mechanism = "scram-sha-256"
        username  = "username"
        password  = "password"
      }
    }
    enveloped = true
    format    = "json"
  }
}
