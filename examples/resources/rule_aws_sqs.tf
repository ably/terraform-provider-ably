resource "ably_rule_sqs" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  target = {
    region         = "us-west-1",
    aws_account_id = "123456789012",
    queue_name     = "aaaaaa",
    enveloped      = false,
    format         = "json"
    aws_authentication = {
      mode     = "assumeRole",
      role_arn = "cccc"
    }
  }
}
