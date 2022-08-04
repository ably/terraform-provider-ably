resource "ably_rule" "http_ifttt" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "http/ifttt"
  request_mode = "single"
  source       = ably_rule_source.http_ifttt
  target       = ably_rule_target_http_ifttt.http_ifttt
}

resource "ably_rule_source" "http_ifttt" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_http_ifttt" "http_ifttt" {
  webhook_key = "aBcd12Ef98-Z1ab3yTe-EXAMPLE"
  event_name  = "MyAppletName"
}
