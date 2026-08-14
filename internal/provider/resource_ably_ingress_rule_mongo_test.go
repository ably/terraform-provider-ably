// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyIngressRuleMongo(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName
	updateMongoURL := "mongodb://me:lon@honeydew.io:27017"
	testMongoURL := "mongodb://coco:nut@coco.io:27017"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyIngressRuleMongoConfig(
					appName,
					"enabled",
					testMongoURL,
					"coconut",
					"coconut",
					"off",
					"off",
					"us-east-1-A",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "status", "enabled"),
					resource.TestCheckResourceAttrSet("ably_ingress_rule_mongodb.rule0", "target.url"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.database", "coconut"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.collection", "coconut"),
					resource.TestCheckResourceAttrSet("ably_ingress_rule_mongodb.rule0", "target.pipeline"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.full_document", "off"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.full_document_before_change", "off"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.primary_site", "us-east-1-A"),
				),
			},
			// ImportState testing of ably_ingress_rule_mongodb.rule0
			{
				ResourceName:      "ably_ingress_rule_mongodb.rule0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_ingress_rule_mongodb.rule0"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{"target.url"},
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyIngressRuleMongoConfig(
					updateAppName,
					"enabled",
					updateMongoURL,
					"melon",
					"melon",
					"off",
					"off",
					"us-east-1-A",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "status", "enabled"),
					resource.TestCheckResourceAttrSet("ably_ingress_rule_mongodb.rule0", "target.url"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.database", "melon"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.collection", "melon"),
					resource.TestCheckResourceAttrSet("ably_ingress_rule_mongodb.rule0", "target.pipeline"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.full_document", "off"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.full_document_before_change", "off"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.primary_site", "us-east-1-A"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccAblyIngressRuleMongo_Minimal(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	config := minimalIngressRuleConfig(appName, "ably_ingress_rule_mongodb", `target = {
		url                          = "mongodb://user:pass@example.com:27017"
		database                     = "testdb"
		collection                   = "testcol"
		pipeline                     = "[]"
		full_document                = "updateLookup"
		full_document_before_change  = "off"
		primary_site                 = "us-east-1-A"
	}`)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ably_ingress_rule_mongodb.rule0", "id"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "status", "enabled"),
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
func testAccAblyIngressRuleMongoConfig(
	appName string,
	ruleStatus string,
	targetURL string,
	targetDatabase string,
	targetCollection string,
	targetFullDocument string,
	targetFullDocumentBeforeChange string,
	targetPrimarySite string,
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

resource "ably_ingress_rule_mongodb" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q

	target = {
		url = %[3]q
		database = %[4]q
		collection = %[5]q
		pipeline = jsonencode([
		{
		"$set" = {
			"_ablyChannel" = "myChannel"
		}
		}
	])
		full_document = %[6]q
		full_document_before_change = %[7]q
		primary_site = %[8]q

	}
  }
`, appName, ruleStatus, targetURL, targetDatabase, targetCollection, targetFullDocument, targetFullDocumentBeforeChange, targetPrimarySite)
}
