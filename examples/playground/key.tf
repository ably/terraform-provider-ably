resource "ably_api_key" "api_key_0" {
  app_id           = ably_app.app0.id
  name             = "key-0000"
  revocable_tokens = true
  capabilities = {
    "channel2"  = ["publish"],
    "channel3"  = ["subscribe"],
    "channel33" = ["subscribe"],
  }
}

resource "ably_api_key" "api_key_1" {
  app_id           = ably_app.app0.id
  name             = "key-0001"
  revocable_tokens = false
  capabilities = {
    "channel1" = ["subscribe"],
    "channel2" = ["publish"],
    "channel3" = ["subscribe"],
  }
}
