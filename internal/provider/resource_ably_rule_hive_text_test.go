// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleHiveText(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAblyRuleHiveTextConfig(appName, "/room-.*/", `model_url = "https://api.thehive.ai/api/v2/task/sync"`, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_hive_text.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_hive_text.rule0", "invocation_mode", "BEFORE_PUBLISH"),
					resource.TestCheckResourceAttr("ably_rule_hive_text.rule0", "chat_room_filter", "/room-.*/"),
					resource.TestCheckResourceAttr("ably_rule_hive_text.rule0", "target.model_url", "https://api.thehive.ai/api/v2/task/sync"),
					resource.TestCheckResourceAttr("ably_rule_hive_text.rule0", "target.thresholds.bullying", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ably_rule_hive_text.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIDFunc("ably_rule_hive_text.rule0"),
			},
			// Update and Read testing, dropping the optional model_url. An absent
			// optional string must come back as null rather than "".
			{
				Config: testAccAblyRuleHiveTextConfig(updateAppName, "/chat-.*/", "", 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_hive_text.rule0", "chat_room_filter", "/chat-.*/"),
					resource.TestCheckNoResourceAttr("ably_rule_hive_text.rule0", "target.model_url"),
					resource.TestCheckResourceAttr("ably_rule_hive_text.rule0", "target.thresholds.bullying", "3"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app and a Hive text moderation
// rule; extraTargetAttrs is spliced into the target block so the same config can
// be rendered with and without an optional attribute.
func testAccAblyRuleHiveTextConfig(
	appName string,
	chatRoomFilter string,
	extraTargetAttrs string,
	bullyingThreshold int,
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

resource "ably_rule_hive_text" "rule0" {
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
		api_key = "my-hive-api-key"
		%[3]s
		thresholds = {
			bullying = %[4]d
		}
	}
}
`, appName, chatRoomFilter, extraTargetAttrs, bullyingThreshold)
}
