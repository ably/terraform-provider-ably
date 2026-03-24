// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyRuleAzureFunction(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName
	originalHeadersBlock := `[
	{
		name : "User-Agent-Conf",
		value : "user-agent-string",
	},
	]`
	updateHeadersBlock := `[
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyRuleAzureFunctionConfig(
					appName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					"demo",
					"function0",
					originalHeadersBlock,
					"ably_api_key.api_key_0.id",
					"json",
					"true",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.azure_app_id", "demo"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.function_name", "function0"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.0.name", "User-Agent-Conf"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.0.value", "user-agent-string"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.format", "json"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttrSet("ably_rule_azure_function.rule0", "target.signing_key_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ably_rule_azure_function.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_azure_function.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found: ably_rule_azure_function.rule0")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{"target.signing_key_id"},
			},
			{
				Config: testAccAblyRuleAzureFunctionConfig(
					updateAppName,
					"disabled",
					"^my-channel.*",
					"channel.presence",
					"single",
					"demo",
					"function1",
					updateHeadersBlock,
					"ably_api_key.api_key_1.id",
					"msgpack",
					"false",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "status", "disabled"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "source.type", "channel.presence"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.azure_app_id", "demo"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.function_name", "function1"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.0.name", "User-Agent-Conf-Update"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.0.value", "user-agent-string-update"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.1.name", "Custom-Header"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.headers.1.value", "custom-header-string"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.format", "msgpack"),
					resource.TestCheckResourceAttr("ably_rule_azure_function.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttrSet("ably_rule_azure_function.rule0", "target.signing_key_id"),
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
	channelFilter string,
	sourceType string,
	requestMode string,
	targetAzureAppID string,
	targetAzureFunctionName string,
	targetHeaders string,
	targetSigningKeyID string,
	targetFormat string,
	enveloped string,
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

resource "ably_rule_azure_function" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q
	source = {
	  channel_filter = %[3]q,
	  type = %[4]q
	}
	request_mode = %[5]q
	target = {
	  azure_app_id = %[6]q,
	  function_name = %[7]q,
	  headers = %[8]s
	  signing_key_id = %[9]s
	  format = %[10]q
	  enveloped = %[11]s
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, requestMode, targetAzureAppID, targetAzureFunctionName, targetHeaders, targetSigningKeyID, targetFormat, enveloped)
}
