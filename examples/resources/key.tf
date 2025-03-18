resource "ably_api_key" "api_key_1" {
  app_id = ably_app.app1.id
  name   = "key-0001"
  capabilities = {
    "channel1" = ["publish", "subscribe"],
    "channel2" = ["history"],
    "channel3" = ["subscribe"],
  }
}
