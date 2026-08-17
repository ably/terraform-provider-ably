// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAblyAppDataSource covers looking an app up by both keys, and the plural
// list, against an app the same config creates.
func TestAccAblyAppDataSource(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAblyAppDataSourceConfig(appName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// By id: every attribute comes from the API, so the data
					// source and the resource must agree.
					resource.TestCheckResourceAttrPair(
						"data.ably_app.by_id", "id", "ably_app.app0", "id"),
					resource.TestCheckResourceAttr("data.ably_app.by_id", "name", appName),
					resource.TestCheckResourceAttr("data.ably_app.by_id", "status", "enabled"),
					resource.TestCheckResourceAttr("data.ably_app.by_id", "tls_only", "true"),
					// By name: same app, found the other way.
					resource.TestCheckResourceAttrPair(
						"data.ably_app.by_name", "id", "ably_app.app0", "id"),
					// account_id defaults to the provider's account rather than
					// having to be passed in.
					resource.TestCheckResourceAttrSet("data.ably_app.by_id", "account_id"),
					// The plural data source lists at least the app we made.
					resource.TestCheckResourceAttrSet("data.ably_apps.all", "apps.#"),
				),
			},
		},
	})
}

// TestAccAblyAppDataSourceLookupErrors covers the two ways a lookup is invalid.
// Both are caught before any API call, so the message has to say which mistake
// was made rather than surfacing an opaque API error.
func TestAccAblyAppDataSourceLookupErrors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccAblyAppDataSourceLookupConfig(`id = "abc123"` + "\n" + `name = "some-app"`),
				ExpectError: regexp.MustCompile(`Ambiguous ably_app lookup`),
			},
			{
				Config:      testAccAblyAppDataSourceLookupConfig(""),
				ExpectError: regexp.MustCompile(`Incomplete ably_app lookup`),
			},
			{
				Config:      testAccAblyAppDataSourceLookupConfig(`id = "does-not-exist"`),
				ExpectError: regexp.MustCompile(`No ably_app with id "does-not-exist"`),
			},
		},
	})
}

func testAccAblyAppDataSourceConfig(appName string) string {
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

data "ably_app" "by_id" {
	id = ably_app.app0.id
}

data "ably_app" "by_name" {
	name = ably_app.app0.name
}

data "ably_apps" "all" {
	depends_on = [ably_app.app0]
}
`, appName)
}

func testAccAblyAppDataSourceLookupConfig(lookupAttrs string) string {
	return fmt.Sprintf(`
terraform {
	required_providers {
		ably = {
			source = "registry.terraform.io/ably/ably"
		}
	}
}
provider "ably" {}

data "ably_app" "lookup" {
	%[1]s
}
`, lookupAttrs)
}
