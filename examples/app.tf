resource "ably_app" "app1" {
  name                      = "app_name"
  status                    = "enabled"
  tls_only                  = false
  fcm_key                   = "AABBQ1KyxCE:APA91bCCYs7r_Q-sqW8HMP_hV4t3vMYx...cJ8344-MhGWODZEuAmg_J4MUJcVQEyDn...I"
  apns_certificate          = "-----BEGIN CERTIFICATE-----MIIFaDCC...EXAMPLE...3Dc=-----END CERTIFICATE-----"
  apns_privateKey           = "-----BEGIN PRIVATE KEY-----ABCFaDCC...EXAMPLE...3Dc=-----END PRIVATE KEY-----"
  apns_use_sandbox_endpoint = false
}
