# Keys

resource "ably_api_key" "api_key_1" {
  app_id = ably_app.app1.app_id
  name   = "KeyName"
  capability = {
    channel1 = ["publish", "subscribe"]
    channel2 = ["history"]
  }
}
