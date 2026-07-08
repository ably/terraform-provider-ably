// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"regexp"
	"strings"
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
			// ImportState testing. The API returns the full rule including the
			// target api_key (verified against production), so import verifies
			// every attribute with no ignores.
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

// TestAccAblyRuleBodyguardEmptyString verifies that explicit "" on optional
// string attributes is rejected at plan time. Empty and unset mean the same
// thing to the Control API, so without the validator a known "" in the plan
// reads back as null and aborts the apply with an opaque "inconsistent values
// for sensitive attribute" error that hides the culprit attribute.
func TestAccAblyRuleBodyguardEmptyString(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: strings.Replace(
					testAccAblyRuleBodyguardConfig(
						appName,
						"enabled",
						"/room-.*/",
						"RETRY",
						"my-bodyguard-api-key",
					),
					`channel_id       = "my-channel"`,
					`channel_id       = ""`,
					1,
				),
				ExpectError: regexp.MustCompile(`string length must be at least 1`),
			},
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
