// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyQueue(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	queueName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyQueueConfig(appName, control.NewQueue{
					Name:      queueName,
					Ttl:       44,
					MaxLength: 83,
					Region:    control.EuWest1A,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_queue.queue0", "name", queueName),
					resource.TestCheckResourceAttr("ably_queue.queue0", "ttl", "44"),
					resource.TestCheckResourceAttr("ably_queue.queue0", "max_length", "83"),
				),
			},
			{

				Config: testAccAblyQueueConfig(appName, control.NewQueue{
					Name:      queueName + "new",
					Ttl:       30,
					MaxLength: 83,
					Region:    control.UsEast1A,
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
func testAccAblyQueueConfig(appName string, queue control.NewQueue) string {
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

resource "ably_queue" "queue0" {
  app_id            = ably_app.app0.id
  name              = %[2]q
  ttl               = %[3]d
  max_length        = %[4]d
  region            = %[5]q
}

`, appName, queue.Name, queue.Ttl, queue.MaxLength, string(queue.Region))
}
