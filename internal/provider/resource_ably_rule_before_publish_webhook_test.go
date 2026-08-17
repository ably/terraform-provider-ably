// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/ably/terraform-provider-ably/control"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleBeforePublishWebhook(t *testing.T) {
	skipIfRuleTypeUnavailable(t, beforePublishWebhookRuleType, control.BeforePublishWebhookRulePost{
		RuleType:            beforePublishWebhookRuleType,
		InvocationMode:      "BEFORE_PUBLISH",
		BeforePublishConfig: beforePublishProbeConfig(),
		Target: control.BeforePublishWebhookTarget{
			URL: "https://example.com/moderate",
		},
	})

	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAblyRuleBeforePublishWebhookConfig(appName, "/room-.*/", "https://example.com/moderate", "custom-header-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_before_publish_webhook.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_webhook.rule0", "invocation_mode", "BEFORE_PUBLISH"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_webhook.rule0", "chat_room_filter", "/room-.*/"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_webhook.rule0", "target.url", "https://example.com/moderate"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_webhook.rule0", "target.headers.0.name", "X-Custom-Header"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_webhook.rule0", "target.headers.0.value", "custom-header-value"),
					// Before-publish rules are not webhook-shaped: no source, no
					// request_mode.
					resource.TestCheckNoResourceAttr("ably_rule_before_publish_webhook.rule0", "source"),
					resource.TestCheckNoResourceAttr("ably_rule_before_publish_webhook.rule0", "request_mode"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ably_rule_before_publish_webhook.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIDFunc("ably_rule_before_publish_webhook.rule0"),
			},
			// Update and Read testing
			{
				Config: testAccAblyRuleBeforePublishWebhookConfig(updateAppName, "/chat-.*/", "https://example.com/moderate-v2", "updated-header-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_before_publish_webhook.rule0", "chat_room_filter", "/chat-.*/"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_webhook.rule0", "target.url", "https://example.com/moderate-v2"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_webhook.rule0", "target.headers.0.value", "updated-header-value"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app and a before-publish webhook
// rule.
func testAccAblyRuleBeforePublishWebhookConfig(appName, chatRoomFilter, url, headerValue string) string {
	return fmt.Sprintf(`
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

resource "ably_rule_before_publish_webhook" "rule0" {
	app_id           = ably_app.app0.id
	status           = "enabled"
	invocation_mode  = "BEFORE_PUBLISH"
	chat_room_filter = %[2]q
	before_publish_config = {
		retry_timeout            = 5000
		max_retries              = 3
		failed_action            = "PUBLISH"
		too_many_requests_action = "RETRY"
	}
	target = {
		url = %[3]q
		headers = [
			{
				name  = "X-Custom-Header"
				value = %[4]q
			},
		]
	}
}
`, appName, chatRoomFilter, url, headerValue)
}
