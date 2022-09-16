package ably_control

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestAccAblyRuleLambda(t *testing.T) {
	app_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	update_app_name := "acc-test-" + app_name
	aws_credentials_auth_block := `authentication = {
		mode = "credentials",
		access_key_id = "gggg"
		secret_access_key = "ffff"
	}`

	aws_assume_role_auth_block := `authentication = {
		mode = "assumeRole",
		role_arn = "cccc"
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleLambdaConfig(
					app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					aws_credentials_auth_block,
					"us-west-1",
					"rule0-testing",
					"false",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
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
					update_app_name,
					"enabled",
					"^my-channel.*",
					"channel.message",
					aws_assume_role_auth_block,
					"us-east-1",
					"rule0-testing",
					"true",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", update_app_name),
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
