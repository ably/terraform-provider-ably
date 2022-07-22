# App

resource "ably_app" "app1" {
  name                      = "app_name"
  status                    = "enabled"
  tls_only                  = "false"
  fcm_key                   = "AABBQ1KyxCE:APA91bCCYs7r_Q-sqW8HMP_hV4t3vMYx...cJ8344-MhGWODZEuAmg_J4MUJcVQEyDn...I"
  apns_certificate          = "-----BEGIN CERTIFICATE-----MIIFaDCC...EXAMPLE...3Dc=-----END CERTIFICATE-----"
  apns_privateKey           = "-----BEGIN PRIVATE KEY-----ABCFaDCC...EXAMPLE...3Dc=-----END PRIVATE KEY-----"
  apns_use_sandbox_endpoint = false
}

# Keys

resource "ably_api_key" "api_key_1" {
  app_id = ably_app.app1.app_id
  name   = "KeyName"
  capability = {
    channel1 = [""]
    channel2 = [""]
  }
}

# Namespaces

resource "ably_namespace" "namespace1" {
  app_id            = ably_app.app1.app_id
  namespace_id      = "namespace"
  authenticated     = false
  persisted         = false
  persist_last      = false
  push_enabled      = false
  tls_only          = false
  expose_timeserial = false
}

# Rules

resource "ably_rule_source" "example_rule_source_1" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_http_rule_target" "example_http_rule_target" {
  url = "https://example.com/webhooks"
  headers = [
    {
      name  = "User-Agent",
      value = "user-agent-string"
    },
    {
      name  = "headerName",
      value = "headerValue"
    }
  ]
  signing_key_id = "bw66AB"
  enveloped      = true
  format         = "json"
}

resource "ably_rule" "example_http_rule" {
  app_id       = ably_app.app1.app_id
  request_mode = "single"
  source       = ably_rule_source.example_rule_source_1
  target       = ably_http_rule_target.example_http_rule_target
}

resource "ably_aws_lambda_rule_target" "example_lambda_rule_target" {
  region        = "us-west-1"
  function_name = "myFunctionName"
  authentication = {
    authentication_mode = "credentials",
    access_key_id       = "AKIAIOSFODNN7EXAMPLE"
    secret_access_key   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  enveloped = true
}

resource "ably_rule" "example_http_rule" {
  app_id       = ably_app.app1.app_id
  request_mode = "single"
  source       = ably_rule_source.example_rule_source_1
  target       = ably_aws_lambda_rule_target.example_lambda_rule_target
}

resource "ably_kafka_rule_target" "example_kafka_rule_target" {
  routing_key = "partitionKey"
  brokers = [
    "kafka.ci.ably.io:19092",
    "kafka.ci.ably.io:19093"
  ]
  auth = {
    sasl = {
      mechanism = "scram-sha-256",
      username  = "username",
      password  = "password"
    }
  }
  enveloped = true
  format    = "json"
}

resource "ably_rule" "example_kafka_rule" {
  app_id       = ably_app.app1.app_id
  request_mode = "single"
  source       = ably_rule_source.example_rule_source_1
  target       = ably_kafka_rule_target.example_kafka_rule_target
}
