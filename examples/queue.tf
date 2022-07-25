# Queues

resource "ably_queue" "example_queue" {
  app_id     = ably_app.app1.app_id
  name       = "queue_name"
  ttl        = 60
  max_length = 10000
  region     = "us-east-1-a"
}
