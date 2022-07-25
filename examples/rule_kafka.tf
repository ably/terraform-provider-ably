resource "ably_rule" "kafka" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "kafka"
  request_mode = "single"
  source       = ably_rule_source.kafka
  target       = ably_rule_target_kafka.kafka
}

resource "ably_rule_source" "kafka" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_kafka" "kafka" {
  routingKey = "partitionKey"
  brokers = [
    "kafka.ci.ably.io:19092",
    "kafka.ci.ably.io:19093"
  ]
  authentication = {
    sasl = {
      mechanism = "scram-sha-256",
      username  = "username",
      password  = "password"
    }
  }
  enveloped = true
  format    = "json"
}
