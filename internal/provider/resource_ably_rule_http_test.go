// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyRuleHTTP(t *testing.T) {
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
		name : "User-Agent-Conf",
		value : "user-agent-string",
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
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleHTTPConfig(
					appName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"single",
					originalHeadersBlock,
					"ably_api_key.api_key_0.id",
					"https://example.com/webhooks",
					"json",
					"true",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.url", "https://example.com/webhooks"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.format", "json"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.headers.0.name", "User-Agent-Conf"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.headers.0.value", "user-agent-string"),
					resource.TestCheckResourceAttrSet("ably_rule_http.rule0", "target.signing_key_id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ably_rule_http.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_http.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found: ably_rule_http.rule0")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{"target.signing_key_id"},
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleHTTPConfig(
					updateAppName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"batch",
					updateHeadersBlock,
					"ably_api_key.api_key_1.id",
					"https://example1.com/webhooks",
					"msgpack",
					"false",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "request_mode", "batch"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.url", "https://example1.com/webhooks"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.format", "msgpack"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.headers.0.name", "User-Agent-Conf"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.headers.0.value", "user-agent-string"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.headers.1.name", "Custom-Header"),
					resource.TestCheckResourceAttr("ably_rule_http.rule0", "target.headers.1.value", "custom-header-string"),
					resource.TestCheckResourceAttrSet("ably_rule_http.rule0", "target.signing_key_id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleHTTPConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	requestMode string,
	targetHeaders string,
	targetSigningKeyID string,
	targetURL string,
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

resource "ably_rule_http" "rule0" {
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
	  url = %[8]q
	  format = %[9]q
	  enveloped = %[10]q
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, requestMode, targetHeaders, targetSigningKeyID, targetURL, targetFormat, enveloped)
}

func TestAccAblyRule_InvalidStatus(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
terraform {
	required_providers {
		ably = {
			source = "registry.terraform.io/ably/ably"
		}
	}
}
provider "ably" {}
resource "ably_app" "app0" { name = "test-negative-status" }
resource "ably_rule_http" "rule0" {
	app_id = ably_app.app0.id
	status = "invalid"
	source = { channel_filter = "^test", type = "channel.message" }
	target = { url = "https://example.com/webhook", format = "json" }
}
`,
				ExpectError: regexp.MustCompile(`.*value must be one of.*`),
			},
		},
	})
}

func TestAccAblyRule_InvalidRequestMode(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
terraform {
	required_providers {
		ably = {
			source = "registry.terraform.io/ably/ably"
		}
	}
}
provider "ably" {}
resource "ably_app" "app0" { name = "test-negative-reqmode" }
resource "ably_rule_http" "rule0" {
	app_id       = ably_app.app0.id
	request_mode = "invalid"
	source = { channel_filter = "^test", type = "channel.message" }
	target = { url = "https://example.com/webhook", format = "json" }
}
`,
				ExpectError: regexp.MustCompile(`.*value must be one of.*`),
			},
		},
	})
}

func TestAccAblyRule_InvalidFormat(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
terraform {
	required_providers {
		ably = {
			source = "registry.terraform.io/ably/ably"
		}
	}
}
provider "ably" {}
resource "ably_app" "app0" { name = "test-negative-format" }
resource "ably_rule_http" "rule0" {
	app_id = ably_app.app0.id
	source = { channel_filter = "^test", type = "channel.message" }
	target = { url = "https://example.com/webhook", format = "xml" }
}
`,
				ExpectError: regexp.MustCompile(`.*value must be one of.*`),
			},
		},
	})
}
