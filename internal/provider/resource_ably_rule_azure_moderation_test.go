// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleAzureModeration(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAblyRuleAzureModerationConfig(appName, "/room-.*/", "https://my-resource.cognitiveservices.azure.com", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_azure_moderation.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_azure_moderation.rule0", "invocation_mode", "BEFORE_PUBLISH"),
					resource.TestCheckResourceAttr("ably_rule_azure_moderation.rule0", "chat_room_filter", "/room-.*/"),
					resource.TestCheckResourceAttr("ably_rule_azure_moderation.rule0", "target.endpoint", "https://my-resource.cognitiveservices.azure.com"),
					resource.TestCheckResourceAttr("ably_rule_azure_moderation.rule0", "target.thresholds.Hate", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ably_rule_azure_moderation.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIDFunc("ably_rule_azure_moderation.rule0"),
			},
			// Update and Read testing
			{
				Config: testAccAblyRuleAzureModerationConfig(updateAppName, "/chat-.*/", "https://other-resource.cognitiveservices.azure.com", 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_azure_moderation.rule0", "chat_room_filter", "/chat-.*/"),
					resource.TestCheckResourceAttr("ably_rule_azure_moderation.rule0", "target.endpoint", "https://other-resource.cognitiveservices.azure.com"),
					resource.TestCheckResourceAttr("ably_rule_azure_moderation.rule0", "target.thresholds.Hate", "4"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app and an Azure moderation rule.
func testAccAblyRuleAzureModerationConfig(
	appName string,
	chatRoomFilter string,
	endpoint string,
	hateThreshold int,
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

resource "ably_rule_azure_moderation" "rule0" {
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
		api_key  = "my-azure-api-key"
		endpoint = %[3]q
		thresholds = {
			Hate = %[4]d
		}
	}
}
`, appName, chatRoomFilter, endpoint, hateThreshold)
}
