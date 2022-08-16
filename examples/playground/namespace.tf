resource "ably_namespace" "namespace0" {
  app_id            = ably_app.app0.id
  id                = "namespace"
  authenticated     = false
  persisted         = false
  persist_last      = false
  push_enabled      = false
  tls_only          = false
  expose_timeserial = false
}
