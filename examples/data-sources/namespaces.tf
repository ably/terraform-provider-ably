# Every namespace (channel rule) in an app, including ones Terraform does not
# manage.
data "ably_namespaces" "all" {
  app_id = data.ably_app.existing.id
}

output "persisted_namespaces" {
  value = [for ns in data.ably_namespaces.all.namespaces : ns.id if ns.persisted]
}
