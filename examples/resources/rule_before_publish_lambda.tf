resource "ably_rule_before_publish_lambda" "rule0" {
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
    region        = "us-west-1"
    function_name = "my-moderation-function"
    authentication = {
      authentication_mode = "credentials"
      access_key_id       = "AKIAIOSFODNN7EXAMPLE"
      secret_access_key   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
    }
  }
}
