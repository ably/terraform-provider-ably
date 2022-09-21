resource "ably_rule_ifttt" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  request_mode = "single"
  target = {
    webhook_key = "aaa",
    event_name  = "bbb"
  }
}
