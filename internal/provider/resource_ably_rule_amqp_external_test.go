// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					"my-exchange",
					true,
					true,
					44,
					originalHeadersBlock,
					"true",
					"json",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttrSet("ably_rule_amqp_external.rule0", "id"),
					resource.TestCheckResourceAttrSet("ably_rule_amqp_external.rule0", "app_id"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttrSet("ably_rule_amqp_external.rule0", "target.url"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.routing_key", "topic:key"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.exchange", "my-exchange"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.mandatory_route", "true"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.persistent_messages", "true"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.message_ttl", "44"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.headers.0.name", "User-Agent-Conf"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.headers.0.value", "user-agent-string"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.format", "json"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ably_rule_amqp_external.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_amqp_external.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{
					"target.url",
					"target.exchange",
				},
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
					"updated-exchange",
					false,
					false,
					23,
					updateHeadersBlock,
					"false",
					"msgpack",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttrSet("ably_rule_amqp_external.rule0", "id"),
					resource.TestCheckResourceAttrSet("ably_rule_amqp_external.rule0", "app_id"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttrSet("ably_rule_amqp_external.rule0", "target.url"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.routing_key", "newtopic:key"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.exchange", "updated-exchange"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.mandatory_route", "false"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.persistent_messages", "false"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.message_ttl", "23"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.headers.0.name", "User-Agent-Conf"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.headers.0.value", "user-agent-string"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.headers.1.name", "Custom-Header"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.headers.1.value", "custom-header-string"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "target.format", "msgpack"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccAblyRuleAMQPExternal_Minimal(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	config := minimalRuleConfig(appName, "ably_rule_amqp_external", `target = {
		url                 = "amqps://user:pass@example.com/vhost"
		routing_key         = "test-key"
		mandatory_route     = false
		persistent_messages = false
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ably_rule_amqp_external.rule0", "id"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp_external.rule0", "request_mode", "single"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
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
	targetExchange string,
	targetMandatoryRoute bool,
	targetPersistentMessages bool,
	targetMessageTTL int,
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
	  exchange = %[7]q
	  mandatory_route = %[8]t
	  persistent_messages = %[9]t
	  message_ttl = %[10]d
	  headers = %[11]s
	  enveloped = %[12]s,
	  format    = %[13]q,

	}
  }
`, appName, ruleStatus, channelFilter, sourceType, targetURL, targetRoutingKey, targetExchange, targetMandatoryRoute, targetPersistentMessages, targetMessageTTL, targetHeaders, targetEnveloped, targetFormat)
}
