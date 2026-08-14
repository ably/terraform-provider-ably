// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"ably": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ABLY_ACCOUNT_TOKEN"); v == "" {
		t.Fatal("ABLY_ACCOUNT_TOKEN must be set for acceptance tests")
	}
}

// tfProvider is the shared Terraform provider configuration block used by
// minimal-config acceptance tests across resource test files.
const tfProvider = `
terraform {
	required_providers {
		ably = {
			source = "registry.terraform.io/ably/ably"
		}
	}
}
provider "ably" {}
`

// minimalRuleConfig builds an HCL config with only required fields for a rule
// resource. The targetHCL argument is the target block content specific to
// each rule type.
func minimalRuleConfig(appName, resourceType, targetHCL string) string {
	return fmt.Sprintf(`%s
resource "ably_app" "app0" {
	name = %q
}

resource %q "rule0" {
	app_id = ably_app.app0.id

	source = {
		type = "channel.message"
	}

	%s
}
`, tfProvider, appName, resourceType, targetHCL)
}

// minimalIngressRuleConfig builds an HCL config with only required fields for
// an ingress rule resource.
func minimalIngressRuleConfig(appName, resourceType, targetHCL string) string {
	return fmt.Sprintf(`%s
resource "ably_app" "app0" {
	name = %q
}

resource %q "rule0" {
	app_id = ably_app.app0.id

	%s
}
`, tfProvider, appName, resourceType, targetHCL)
}
