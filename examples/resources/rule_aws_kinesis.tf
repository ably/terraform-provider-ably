resource "ably_rule_kinesis" "aws_kinesis" {
  app_id    = ably_app.app1.id
  status    = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  aws_authentication = {
    mode              = "credentials",
    access_key_id     = "hhhh"
    secret_access_key = "ffff"
  }
  target = {
    region        = "us-west-1",
    stream_name   = "rule0",
    partition_key = "message name: #{message.name}, clientId: #{message.clientId}",
    enveloped     = false,
    format        = "json"
  }
}
