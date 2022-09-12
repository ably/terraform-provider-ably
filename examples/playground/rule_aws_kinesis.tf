resource "ably_rule_kinesis" "rule0" {
  app_id    = ably_app.app0.id
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
  # aws_authentication = {
  #     mode = "assumeRole",
  #     role_arn = "cccc"
  # }
  target = {
    region        = "us-west-1",
    stream_name   = "rule0",
    partition_key = "message name: #{message.name}, clientId: #{message.clientId}",
    enveloped     = false,
    format        = "json"
  }
}
