# Every API key in an app. Each entry includes its secret, so treat anything
# derived from this data source as a credential.
data "ably_api_keys" "all" {
  app_id = data.ably_app.existing.id
}

output "key_names" {
  value = [for key in data.ably_api_keys.all.keys : key.name]
}
