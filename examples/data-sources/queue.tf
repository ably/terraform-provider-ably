# Look up a queue and use its AMQP connection details.
data "ably_queue" "events" {
  app_id = data.ably_app.existing.id
  name   = "events"
}

output "events_queue_amqp_uri" {
  value     = data.ably_queue.events.amqp.uri
  sensitive = true
}

data "ably_queues" "all" {
  app_id = data.ably_app.existing.id
}
