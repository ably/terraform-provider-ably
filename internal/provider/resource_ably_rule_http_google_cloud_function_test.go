package ably_control

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAblyRuleGoogleFunction(t *testing.T) {
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
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleGoogleFunctionConfig(
					app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					original_headers_block,
					"ably_api_key.api_key_0.id",
					"12345",
					"us",
					"bbbb",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "target.function_name", "bbbb"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "target.project_id", "12345"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleGoogleFunctionConfig(
					update_app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"batch",
					update_headers_block,
					"ably_api_key.api_key_1.id",
					"12345",
					"us",
					"bbbb",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", update_app_name),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "request_mode", "batch"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "target.function_name", "bbbb"),
					resource.TestCheckResourceAttr("ably_rule_google_function.rule0", "target.project_id", "12345"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleGoogleFunctionConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	requestMode string,
	targetHeaders string,
	targetSigningKeyId string,
	TargetProjectId string,
	TargetRegion string,
	TargetFunctionName string,
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

resource "ably_rule_google_function" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q
	source = {
	  channel_filter = %[3]q,
	  type           = %[4]q
	}
	request_mode = %[5]q
	target = {
	  headers = %[6]s
	  signing_key_id = %[7]s
	  project_id = %[8]q
	  region = %[9]q
	  function_name = %[10]q
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, requestMode, targetHeaders, targetSigningKeyId, TargetProjectId, TargetRegion, TargetFunctionName)
}
