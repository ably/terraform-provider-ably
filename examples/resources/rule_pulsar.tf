resource "ably_rule" "pulsar" {
  enabled      = true
  app_id       = ably_app.app1.app_id
  rule_type    = "pulsar"
  request_mode = "single"
  source       = ably_rule_source.pulsar
  target       = ably_rule_target_pulsar.pulsar
}

resource "ably_rule_source" "pulsar" {
  channelFilter = "^my-channel.*"
  type          = "channel.message"
}

resource "ably_rule_target_pulsar" "pulsar" {
  routingKey = "test-key"
  topic      = "persistent://my-tenant/my-namespace/my-topic"
  serviceUrl = "pulsar://pulsar.us-west.example.com:6650/"
  tls_trust_certs = [
    "-----BEGIN CERTIFICATE----- MIIFiTCCA3GgAwIBAgIUYO1Lomxzj7VRawWwEFiQht9OLpUwDQYJKoZIhvcNAQEL BQAwTDELMAkGA1UEBhMCVVMxETAPBgNVBAgMCE1pY2hpZ2FuMQ8wDQYDVQQHDAZX ...snip... TOfReTlUQzgpXRW5h3n2LVXbXQhPGcVitb88Cm2R8cxQwgB1VncM8yvmKhREo2tz 7Y+sUx6eIl4dlNl9kVrH1TD3EwwtGsjUNlFSZhg= -----END CERTIFICATE-----"
  ]
  authentication = {
    authentication_mode = "token"
    "token" : "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
  }
  enveloped = true
  format    = "json"
}
