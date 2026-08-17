// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAblyAPIKeyDataSource covers looking a key up by id and by name, and the
// plural list, against a key the same config creates.
func TestAccAblyAPIKeyDataSource(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyKeyDataSourceConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.ably_api_key.by_id", "id", "ably_api_key.key0", "id"),
					resource.TestCheckResourceAttr("data.ably_api_key.by_id", "name", "key0"),
					resource.TestCheckResourceAttr("data.ably_api_key.by_id", "capabilities.channel1.#", "2"),
					resource.TestCheckResourceAttrPair(
						"data.ably_api_key.by_name", "id", "ably_api_key.key0", "id"),
					resource.TestCheckResourceAttrSet("data.ably_api_keys.all", "keys.#"),
				),
			},
		},
	})
}

// TestAccAblyAPIKeyDataSourceAmbiguousName is the regression test for the deliberate
// refusal to guess. Key names are not unique, so two keys sharing a name must be
// an error naming both ids, not an arbitrary pick that changes with API ordering.
func TestAccAblyAPIKeyDataSourceAmbiguousName(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAblyKeyDataSourceDuplicateConfig(appName),
				ExpectError: regexp.MustCompile(`More than one ably_api_key named "twin"`),
			},
		},
	})
}

func testAccAblyKeyDataSourceConfig(appName string) string {
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

resource "ably_api_key" "key0" {
	app_id = ably_app.app0.id
	name   = "key0"
	capabilities = {
		"channel1" = ["publish", "subscribe"],
	}
}

data "ably_api_key" "by_id" {
	app_id = ably_app.app0.id
	id     = ably_api_key.key0.id
}

data "ably_api_key" "by_name" {
	app_id = ably_app.app0.id
	name   = ably_api_key.key0.name
}

data "ably_api_keys" "all" {
	app_id     = ably_app.app0.id
	depends_on = [ably_api_key.key0]
}
`, appName)
}

func testAccAblyKeyDataSourceDuplicateConfig(appName string) string {
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

resource "ably_api_key" "twin_a" {
	app_id = ably_app.app0.id
	name   = "twin"
	capabilities = {
		"channel1" = ["publish"],
	}
}

resource "ably_api_key" "twin_b" {
	app_id = ably_app.app0.id
	name   = "twin"
	capabilities = {
		"channel2" = ["subscribe"],
	}
}

data "ably_api_key" "ambiguous" {
	app_id     = ably_app.app0.id
	name       = "twin"
	depends_on = [ably_api_key.twin_a, ably_api_key.twin_b]
}
`, appName)
}
