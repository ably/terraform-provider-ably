// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/ably/terraform-provider-ably/control"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyRuleBeforePublishLambda(t *testing.T) {
	skipIfRuleTypeUnavailable(t, beforePublishLambdaRuleType, control.BeforePublishAWSLambdaRulePost{
		RuleType:            beforePublishLambdaRuleType,
		InvocationMode:      "BEFORE_PUBLISH",
		BeforePublishConfig: beforePublishProbeConfig(),
		Target: control.BeforePublishAWSLambdaTarget{
			Region:       "us-west-1",
			FunctionName: "my-moderation-function",
			Authentication: control.AWSAuthentication{
				AuthenticationMode: string(control.AWSAuthModeAssumeRole),
				AssumeRoleArn:      "arn:aws:iam::123456789012:role/ably-moderation",
			},
		},
	})

	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAblyRuleBeforePublishLambdaConfig(appName, "us-west-1", "my-moderation-function"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "invocation_mode", "BEFORE_PUBLISH"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "source.channel_filter", "^room:"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "target.region", "us-west-1"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "target.function_name", "my-moderation-function"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "target.authentication.authentication_mode", "credentials"),
					// secret_access_key is write-only on the Control API, so it
					// has to survive in state from the plan.
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "target.authentication.secret_access_key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
				),
			},
			// ImportState testing. secret_access_key cannot be imported (the API
			// never returns it), so it is skipped in the verify.
			{
				ResourceName:            "ably_rule_before_publish_lambda.rule0",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"target.authentication.secret_access_key"},
				ImportStateIdFunc:       importStateIDFunc("ably_rule_before_publish_lambda.rule0"),
			},
			// Update and Read testing
			{
				Config: testAccAblyRuleBeforePublishLambdaConfig(updateAppName, "eu-west-2", "my-other-function"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "target.region", "eu-west-2"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "target.function_name", "my-other-function"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccAblyRuleBeforePublishLambdaAssumeRole covers the other AWS auth mode:
// assumeRole sends only the ARN, and the credentials attributes must be absent
// from state rather than empty strings.
func TestAccAblyRuleBeforePublishLambdaAssumeRole(t *testing.T) {
	skipIfRuleTypeUnavailable(t, beforePublishLambdaRuleType, control.BeforePublishAWSLambdaRulePost{
		RuleType:            beforePublishLambdaRuleType,
		InvocationMode:      "BEFORE_PUBLISH",
		BeforePublishConfig: beforePublishProbeConfig(),
		Target: control.BeforePublishAWSLambdaTarget{
			Region:       "us-west-1",
			FunctionName: "my-moderation-function",
			Authentication: control.AWSAuthentication{
				AuthenticationMode: string(control.AWSAuthModeAssumeRole),
				AssumeRoleArn:      "arn:aws:iam::123456789012:role/ably-moderation",
			},
		},
	})

	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyRuleBeforePublishLambdaAssumeRoleConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "target.authentication.authentication_mode", "assumeRole"),
					resource.TestCheckResourceAttr("ably_rule_before_publish_lambda.rule0", "target.authentication.assume_role_arn", "arn:aws:iam::123456789012:role/ably-moderation"),
					resource.TestCheckNoResourceAttr("ably_rule_before_publish_lambda.rule0", "target.authentication.access_key_id"),
					resource.TestCheckNoResourceAttr("ably_rule_before_publish_lambda.rule0", "target.authentication.secret_access_key"),
				),
			},
		},
	})
}

// TestAccAblyRuleBeforePublishLambdaRejectsWebhookSource pins the source type at
// plan time. This rule shares its spec schema with the webhook/firehose source,
// whose documented values are channel.message and friends, but the API validates
// it against an undocumented chat-specific enum and rejects channel.message
// outright. The hermetic fake echoes whatever it is sent, so nothing but a
// plan-time validator catches this before a real apply does.
func TestAccAblyRuleBeforePublishLambdaRejectsWebhookSource(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: strings.Replace(
					testAccAblyRuleBeforePublishLambdaConfig(appName, "us-west-1", "my-moderation-function"),
					`type           = "chat.message"`,
					`type           = "channel.message"`,
					1,
				),
				ExpectError: regexp.MustCompile(`value must be one of: \["chat.message"\]`),
			},
		},
	})
}

// Function with inline HCL to provision an ably_app and a before-publish Lambda
// rule authenticating with AWS credentials.
func testAccAblyRuleBeforePublishLambdaConfig(appName, region, functionName string) string {
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

resource "ably_rule_before_publish_lambda" "rule0" {
	app_id          = ably_app.app0.id
	status          = "enabled"
	invocation_mode = "BEFORE_PUBLISH"
	before_publish_config = {
		retry_timeout            = 5000
		max_retries              = 3
		failed_action            = "PUBLISH"
		too_many_requests_action = "RETRY"
	}
	source = {
		channel_filter = "^room:"
		type           = "chat.message"
	}
	target = {
		region        = %[2]q
		function_name = %[3]q
		authentication = {
			authentication_mode = "credentials"
			access_key_id       = "AKIAIOSFODNN7EXAMPLE"
			secret_access_key   = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
		}
	}
}
`, appName, region, functionName)
}

// Function with inline HCL to provision an ably_app and a before-publish Lambda
// rule authenticating with an assumable role.
func testAccAblyRuleBeforePublishLambdaAssumeRoleConfig(appName string) string {
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

resource "ably_rule_before_publish_lambda" "rule0" {
	app_id          = ably_app.app0.id
	status          = "enabled"
	invocation_mode = "BEFORE_PUBLISH"
	before_publish_config = {
		retry_timeout            = 5000
		max_retries              = 3
		failed_action            = "PUBLISH"
		too_many_requests_action = "RETRY"
	}
	target = {
		region        = "us-west-1"
		function_name = "my-moderation-function"
		authentication = {
			authentication_mode = "assumeRole"
			assume_role_arn     = "arn:aws:iam::123456789012:role/ably-moderation"
		}
	}
}
`, appName)
}
