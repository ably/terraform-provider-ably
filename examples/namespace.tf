# Namespaces

resource "ably_namespace" "namespace1" {
  app_id            = ably_app.app1.app_id
  namespace_id      = "namespace"
  authenticated     = false
  persisted         = false
  persist_last      = false
  push_enabled      = false
  tls_only          = false
  expose_timeserial = false
}
