// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleZapier(t *testing.T) {
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
				Config: testAccAblyRuleZapierConfig(
					appName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					"https://example.com/webhooks",
					originalHeadersBlock,
					"ably_api_key.api_key_0.id",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "target.url", "https://example.com/webhooks"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleZapierConfig(
					updateAppName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					// TODO: change to batch when control api not broken #147
					"single",
					"https://example1.com/webhooks",
					updateHeadersBlock,
					"ably_api_key.api_key_1.id",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_zapier.rule0", "target.url", "https://example1.com/webhooks"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleZapierConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	requestMode string,
	targetURL string,
	targetHeaders string,
	targetSigningKeyID string,
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

resource "ably_rule_zapier" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q
	source = {
	  channel_filter = %[3]q,
	  type           = %[4]q
	}
	request_mode = %[5]q
	target = {
	  url =	%[6]q,
	  headers = %[7]s
	  signing_key_id = %[8]s
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, requestMode, targetURL, targetHeaders, targetSigningKeyID)
}
