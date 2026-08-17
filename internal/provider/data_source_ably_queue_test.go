// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAblyQueueDataSource covers looking a queue up by id and by name, and the
// plural list.
//
// The data sources carry the API's own queue shape, so this also checks the
// nested amqp block the ably_queue resource flattens into amqp_uri.
func TestAccAblyQueueDataSource(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	queueName := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyQueueDataSourceConfig(appName, queueName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.ably_queue.by_id", "id", "ably_queue.queue0", "id"),
					resource.TestCheckResourceAttr("data.ably_queue.by_id", "name", queueName),
					resource.TestCheckResourceAttr("data.ably_queue.by_id", "region", "us-east-1-a"),
					resource.TestCheckResourceAttr("data.ably_queue.by_id", "ttl", "60"),
					resource.TestCheckResourceAttrSet("data.ably_queue.by_id", "amqp.uri"),
					resource.TestCheckResourceAttrPair(
						"data.ably_queue.by_name", "id", "ably_queue.queue0", "id"),
					resource.TestCheckResourceAttrSet("data.ably_queues.all", "queues.#"),
				),
			},
		},
	})
}

func testAccAblyQueueDataSourceConfig(appName, queueName string) string {
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

resource "ably_queue" "queue0" {
	app_id     = ably_app.app0.id
	name       = %[2]q
	ttl        = 60
	max_length = 10000
	region     = "us-east-1-a"
}

data "ably_queue" "by_id" {
	app_id = ably_app.app0.id
	id     = ably_queue.queue0.id
}

data "ably_queue" "by_name" {
	app_id = ably_app.app0.id
	name   = ably_queue.queue0.name
}

data "ably_queues" "all" {
	app_id     = ably_app.app0.id
	depends_on = [ably_queue.queue0]
}
`, appName, queueName)
}
