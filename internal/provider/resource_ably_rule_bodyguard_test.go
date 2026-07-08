// Package provider implements the Ably provider for Terraform
package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyRuleBodyguard(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccAblyRuleBodyguardConfig(
					appName,
					"enabled",
					"/room-.*/",
					"RETRY",
					"my-bodyguard-api-key",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "invocation_mode", "BEFORE_PUBLISH"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "chat_room_filter", "/room-.*/"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "before_publish_config.max_retries", "3"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "before_publish_config.too_many_requests_action", "RETRY"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "target.channel_id", "my-channel"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "target.default_language", "en"),
				),
			},
			// ImportState testing. The API returns the full rule including the
			// target api_key (verified against production), so import verifies
			// every attribute with no ignores.
			{
				ResourceName:      "ably_rule_bodyguard.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_bodyguard.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found: ably_rule_bodyguard.rule0")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},
			// Update and Read testing
			{
				Config: testAccAblyRuleBodyguardConfig(
					updateAppName,
					"enabled",
					"/chat-.*/",
					"FAIL",
					"my-bodyguard-api-key-updated",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "chat_room_filter", "/chat-.*/"),
					resource.TestCheckResourceAttr("ably_rule_bodyguard.rule0", "before_publish_config.too_many_requests_action", "FAIL"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// TestAccAblyRuleBodyguardEmptyString verifies that explicit "" on optional
// string attributes is rejected at plan time. Empty and unset mean the same
// thing to the Control API, so without the validator a known "" in the plan
// reads back as null and aborts the apply with an opaque "inconsistent values
// for sensitive attribute" error that hides the culprit attribute.
func TestAccAblyRuleBodyguardEmptyString(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: strings.Replace(
					testAccAblyRuleBodyguardConfig(
						appName,
						"enabled",
						"/room-.*/",
						"RETRY",
						"my-bodyguard-api-key",
					),
					`channel_id       = "my-channel"`,
					`channel_id       = ""`,
					1,
				),
				ExpectError: regexp.MustCompile(`string length must be at least 1`),
			},
		},
	})
}

// TestAccAblyRuleBodyguardClearChatRoomFilter is the regression test for the
// review PoC: removing chat_room_filter from config must actually clear it on
// the Control API. The rule PATCH schema can neither null the field nor
// accept "" (pattern ^/.*/$), so the plan must be a REPLACEMENT, and after
// apply the API record must not contain chatRoomFilter at all. Before the
// RequiresReplaceWhenCleared modifier, the PATCH body simply omitted the
// field, the API kept the old value, and state recorded null: a permanent,
// invisible divergence.
func TestAccAblyRuleBodyguardClearChatRoomFilter(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyRuleBodyguardConfigExtra(appName, `chat_room_filter = "/room-.*/"`),
				Check: resource.TestCheckResourceAttr(
					"ably_rule_bodyguard.rule0", "chat_room_filter", "/room-.*/"),
			},
			{
				// Same config with chat_room_filter removed.
				Config: testAccAblyRuleBodyguardConfigExtra(appName, ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"ably_rule_bodyguard.rule0", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"ably_rule_bodyguard.rule0", "chat_room_filter"),
					// The API record must agree with state: no chatRoomFilter.
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["ably_rule_bodyguard.rule0"]
						if !ok {
							return fmt.Errorf("rule not in state")
						}
						url := fmt.Sprintf("%s/apps/%s/rules/%s",
							os.Getenv("ABLY_URL"), rs.Primary.Attributes["app_id"], rs.Primary.ID)
						httpResp, err := http.Get(url)
						if err != nil {
							return err
						}
						defer httpResp.Body.Close()
						var rec map[string]any
						if err := json.NewDecoder(httpResp.Body).Decode(&rec); err != nil {
							return err
						}
						if v, ok := rec["chatRoomFilter"]; ok {
							return fmt.Errorf("state has chat_room_filter cleared but the API still holds chatRoomFilter = %q", v)
						}
						return nil
					},
				),
			},
		},
	})
}

// testAccAblyRuleBodyguardConfigExtra renders a minimal app + bodyguard rule;
// extraRuleAttrs is spliced into the rule block (e.g. a chat_room_filter
// line), so tests can render the same rule with and without an attribute.
func testAccAblyRuleBodyguardConfigExtra(appName, extraRuleAttrs string) string {
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

resource "ably_rule_bodyguard" "rule0" {
	app_id = ably_app.app0.id
	before_publish_config = {
		retry_timeout            = 5000
		max_retries              = 3
		failed_action            = "PUBLISH"
		too_many_requests_action = "RETRY"
	}
	target = {
		api_key          = "my-bodyguard-api-key"
		channel_id       = "my-channel"
		default_language = "en"
	}
	%[2]s
}
`, appName, extraRuleAttrs)
}

// Function with inline HCL to provision an ably_app and a bodyguard rule.
func testAccAblyRuleBodyguardConfig(
	appName string,
	ruleStatus string,
	chatRoomFilter string,
	tooManyRequestsAction string,
	apiKey string,
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

resource "ably_rule_bodyguard" "rule0" {
	app_id           = ably_app.app0.id
	status           = %[2]q
	invocation_mode  = "BEFORE_PUBLISH"
	chat_room_filter = %[3]q
	before_publish_config = {
		retry_timeout            = 5000
		max_retries              = 3
		failed_action            = "PUBLISH"
		too_many_requests_action = %[4]q
	}
	target = {
		api_key          = %[5]q
		channel_id       = "my-channel"
		default_language = "en"
	}
}
`, appName, ruleStatus, chatRoomFilter, tooManyRequestsAction, apiKey)
}
