// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
			// ImportState testing
			{
				ResourceName:      "ably_rule_ifttt.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_ifttt.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found: ably_rule_ifttt.rule0")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{"target.webhook_key"},
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleIFTTTConfig(
					updateAppName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					// IFTTT does not support batch mode.
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
