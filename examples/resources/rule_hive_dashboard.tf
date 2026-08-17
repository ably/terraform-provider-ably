resource "ably_rule_hive_dashboard" "rule0" {
  app_id           = ably_app.app0.id
  status           = "enabled"
  invocation_mode  = "AFTER_PUBLISH"
  chat_room_filter = "/room-.*/"
  target = {
    api_key           = "my-hive-api-key"
    check_watch_lists = true
  }
}
