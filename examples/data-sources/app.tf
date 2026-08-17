# Look up an app this configuration does not manage, by name.
data "ably_app" "existing" {
  name = "my-existing-app"
}

# Or by ID, if you have it.
data "ably_app" "by_id" {
  id = "abcdef"
}

# Every app in the account the provider's token belongs to.
data "ably_apps" "all" {}

output "app_ids" {
  value = [for app in data.ably_apps.all.apps : app.id]
}
