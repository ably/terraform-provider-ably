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

resource "ably_namespace" "namespace_batching" {
  app_id            = ably_app.app0.id
  id                = "namespace"
  authenticated     = false
  persisted         = false
  persist_last      = false
  push_enabled      = false
  tls_only          = false
  expose_timeserial = false
  batching_enabled  = true
  batching_policy   = "some-policy"
  batching_interval = 100
}
