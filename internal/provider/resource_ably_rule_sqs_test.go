// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyRuleSqs(t *testing.T) {
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
				Config: testAccAblyRuleSqsConfig(
					appName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					awsCredentialsAuthBlock,
					"us-west-1",
					"123456789012",
					"aaaa",
					"false",
					"json",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.region", "us-west-1"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.aws_account_id", "123456789012"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.queue_name", "aaaa"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.format", "json"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.authentication.mode", "credentials"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.authentication.access_key_id", "gggg"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.authentication.secret_access_key", "ffff"),
				),
			},
			// ImportState testing of ably_rule_sqs.rule0
			{
				ResourceName:      "ably_rule_sqs.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_sqs.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{"target.authentication"},
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleSqsConfig(
					updateAppName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					awsAssumeRoleAuthBlock,
					"us-east-1",
					"123456789012",
					"bbbb",
					"true",
					"json",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.region", "us-east-1"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.aws_account_id", "123456789012"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.queue_name", "bbbb"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.format", "json"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.authentication.mode", "assumeRole"),
					resource.TestCheckResourceAttr("ably_rule_sqs.rule0", "target.authentication.role_arn", "cccc"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleSqsConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	awsAuthBlock string,
	targetRegion string,
	awsAccountID string,
	queueName string,
	enveloped string,
	format string,
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

resource "ably_rule_sqs" "rule0" {
	app_id       = ably_app.app0.id
	status       = %[2]q
	source = {
	  channel_filter = %[3]q,
	  type           = %[4]q
	}

	target = {
	  region         = %[6]q,
	  aws_account_id = %[7]q,
	  queue_name     = %[8]q,
	  enveloped      = %[9]s,
	  format         = %[10]q
	  %[5]s
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, awsAuthBlock, targetRegion, awsAccountID, queueName, enveloped, format)
}
