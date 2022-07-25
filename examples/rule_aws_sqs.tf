resource "ably_rule" "aws_sqs" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "aws/sqs"
  request_mode = "single"
  source       = ably_rule_source.aws_sqs
  target       = ably_rule_target_aws_sqs.aws_sqs
}

resource "ably_rule_source" "aws_sqs" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_aws_sqs" "aws_sqs" {
  region         = "eu-west-1"
  aws_account_id = "123456789012"
  queue_name     = "MyQueue"
  authentication = {
    authentication_mode = "credentials"
    access_key_id       = "AKIAIOSFODNN7EXAMPLE"
    secret_access_key   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  enveloped = true
  format    = "json"
}
