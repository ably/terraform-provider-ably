// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyRuleBodyguard(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAblyRuleBodyguardConfig(
					appName,
					"enabled",
					"/room-.*/",
					"RETRY",
					"my-bodyguard-api-key",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "invocation_mode", "BEFORE_PUBLISH"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "chat_room_filter", "/room-.*/"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "before_publish_config.max_retries", "3"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "before_publish_config.too_many_requests_action", "RETRY"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "target.channel_id", "my-channel"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "target.default_language", "en"),
				),
			},
			// ImportState testing. api_key is write-only and never returned by the
			// API, so it cannot be verified on import.
			{
				ResourceName:      "ably_rule_bodyguard.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_bodyguard.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found: ably_rule_bodyguard.rule0")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{
					"target.api_key",
				},
			},
			// Update and Read testing
			{
				Config: testAccAblyRuleBodyguardConfig(
					updateAppName,
					"enabled",
					"/chat-.*/",
					"FAIL",
					"my-bodyguard-api-key-updated",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "chat_room_filter", "/chat-.*/"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "before_publish_config.too_many_requests_action", "FAIL"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app and a bodyguard rule.
func testAccAblyRuleBodyguardConfig(
	appName string,
	ruleStatus string,
	chatRoomFilter string,
	tooManyRequestsAction string,
	apiKey string,
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

resource "ably_rule_bodyguard" "rule0" {
	app_id           = ably_app.app0.id
	status           = %[2]q
	invocation_mode  = "BEFORE_PUBLISH"
	chat_room_filter = %[3]q
	before_publish_config = {
		retry_timeout            = 5000
		max_retries              = 3
		failed_action            = "PUBLISH"
		too_many_requests_action = %[4]q
	}
	target = {
		api_key          = %[5]q
		channel_id       = "my-channel"
		default_language = "en"
	}
}
`, appName, ruleStatus, chatRoomFilter, tooManyRequestsAction, apiKey)
}
