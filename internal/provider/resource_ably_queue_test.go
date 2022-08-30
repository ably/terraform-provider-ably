package ably_control

import (
	"fmt"
	"testing"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAblyQueue(t *testing.T) {
	app_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	queue_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyQueueConfig(app_name, ably_control_go.NewQueue{
					Name:      queue_name,
					Ttl:       44,
					MaxLength: 83,
					Region:    ably_control_go.EuWest1A,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_queue.queue0", "name", queue_name),
					resource.TestCheckResourceAttr("ably_queue.queue0", "ttl", "44"),
					resource.TestCheckResourceAttr("ably_queue.queue0", "max_length", "83"),
				),
			},
			{

				Config: testAccAblyQueueConfig(app_name, ably_control_go.NewQueue{
					Name:      queue_name + "new",
					Ttl:       30,
					MaxLength: 83,
					Region:    ably_control_go.UsEast1A,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_queue.queue0", "name", queue_name+"new"),
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
func testAccAblyQueueConfig(appName string, queue ably_control_go.NewQueue) string {
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
