// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleIFTTT(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleIFTTTConfig(
					appName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					"aaaa",
					"bbbb",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "target.webhook_key", "aaaa"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "target.event_name", "bbbb"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleIFTTTConfig(
					updateAppName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					// TODO: change to batch when control api not broken #147
					"single",
					"dddd",
					"eeee",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "target.webhook_key", "dddd"),
					resource.TestCheckResourceAttr("ably_rule_ifttt.rule0", "target.event_name", "eeee"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleIFTTTConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	requestMode string,
	targetWebhookKey string,
	targetEventName string,
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

resource "ably_api_key" "api_key_0" {
	app_id = ably_app.app0.id
	name   = "key-0000"
	capabilities = {
	  "channel2"  = ["publish"],
	  "channel3"  = ["subscribe"],
	  "channel33" = ["subscribe"],
	}
	revocable_tokens = true
  }

  resource "ably_api_key" "api_key_1" {
	app_id = ably_app.app0.id
	name   = "key-0001"
	capabilities = {
	  "channel2"  = ["publish"],
	}
	revocable_tokens = false
  }

resource "ably_rule_ifttt" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q
	source = {
	  channel_filter = %[3]q,
	  type           = %[4]q
	}
	request_mode = %[5]q
	target = {
	  webhook_key =	%[6]q,
	  event_name = %[7]q
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, requestMode, targetWebhookKey, targetEventName)
}
