package ably_control

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAblyIngressRuleMongo(t *testing.T) {
	app_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	update_app_name := "acc-test-" + app_name
	update_mongo_url := "mongodb://me:lon@honeydew.io:27017"
	test_mongo_url := "mongodb://coco:nut@coco.io:27017"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyIngressRuleMongoConfig(
					app_name,
					"enabled",
					test_mongo_url,
					"coconut",
					"coconut",
					"off",
					"off",
					"us-east-1-A",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.url", test_mongo_url),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.collection", "coconut"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.database", "coconut"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.full_document", "off"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.full_document_before_change", "off"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.primary_site", "us-east-1-A"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyIngressRuleMongoConfig(
					update_app_name,
					"enabled",
					update_mongo_url,
					"melon",
					"melon",
					"off",
					"off",
					"us-east-1-A",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", update_app_name),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.url", update_mongo_url),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.collection", "melon"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.database", "melon"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.full_document", "off"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.full_document_before_change", "off"),
					resource.TestCheckResourceAttr("ably_ingress_rule_mongodb.rule0", "target.primary_site", "us-east-1-A"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyIngressRuleMongoConfig(
	appName string,
	ruleStatus string,
	targetUrl string,
	targetCollection string,
	targetDatabase string,
	targetFullDocument string,
	targetFullDocumentBeforeChange string,
	targetPrimarySite string,
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
`, appName, ruleStatus, targetUrl, targetDatabase, targetCollection, targetFullDocument, targetFullDocumentBeforeChange, targetPrimarySite)
}
