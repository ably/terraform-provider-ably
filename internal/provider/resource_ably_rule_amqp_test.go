// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyRuleAMQP(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	originalHeadersBlock := `[
	{
		name : "User-Agent-Conf",
		value : "user-agent-string",
	},
	]`
	updateHeadersBlock := `[
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleAMQPConfig(
					appName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					originalHeadersBlock,
					"true",
					"json",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttrSet("ably_rule_amqp.rule0", "id"),
					resource.TestCheckResourceAttrSet("ably_rule_amqp.rule0", "app_id"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttrSet("ably_rule_amqp.rule0", "target.queue_id"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.headers.0.name", "User-Agent-Conf"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.headers.0.value", "user-agent-string"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.format", "json"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ably_rule_amqp.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_amqp.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleAMQPConfig(
					updateAppName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					updateHeadersBlock,
					"false",
					"msgpack",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttrSet("ably_rule_amqp.rule0", "id"),
					resource.TestCheckResourceAttrSet("ably_rule_amqp.rule0", "app_id"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttrSet("ably_rule_amqp.rule0", "target.queue_id"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.headers.0.name", "User-Agent-Conf"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.headers.0.value", "user-agent-string"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.headers.1.name", "Custom-Header"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.headers.1.value", "custom-header-string"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_amqp.rule0", "target.format", "msgpack"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleAMQPConfig(
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
# You can provide your Ably Token & URL inline or use environment variables ABLY_ACCOUNT_TOKEN & ABLY_URL
terraform {
	required_providers {
		ably = {
			source = "registry.terraform.io/ably/ably"
		}
	}
}
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
