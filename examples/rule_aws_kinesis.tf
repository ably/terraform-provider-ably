resource "ably_rule" "aws_kinesis" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "aws/kinesis"
  request_mode = "single"
  source       = ably_rule_source.aws_kinesis
  target       = ably_rule_target_aws_kinesis.aws_kinesis
}

resource "ably_rule_source" "aws_kinesis" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_aws_kinesis" "aws_kinesis" {
  region        = "eu-west-1"
  stream_name   = "myStreamName"
  partition_key = "message name: #{message.name}, clientId: #{message.clientId}"
  authentication = {
    authentication_mode = "credentials"
    access_key_id       = "AKIAIOSFODNN7EXAMPLE"
    secret_access_key   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  enveloped = true
  format    = "json"
}
