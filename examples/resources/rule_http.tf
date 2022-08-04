resource "ably_rule" "http_standard" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "http"
  request_mode = "single"
  source       = ably_rule_source.http_standard
  target       = ably_rule_target_http.http_standard
}

resource "ably_rule_source" "http_standard" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_http" "http_standard" {
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
  signing_key_id = "bw66AB"
  enveloped      = true
  format         = "json"
}
