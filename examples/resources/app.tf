resource "ably_app" "app0" {
  name     = "ably-tf-provider-app-0000"
  status   = "enabled"
  tls_only = true
}

resource "ably_app" "app1" {
  name                      = "ably-tf-provider-app-0001"
  status                    = "enabled"
  tls_only                  = true
  fcm_key                   = "AABBQ1KyxCE:APA91bCCYs7r_Q-sqW8HMP_hV4t3vMYx...cJ8344-MhGWODZEuAmg_J4MUJcVQEyDn...I"
  apns_certificate          = "-----BEGIN CERTIFICATE-----MIIFaDCC...EXAMPLE...3Dc=-----END CERTIFICATE-----"
  apns_privateKey           = "-----BEGIN PRIVATE KEY-----ABCFaDCC...EXAMPLE...3Dc=-----END PRIVATE KEY-----"
  apns_use_sandbox_endpoint = false
}


