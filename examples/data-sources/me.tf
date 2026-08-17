# Describes the token the provider is configured with, and the account it
# belongs to. Useful for referencing your account ID without hardcoding it.
data "ably_me" "current" {}

output "account_id" {
  value = data.ably_me.current.account.id
}
