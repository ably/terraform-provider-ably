resource "ably_rule_amqp" "rule0" {
	app_id = ably_app.app0.id
	status = "enabled"
	source = {
		channel_filter = "^my-channel.*",
		type           = "channel.message"
	}
	request_mode = "single"
	target = {
		queue_id = ably_queue.example_queue.id
    headers = [
      {
        name : "User-Agent",
        value : "user-agent-string",
      },
      {
        name : "User-Agent-Extra",
        value : "user-agent-string",
      },
    ]

		enveloped = false
		format = "json"
	}
  }

