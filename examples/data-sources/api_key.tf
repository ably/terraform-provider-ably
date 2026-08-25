# Look up an existing API key by name. The result includes the key secret, so
# treat anything derived from it as a credential.
data "ably_api_key" "publisher" {
  app_id = data.ably_app.existing.id
  name   = "publisher"
}
