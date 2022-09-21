resource "ably_rule_google_function" "google_function" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  target = {
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
    signing_key_id = ably_api_key.api_key_0.id
    project_id     = "foo"
    region         = "us"
    function_name  = "bar"
  }
}
