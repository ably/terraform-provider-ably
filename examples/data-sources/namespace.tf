# Read the settings of a namespace (channel rule) this configuration does not
# manage. A namespace's ID is its channel name prefix.
data "ably_namespace" "chat" {
  app_id = data.ably_app.existing.id
  id     = "chat"
}

data "ably_namespaces" "all" {
  app_id = data.ably_app.existing.id
}
