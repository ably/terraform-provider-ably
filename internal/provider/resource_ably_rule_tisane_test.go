// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyRuleTisane(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAblyRuleTisaneConfig(appName, "/room-.*/", "RETRY", "en", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "invocation_mode", "BEFORE_PUBLISH"),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "chat_room_filter", "/room-.*/"),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "before_publish_config.too_many_requests_action", "RETRY"),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "target.default_language", "en"),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "target.thresholds.abuse", "2"),
				),
			},
			// ImportState testing. The API returns the full target for moderation
			// rules, api_key included, so import verifies every attribute.
			{
				ResourceName:      "ably_rule_tisane.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIDFunc("ably_rule_tisane.rule0"),
			},
			// Update and Read testing
			{
				Config: testAccAblyRuleTisaneConfig(updateAppName, "/chat-.*/", "FAIL", "fr", 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "chat_room_filter", "/chat-.*/"),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "before_publish_config.too_many_requests_action", "FAIL"),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "target.default_language", "fr"),
					resource.TestCheckResourceAttr("ably_rule_tisane.rule0", "target.thresholds.abuse", "3"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app and a Tisane moderation rule.
func testAccAblyRuleTisaneConfig(
	appName string,
	chatRoomFilter string,
	tooManyRequestsAction string,
	defaultLanguage string,
	abuseThreshold int,
) string {
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

resource "ably_rule_tisane" "rule0" {
	app_id           = ably_app.app0.id
	status           = "enabled"
	invocation_mode  = "BEFORE_PUBLISH"
	chat_room_filter = %[2]q
	before_publish_config = {
		retry_timeout            = 5000
		max_retries              = 3
		failed_action            = "PUBLISH"
		too_many_requests_action = %[3]q
	}
	target = {
		api_key          = "my-tisane-api-key"
		default_language = %[4]q
		thresholds = {
			abuse = %[5]d
		}
	}
}
`, appName, chatRoomFilter, tooManyRequestsAction, defaultLanguage, abuseThreshold)
}

// importStateIDFunc builds the "app_id,id" import ID every rule resource takes.
func importStateIDFunc(resourceAddress string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceAddress]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceAddress)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
	}
}
