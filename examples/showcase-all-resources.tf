# App

resource "ably_app" "app1" {
  status                 = "enabled"
  tlsOnly                = "false"
  fcmKey                 = "AABBQ1KyxCE:APA91bCCYs7r_Q-sqW8HMP_hV4t3vMYx...cJ8344-MhGWODZEuAmg_J4MUJcVQEyDn...I"
  apnsCertificate        = "-----BEGIN CERTIFICATE-----MIIFaDCC...EXAMPLE...3Dc=-----END CERTIFICATE-----"
  apnsPrivateKey         = "-----BEGIN PRIVATE KEY-----ABCFaDCC...EXAMPLE...3Dc=-----END PRIVATE KEY-----"
  apnsUseSandboxEndpoint = false
}

# Keys

resource "ably_api_key" "api_key_1" {
  app  = ably-app.app1.app_id
  name = "KeyName"
  capability = {
    channel1 = [""]
    channel2 = [""]
  }
}

# Namespaces

resource "ably_namespace" "namespace1" {
  app = ably-app.app1.app_id
  #...
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
  signingKeyId = "bw66AB"
  enveloped    = true
  format       = "json"
}

resource "ably_rule" "example_http_rule" {
  app         = ably-app.app1.app_id
  requestMode = "single"
  source      = ably-rule-source.example-rule-source-1
  target      = ably-http-rule-target.example-http-rule-target
}

resource "ably_aws_lambda_rule_target" "example-lambda-rule-target" {
  region       = "us-west-1"
  functionName = "myFunctionName"
  authentication = {
    authenticationMode = "credentials",
    accessKeyId        = "AKIAIOSFODNN7EXAMPLE"
    secretAccessKey    = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
  enveloped = true
}

resource "ably_rule" "example-http-rule" {
  app         = ably-app.app1.app_id
  requestMode = "single"
  source      = ably-rule-source.example-rule-source-1
  target      = ably-aws-lambda-rule-target.example-lambda-rule-target
}

resource "ably_kafka_rule_target" "example-kafka-rule-target" {
  routingKey = "partitionKey"
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

resource "ably_rule" "example-kafka-rule" {
  app         = ably-app.app1.app_id
  requestMode = "single"
  source      = ably-rule-source.example-rule-source-1
  target      = ably-kafka-rule-target.example-kafka-rule-target
}




