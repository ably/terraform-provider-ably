package ably_control

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleAmqp(t *testing.T) {
	app_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	update_app_name := "acc-test-" + app_name

	original_headers_block := `[
	{
		name : "User-Agent-Conf",
		value : "user-agent-string",
	},
	]`
	update_headers_block := `[
	{
		name : "User-Agent-Conf",
		value : "user-agent-string",
	},
	{
		name: "Custom-Header",
		value : "custom-header-string",
	}
	]`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleAmqpConfig(
					app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					original_headers_block,
					"true",
					"json",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.format", "json"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleAmqpConfig(
					update_app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					update_headers_block,
					"false",
					"msgpack",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", update_app_name),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.format", "msgpack"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleAmqpConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	requestMode string,
	targetHeaders string,
	targetEnveloped string,
	targetFormat string,
) string {
	return fmt.Sprintf(`
terraform {
	required_providers {
		ably = {
		source = "github.com/ably/ably"
		}
	}
}

# You can provide your Ably Token & URL inline or use environment variables ABLY_ACCOUNT_TOKEN & ABLY_URL
provider "ably" {}

resource "ably_app" "app0" {
	name     = %[1]q
	status   = "enabled"
	tls_only = true
}

resource "ably_queue" "queue0" {
  app_id     = ably_app.app0.id
  name       = "queue_name"
  ttl        = 60
  max_length = 10000
  region     = "us-east-1-a"
}

resource "ably_rule_amqp" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q
	source = {
	  channel_filter = %[3]q,
	  type           = %[4]q
	}
	request_mode = %[5]q
	target = {
		queue_id = ably_queue.queue0.id
		headers = %[6]s
		enveloped = %[7]s
		format = %[8]q
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, requestMode, targetHeaders, targetEnveloped, targetFormat)
}
