# Every queue in an app, with connection details and live message counts.
data "ably_queues" "all" {
  app_id = data.ably_app.existing.id
}

output "queue_backlogs" {
  value = { for queue in data.ably_queues.all.queues : queue.name => queue.messages.total }
}
