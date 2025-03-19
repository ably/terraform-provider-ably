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
  id                = "namespace_batching"
  authenticated     = false
  persisted         = false
  persist_last      = false
  push_enabled      = false
  tls_only          = false
  expose_timeserial = false
  batching_enabled  = true
  batching_interval = 100
}

resource "ably_namespace" "namespace_conflation" {
  app_id              = ably_app.app0.id
  id                  = "namespace_conflation"
  authenticated       = false
  persisted           = false
  persist_last        = false
  push_enabled        = false
  tls_only            = false
  expose_timeserial   = false
  conflation_enabled  = true
  conflation_interval = 100
  conflation_key      = "test"
}
