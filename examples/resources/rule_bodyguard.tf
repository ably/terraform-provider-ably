resource "ably_rule_bodyguard" "rule0" {
  app_id           = ably_app.app0.id
  status           = "enabled"
  invocation_mode  = "BEFORE_PUBLISH"
  chat_room_filter = "/room-.*/"
  before_publish_config = {
    retry_timeout            = 5000
    max_retries              = 3
    failed_action            = "PUBLISH"
    too_many_requests_action = "RETRY"
  }
  target = {
    api_key          = "my-bodyguard-api-key"
    channel_id       = "my-channel"
    default_language = "en"
  }
}
