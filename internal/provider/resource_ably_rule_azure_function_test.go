package ably_control

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccAblyRuleAzureFunction(t *testing.T) {
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
		name : "User-Agent-Conf-Update",
		value : "user-agent-string-update",
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
			{
				Config: testAccAblyRuleAzureFunctionConfig(
					app_name,
					"enabled",
					"channel.message",
					"batch",
					"coms",
					"function0",
					original_headers_block,
					"ably_api_key.api_key_0.id",
					"json",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "request_mode", "batch"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.function_name", "function0"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.0.name", "User-Agent-Conf"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.0.value", "user-agent-string"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.format", "json"),
				),
			},
			{
				Config: testAccAblyRuleAzureFunctionConfig(
					update_app_name,
					"disabled",
					"channel.presence",
					"batch",
					"coms",
					"function1",
					update_headers_block,
					"ably_api_key.api_key_1.id",
					"msgpack",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", update_app_name),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "status", "disabled"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "source.type", "channel.presence"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "request_mode", "batch"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.function_name", "function1"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.0.name", "User-Agent-Conf-Update"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.0.value", "user-agent-string-update"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.1.name", "Custom-Header"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.1.value", "custom-header-string"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.format", "msgpack"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleAzureFunctionConfig(
	appName string,
	ruleStatus string,
	sourceType string,
	requestMode string,
	targetAzureAppId string,
	targetAzureFunctionName string,
	targetHeaders string,
	targetSigningKeyId string,
	targetFormat string,
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
  }

  resource "ably_api_key" "api_key_1" {
	app_id = ably_app.app0.id
	name   = "key-0001"
	capabilities = {
	  "channel2"  = ["publish"],
	}
  }

resource "ably_rule_azure_function" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q
	source = {
	  type = %[3]q
	}
	request_mode = %[4]q
	target = {
	  azure_app_id = %[5]q,
	  function_name = %[6]q,
	  headers = %[7]s
	  signing_key_id = %[8]s
	  format = %[9]q
	}
  }
`, appName, ruleStatus, sourceType, requestMode, targetAzureAppId, targetAzureFunctionName, targetHeaders, targetSigningKeyId, targetFormat)
}
