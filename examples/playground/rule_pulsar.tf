# NOTE: This is a dummy certificate generated for testing
data "tls_certificate" "example" {
  content = <<EOT
-----BEGIN CERTIFICATE-----
MIIFqDCCA5ACCQDLT5J/mNSX4TANBgkqhkiG9w0BAQsFADCBlTELMAkGA1UEBhMC
R0IxDzANBgNVBAgMBkxvbmRvbjEPMA0GA1UEBwwGTG9uZG9uMRAwDgYDVQQKDAdU
ZXN0T3JnMQwwCgYDVQQLDANTREsxIzAhBgNVBAMMGnB1bHNhci51cy13ZXN0LmV4
YW1wbGUuY29tMR8wHQYJKoZIhvcNAQkBFhB0ZXN0QGV4YW1wbGUuY29tMB4XDTIy
MDkyNjEwNTU0MFoXDTIzMDkyNjEwNTU0MFowgZUxCzAJBgNVBAYTAkdCMQ8wDQYD
VQQIDAZMb25kb24xDzANBgNVBAcMBkxvbmRvbjEQMA4GA1UECgwHVGVzdE9yZzEM
MAoGA1UECwwDU0RLMSMwIQYDVQQDDBpwdWxzYXIudXMtd2VzdC5leGFtcGxlLmNv
bTEfMB0GCSqGSIb3DQEJARYQdGVzdEBleGFtcGxlLmNvbTCCAiIwDQYJKoZIhvcN
AQEBBQADggIPADCCAgoCggIBAMx7TXLIvBh+CQDat9PIUlTLAFSR6KAJe5j659O1
17Lyue2QcnpOTYAf5QOYyvNmC91l/KoAlPVr6DRig2vZAB/cHDans95+CRJfzA8r
pTwHbT2C9a14tTY0T4E5GAEGEBU7tv5fgfD0smwTtv6eiJ4In9EzQO3p0OLOsAeD
pxTnLoLSTMoTUTgg3v5A8BBtzTb3lI3+HDxGe8anb5c5cVirRca5KSkQNZR+QBPg
9KF6RTsEKhdq+ptteHFbIEw0cM5MitEyeWFmG2kf4V3SX+8+Ntrf1EopGenRCJEj
bZit8vOPI43kgP0mGHOzoQQRnhGyTNmjtE+Z2xxEzs7eYSXQa8kxO+kb37mAwRuX
RhAfsL8oj8Hxs+UmJk1F+XJIma37F/JBW671R+L7vZmJE3OLM53IwmtrELFoxLsi
oc8urBM6onSe5ZxZ8B+VLGkVZpTJ1PWeqbKCsp6RCweuOxZXb5M0kz6HewXwLKNK
t4A4CqIfEngZR74HuH0r9G1Ql4YkhkWCsj4+9b5Uq21d/aZU2C3wTbbPWJ9khqVT
NjWWi78FyoC1HCjWYgKCK1SQsYcqhq2nWj+MbqMN13k5Qc85hjFcmCB1SiWH+gv5
XQLUbXZAN4DKuN/iVGLM33teBPp7yVZpZbNfdaTAQWLWiw1ROUEcyKRt+B2rp+F+
xP5VAgMBAAEwDQYJKoZIhvcNAQELBQADggIBADraBsNjnURUF6Zn/gTpF2nlqzhf
BYOhlUv+6k9q5IJlqYtFT7mo+EhWf8xbso4vWipEJPy85DTG7gr/P1gJC4FBIaOe
R0WIlwZukz/S1W9KJ4eeh3b92QjYn+Sbx1Mc8qUaZk45MsLZrpSyHsrbvXGQsDwj
CyRAexJN7gGMBteHMgfZGQINQe3Lya76rl2xPM4jsd8mWWASwT715fSiqRbWi7a/
XbTP/ENtUj5PRrHliXFL+6nCQa6y73Qdt2o3Ob6ZWlFywv3of2wKas1bYdE1ZxKw
5Br9/m1hhxrH52AnDuR9BfNIp3Z/eCFCXLI0WHsxBEEgPZfUmo5iwRKWrPVcVwLS
lTNCPTuMG/Fnl+MbXtvu30bVjLrH54yKFQEv39cPP9OpuC/YW/nW56eR3h4MLmqP
jX4y7IOBkUAczjZHPsfMM8DcemUYcswIjTtk8piz9YPDo3qNQGsnZNba4uDulQ0U
rfEDa9HzWB6hiJ02g+XssiSo9mbann0qU0ZWmCxiBDN5eMQYJ//RMZym0ccAu9Ug
xapS7YtDmqkq2FQdj++IFst0ktBvXDV8AVz4MuZwY9adSZFmwHonHomiLfgySPRR
cYzK74pwWRa5PWLzBXHU9oC316izLXBQO4OhUdJtaqwqNd22L4UFinQcJL12Up5c
4XIgNCFSXyfq8ZGj
-----END CERTIFICATE-----
	EOT
}

resource "ably_rule_pulsar" "rule0" {
  app_id = ably_app.app0.id
  status = "enabled"
  source = {
    channel_filter = "^my-channel.*",
    type           = "channel.message"
  }
  request_mode = "batch"
  target = {
    routing_key     = "test-key"
    topic           = "persistent://my-tenant/my-namespace/my-topic"
    service_url     = "pulsar://pulsar.us-west.example.com:6650/"
    tls_trust_certs = [data.tls_certificate.example.certificates[0].cert_pem]
    enveloped       = true
    format          = "json"
    authentication = {
      mode  = "token"
      token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
    }
  }
}
