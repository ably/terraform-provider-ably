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

data "ably_app" "app0" {
  id = ably_app.app0.id
}

output "ably_app_app0_status" {
  value = data.ably_app.app0.status
}

output "ably_app0_app_name" {
  value = data.ably_app.app0.name
}

output "ably_app0_tls_only" {
  value = data.ably_app.app0.tls_only
}

output "ably_app_account_id" {
  value = data.ably_app.app0.account_id
}

output "ably_app_app0_id" {
  value = data.ably_app.app0.id
}
