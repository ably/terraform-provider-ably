resource "ably_rule" "http_google_cloud_function" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "http/google-cloud-function"
  request_mode = "single"
  source       = ably_rule_source.http_google_cloud_function
  target       = ably_rule_target_http_google_cloud_function.http_google_cloud_function
}

resource "ably_rule_source" "http_google_cloud_function" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_http_google_cloud_function" "http_google_cloud_function" {
  region        = "us-west1"
  project_id    = "my-sample-project-191923"
  function_name = "myFunctionName"
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
