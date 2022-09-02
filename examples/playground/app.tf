resource "ably_app" "app0" {
  name     = "ably-tf-provider-app-0000"
  status   = "enabled"
  tls_only = true
}

resource "ably_app" "app1" {
  name     = "ably-tf-provider-app-0001"
  status   = "enabled"
  tls_only = true
}
