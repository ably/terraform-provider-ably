# Every app in the account the provider's token belongs to.
data "ably_apps" "all" {}

output "app_ids" {
  value = [for app in data.ably_apps.all.apps : app.id]
}

# A different account, if the token can see it.
data "ably_apps" "other_account" {
  account_id = "abcdef"
}
