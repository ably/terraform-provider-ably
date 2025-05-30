// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.mode", "credentials"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.access_key_id", "gggg"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.secret_access_key", "ffff"),
				),
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
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.mode", "assumeRole"),
					resource.TestCheckResourceAttr("ably_rule_lambda.rule0", "target.authentication.role_arn", "cccc"),
				),
			},
			// Delete testing automatically occurs in TestCase
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
