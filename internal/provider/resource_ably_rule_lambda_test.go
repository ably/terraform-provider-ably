// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyRuleLambda(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName
	awsCredentialsAuthBlock := `authentication = {
		mode = "credentials",
		access_key_id = "gggg"
		secret_access_key = "ffff"
	}`

	awsAssumeRoleAuthBlock := `authentication = {
		mode = "assumeRole",
		role_arn = "cccc"
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleLambdaConfig(
					appName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					awsCredentialsAuthBlock,
					"us-west-1",
					"rule0-testing",
					"false",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.region", "us-west-1"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.function_name", "rule0-testing"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.mode", "credentials"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.access_key_id", "gggg"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.secret_access_key", "ffff"),
				),
			},
			// ImportState testing of ably_rule_lambda.rule0
			{
				ResourceName:      "ably_rule_lambda.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_lambda.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{
					"target.authentication.access_key_id",
					"target.authentication.secret_access_key",
				},
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleLambdaConfig(
					updateAppName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					awsAssumeRoleAuthBlock,
					"us-east-1",
					"rule0-testing",
					"true",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.region", "us-east-1"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.function_name", "rule0-testing"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.mode", "assumeRole"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.role_arn", "cccc"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccAblyRuleLambda_Minimal(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	config := minimalRuleConfig(appName, "ably_rule_lambda", `target = {
		region        = "us-west-1"
		function_name = "test-function"
		authentication = {
			mode              = "credentials"
			access_key_id     = "AKIAIOSFODNN7EXAMPLE"
			secret_access_key = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		}
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ably_rule_lambda.rule0", "id"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "request_mode", "single"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleLambdaConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	awsAuthBlock string,
	targetRegion string,
	functionName string,
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

resource "ably_rule_lambda" "rule0" {
	app_id       = ably_app.app0.id
	status       = %[2]q
	source = {
	  channel_filter = %[3]q,
	  type           = %[4]q
	}
	target = {
	  region        = %[6]q,
	  function_name   = %[7]q,
	  enveloped = %[8]s,
	  %[5]s
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, awsAuthBlock, targetRegion, functionName, enveloped)
}
