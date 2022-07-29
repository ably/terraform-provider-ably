resource "ably_api_key" "api_key_1" {
  app_id = ably_app.app1.app_id
  name   = "KeyName"
}

resource "ably_api_key_capability" "channel1" {
  api_key = ably_api_key.api_key_1.name
  resource = "channel1"
  capabilities = ["publish", "subscribe"]
}

resource "ably_api_key_capability" "channel2" {
  api_key = ably_api_key.api_key_1.name
  resource = "channel2"
  capabilities = ["history"]
}
