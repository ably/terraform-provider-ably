resource "ably_rule" "http_zapier" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "http/zapier"
  request_mode = "single"
  source       = ably_rule_source.http_zapier
  target       = ably_rule_target_http_zapier.http_zapier
}

resource "ably_rule_source" "http_zapier" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_http_zapier" "http_zapier" {
  url = "https://example.com/webhooks"
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
  signingKeyId = "bw66AB"
  enveloped    = true
  format       = "json"
}
