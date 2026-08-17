// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleHiveDashboard(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAblyRuleHiveDashboardConfig(appName, "/room-.*/", "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_hive_dashboard.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_hive_dashboard.rule0", "invocation_mode", "AFTER_PUBLISH"),
					resource.TestCheckResourceAttr("ably_rule_hive_dashboard.rule0", "chat_room_filter", "/room-.*/"),
					resource.TestCheckResourceAttr("ably_rule_hive_dashboard.rule0", "target.check_watch_lists", "true"),
					// Hive dashboard rules carry no before_publish_config; the
					// schema must not have grown one from a copied family.
					resource.TestCheckNoResourceAttr("ably_rule_hive_dashboard.rule0", "before_publish_config"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ably_rule_hive_dashboard.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIDFunc("ably_rule_hive_dashboard.rule0"),
			},
			// Update and Read testing
			{
				Config: testAccAblyRuleHiveDashboardConfig(updateAppName, "/chat-.*/", "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_hive_dashboard.rule0", "chat_room_filter", "/chat-.*/"),
					resource.TestCheckResourceAttr("ably_rule_hive_dashboard.rule0", "target.check_watch_lists", "false"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app and a Hive dashboard rule.
func testAccAblyRuleHiveDashboardConfig(appName, chatRoomFilter, checkWatchLists string) string {
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

resource "ably_rule_hive_dashboard" "rule0" {
	app_id           = ably_app.app0.id
	status           = "enabled"
	invocation_mode  = "AFTER_PUBLISH"
	chat_room_filter = %[2]q
	target = {
		api_key           = "my-hive-api-key"
		check_watch_lists = %[3]s
	}
}
`, appName, chatRoomFilter, checkWatchLists)
}
