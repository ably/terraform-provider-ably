resource "ably_rule" "http_cloudflare_worker" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "http/cloudflare-worker"
  request_mode = "single"
  source       = ably_rule_source.http_cloudflare_worker
  target       = ably_rule_target_http_cloudflare_worker.http_cloudflare_worker
}

resource "ably_rule_source" "http_cloudflare_worker" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_http_cloudflare_worker" "http_cloudflare_worker" {
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
}
