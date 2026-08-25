# Look up an app this configuration does not manage, by name.
data "ably_app" "existing" {
  name = "my-existing-app"
}

# Or by ID, if you have it.
data "ably_app" "by_id" {
  id = "abcdef"
}
