// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAblyQueue(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	queueName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyQueueConfig(appName, control.Queue{
					Name:      queueName,
					TTL:       44,
					MaxLength: 83,
					Region:    "eu-west-1-a",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_queue.queue0", "name", queueName),
					resource.TestCheckResourceAttr("ably_queue.queue0", "ttl", "44"),
					resource.TestCheckResourceAttr("ably_queue.queue0", "max_length", "83"),
					resource.TestCheckResourceAttr("ably_queue.queue0", "region", "eu-west-1-a"),
					resource.TestCheckResourceAttrSet("ably_queue.queue0", "id"),
					resource.TestCheckResourceAttrSet("ably_queue.queue0", "app_id"),
					resource.TestCheckResourceAttrSet("ably_queue.queue0", "amqp_uri"),
					resource.TestCheckResourceAttrSet("ably_queue.queue0", "amqp_queue_name"),
					resource.TestCheckResourceAttrSet("ably_queue.queue0", "stomp_uri"),
					resource.TestCheckResourceAttrSet("ably_queue.queue0", "stomp_host"),
					resource.TestCheckResourceAttrSet("ably_queue.queue0", "stomp_destination"),
					resource.TestCheckResourceAttrSet("ably_queue.queue0", "state"),
				),
			},
			// ImportState testing of ably_queue.queue0
			{
				ResourceName:      "ably_queue.queue0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_queue.queue0"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},
			{

				Config: testAccAblyQueueConfig(appName, control.Queue{
					Name:      queueName + "new",
					TTL:       30,
					MaxLength: 83,
					Region:    "us-east-1-a",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_queue.queue0", "name", queueName+"new"),
					resource.TestCheckResourceAttr("ably_queue.queue0", "ttl", "30"),
					resource.TestCheckResourceAttr("ably_queue.queue0", "max_length", "83"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
// Takes App name, status and tls_only status as function params.
func testAccAblyQueueConfig(appName string, queue control.Queue) string {
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

resource "ably_queue" "queue0" {
  app_id            = ably_app.app0.id
  name              = %[2]q
  ttl               = %[3]d
  max_length        = %[4]d
  region            = %[5]q
}

`, appName, queue.Name, queue.TTL, queue.MaxLength, queue.Region)
}

func TestAccAblyQueue_InvalidTTL(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
terraform {
	required_providers {
		ably = {
			source = "registry.terraform.io/ably/ably"
		}
	}
}
provider "ably" {}
resource "ably_app" "app0" { name = %q }
resource "ably_queue" "queue0" {
	app_id     = ably_app.app0.id
	name       = "test-queue"
	ttl        = 0
	max_length = 100
	region     = "us-east-1-a"
}
`, appName),
				ExpectError: regexp.MustCompile(`.*must be at least 1.*`),
			},
		},
	})
}

func TestAccAblyQueue_InvalidRegion(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
terraform {
	required_providers {
		ably = {
			source = "registry.terraform.io/ably/ably"
		}
	}
}
provider "ably" {}
resource "ably_app" "app0" { name = %q }
resource "ably_queue" "queue0" {
	app_id     = ably_app.app0.id
	name       = "test-queue"
	ttl        = 60
	max_length = 100
	region     = "invalid-region"
}
`, appName),
				ExpectError: regexp.MustCompile(`.*value must be one of.*`),
			},
		},
	})
}
