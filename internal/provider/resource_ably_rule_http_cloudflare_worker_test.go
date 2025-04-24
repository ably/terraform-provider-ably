package ably_control

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleCloudflareWorker(t *testing.T) {
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleCloudflareWorkerConfig(
					app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					"https://example.com/webhooks",
					original_headers_block,
					"ably_api_key.api_key_0.id",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "target.url", "https://example.com/webhooks"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleCloudflareWorkerConfig(
					update_app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					// TODO: change to batch when control api not broken #147
					"single",
					"https://example1.com/webhooks",
					update_headers_block,
					"ably_api_key.api_key_1.id",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", update_app_name),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_cloudflare_worker.rule0", "target.url", "https://example1.com/webhooks"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleCloudflareWorkerConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	requestMode string,
	targetUrl string,
	targetHeaders string,
	targetSigningKeyId string,
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

resource "ably_rule_cloudflare_worker" "rule0" {
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
`, appName, ruleStatus, channelFilter, sourceType, requestMode, targetUrl, targetHeaders, targetSigningKeyId)
}
