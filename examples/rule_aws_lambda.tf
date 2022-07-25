resource "ably_rule" "aws_lambda" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "aws/lambda"
  request_mode = "single"
  source       = ably_rule_source.aws_lambda
  target       = ably_rule_target_aws_lambda.aws_lambda
}

resource "ably_rule_source" "aws_lambda" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_aws_lambda" "aws_lambda" {
  region        = "eu-west-1"
  function_name = "myFunctionName"
  authentication = {
    authentication_mode = "credentials"
    access_key_id       = "AKIAIOSFODNN7EXAMPLE"
    secret_access_key   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  enveloped = true
}
