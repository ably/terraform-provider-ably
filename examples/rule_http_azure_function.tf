resource "ably_rule" "http_azure_function" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "http/azure-function"
  request_mode = "single"
  source       = ably_rule_source.http_azure_function
  target       = ably_rule_target_http_azure_function.http_azure_function
}

resource "ably_rule_source" "http_azure_function" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_http_azure_function" "http_azure_function" {
  azure_app_id        = "d1e9f419-c438-6032b32df979"
  azure_function_name = "myFunctionName"
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
