resource "ably_rule_lambda" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  target = {
    region        = "us-west-1",
    function_name = "rule0",
    enveloped     = false,
    format        = "json"
    authentication = {
      mode              = "credentials",
      access_key_id     = "hhhh"
      secret_access_key = "ffff"
    }
  }
}
