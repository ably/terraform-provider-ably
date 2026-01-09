// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleAMQPExternal(t *testing.T) {
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
				Config: testAccAblyRuleAMQPExternalConfig(
					appName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"amqps://test.example",
					"topic:key",
					true,
					true,
					44,
					originalHeadersBlock,
					"true",
					"json",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.routing_key", "topic:key"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.format", "json"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleAMQPExternalConfig(
					updateAppName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"amqps://test.example",
					"newtopic:key",
					false,
					false,
					23,
					updateHeadersBlock,
					"false",
					"msgpack",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.routing_key", "newtopic:key"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.format", "msgpack"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleAMQPExternalConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	targetURL string,
	targetRoutingKey string,
	targetMandatoryRoute bool,
	targetPersistentMessages bool,
	targetMessageTTL int,
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

resource "ably_rule_amqp_external" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q
	source = {
	  channel_filter =  %[3]q,
	  type           = %[4]q
	}
	target = {
	  url = %[5]q
	  routing_key = %[6]q,
	  mandatory_route = %[7]t
	  persistent_messages = %[8]t
	  message_ttl = %[9]d
	  headers = %[10]s
	  enveloped = %[11]s,
	  format    = %[12]q,

	}
  }
`, appName, ruleStatus, channelFilter, sourceType, targetURL, targetRoutingKey, targetMandatoryRoute, targetPersistentMessages, targetMessageTTL, targetHeaders, targetEnveloped, targetFormat)
}
