// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAblyNamespaceDataSource covers the singular lookup and the plural list
// against a namespace the same config creates.
//
// It also pins the identified/authenticated pair: the vendored spec documents only
// the deprecated authenticated flag, so identified is added to the data source
// schema by hand and must report the same value.
func TestAccAblyNamespaceDataSource(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyNamespaceDataSourceConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.ably_namespace.by_id", "id", "chat"),
					resource.TestCheckResourceAttr("data.ably_namespace.by_id", "persisted", "true"),
					resource.TestCheckResourceAttr("data.ably_namespace.by_id", "identified", "true"),
					resource.TestCheckResourceAttr("data.ably_namespace.by_id", "authenticated", "true"),
					resource.TestCheckResourceAttr("data.ably_namespace.by_id", "expose_timeserial", "false"),
					resource.TestCheckResourceAttrSet("data.ably_namespaces.all", "namespaces.#"),
				),
			},
		},
	})
}

func testAccAblyNamespaceDataSourceConfig(appName string) string {
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

resource "ably_namespace" "chat" {
	app_id     = ably_app.app0.id
	id         = "chat"
	persisted  = true
	identified = true
}

data "ably_namespace" "by_id" {
	app_id = ably_app.app0.id
	id     = ably_namespace.chat.id
}

data "ably_namespaces" "all" {
	app_id     = ably_app.app0.id
	depends_on = [ably_namespace.chat]
}
`, appName)
}
