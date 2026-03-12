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

func TestAccAblyRuleKafka(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleKafkaConfig(
					appName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"topic:key",
					"[\"kafka.ci.ably.io:19092\", \"kafka.ci.ably.io:19093\"]",
					"scram-sha-256",
					"username",
					"password",
					"true",
					"json",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttrSet("ably_rule_kafka.rule0", "id"),
					resource.TestCheckResourceAttrSet("ably_rule_kafka.rule0", "app_id"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.routing_key", "topic:key"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.brokers.0", "kafka.ci.ably.io:19092"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.brokers.1", "kafka.ci.ably.io:19093"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.auth.sasl.mechanism", "scram-sha-256"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.auth.sasl.username", "username"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.auth.sasl.password", "password"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.format", "json"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "ably_rule_kafka.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_kafka.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{
					"target.auth.sasl.username",
					"target.auth.sasl.password",
				},
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyRuleKafkaConfig(
					updateAppName,
					"enabled",
					"^my-channel.*",
					"channel.message",
					"newtopic:key",
					"[\"kafka.ci.ably.io:19092\", \"kafka.ci.ably.io:19094\"]",
					"scram-sha-512",
					"username1",
					"password1",
					"false",
					"msgpack",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttrSet("ably_rule_kafka.rule0", "id"),
					resource.TestCheckResourceAttrSet("ably_rule_kafka.rule0", "app_id"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.routing_key", "newtopic:key"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.brokers.0", "kafka.ci.ably.io:19092"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.brokers.1", "kafka.ci.ably.io:19094"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.auth.sasl.mechanism", "scram-sha-512"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.auth.sasl.username", "username1"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.auth.sasl.password", "password1"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_kafka.rule0", "target.format", "msgpack"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyRuleKafkaConfig(
	appName string,
	ruleStatus string,
	channelFilter string,
	sourceType string,
	targetRoutingKey string,
	targetBrokers string,
	targetSaslMechanism string,
	targetSaslUsername string,
	targetSaslPassword string,
	targetEnveloped string,
	targetFormat string,
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

resource "ably_rule_kafka" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q
	source = {
	  channel_filter =  %[3]q,
	  type           = 	%[4]q
	}
	target = {
	  routing_key = %[5]q,
	  brokers     = %[6]s,
	  auth = {
		sasl = {
		  mechanism = %[7]q,
		  username  = %[8]q,
		  password  = %[9]q,
		}
	  }
	  enveloped = %[10]s,
	  format    = %[11]q,
	}
  }
`, appName, ruleStatus, channelFilter, sourceType, targetRoutingKey, targetBrokers, targetSaslMechanism, targetSaslUsername, targetSaslPassword, targetEnveloped, targetFormat)
}

func TestAccAblyKafka_InvalidSASLMechanism(t *testing.T) {
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
resource "ably_app" "app0" { name = "test-negative-kafka" }
resource "ably_rule_kafka" "rule0" {
	app_id = ably_app.app0.id
	source = { channel_filter = "^test", type = "channel.message" }
	target = {
		routing_key = "topic:key"
		brokers     = ["broker1:9092"]
		enveloped   = false
		format      = "json"
		auth = { sasl = { mechanism = "invalid", username = "user", password = "pass" } }
	}
}
`,
				ExpectError: regexp.MustCompile(`.*value must be one of.*`),
			},
		},
	})
}
