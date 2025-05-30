// Package provider implements the Ably provider for Terraform
package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyIngressRulePostgresOutbox(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	updateAppName := "acc-test-" + appName
	testPostgresURL := "postgres://test:test@test.com:5432/your-database-name"
	testUpdatePostgresURL := "postgres://test:test@example.com:5432/your-database-name"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyIngressRulePostgresOutboxConfig(
					appName,
					"enabled",
					testPostgresURL,
					"public",
					"outbox",
					"public",
					"nodes",
					"prefer",
					"-----BEGIN CERTIFICATE----- MIIFiTCCA3GgAwIBAgIUYO1Lomxzj7VRawWwEFiQht9OLpUwDQYJKoZIhvcNAQEL BQAwTDELMAkGA1UEBhMCVVMxETAPBgNVBAgMCE1pY2hpZ2FuMQ8wDQYDVQQHDAZX ...snip... TOfReTlUQzgpXRW5h3n2LVXbXQhPGcVitb88Cm2R8cxQwgB1VncM8yvmKhREo2tz 7Y+sUx6eIl4dlNl9kVrH1TD3EwwtGsjUNlFSZhg= -----END CERTIFICATE-----",
					"us-east-1-A",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.url", testPostgresURL),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.outbox_table_schema", "public"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.outbox_table_name", "outbox"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.nodes_table_schema", "public"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.nodes_table_name", "nodes"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.ssl_mode", "prefer"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.ssl_root_cert", "-----BEGIN CERTIFICATE----- MIIFiTCCA3GgAwIBAgIUYO1Lomxzj7VRawWwEFiQht9OLpUwDQYJKoZIhvcNAQEL BQAwTDELMAkGA1UEBhMCVVMxETAPBgNVBAgMCE1pY2hpZ2FuMQ8wDQYDVQQHDAZX ...snip... TOfReTlUQzgpXRW5h3n2LVXbXQhPGcVitb88Cm2R8cxQwgB1VncM8yvmKhREo2tz 7Y+sUx6eIl4dlNl9kVrH1TD3EwwtGsjUNlFSZhg= -----END CERTIFICATE-----"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.primary_site", "us-east-1-A"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyIngressRulePostgresOutboxConfig(
					updateAppName,
					"enabled",
					testUpdatePostgresURL,
					"public1",
					"outbox1",
					"public1",
					"nodes1",
					"verify-ca",
					"-----BEGIN CERTIFICATE----- MIIFiTCCA3GgAwIBAgIUYO1Lomxzj7VRawWwEFiQht9OLpUwDQYJKoZIhvcNAQEL BQAwTDELMAkGA1UEBhMCVVMxETAPBgNVBAgMCE1pY2hpZ2FuMQ8wDQYDVQQHDAZX ...snip... TOfReTlUQzgpXRW5h3n2LVXbXQhPGcVitb88Cm2R8cxQwgB1VncM8yvmKhREo2tz 7Y+sUx6eIl4dlNl9kVrH1TD3EwwtGsjUNlFSZhg= -----END CERTIFICATE-----",
					"us-east-1-A",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", updateAppName),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.url", testUpdatePostgresURL),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.outbox_table_schema", "public1"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.outbox_table_name", "outbox1"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.nodes_table_schema", "public1"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.nodes_table_name", "nodes1"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.ssl_mode", "verify-ca"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.ssl_root_cert", "-----BEGIN CERTIFICATE----- MIIFiTCCA3GgAwIBAgIUYO1Lomxzj7VRawWwEFiQht9OLpUwDQYJKoZIhvcNAQEL BQAwTDELMAkGA1UEBhMCVVMxETAPBgNVBAgMCE1pY2hpZ2FuMQ8wDQYDVQQHDAZX ...snip... TOfReTlUQzgpXRW5h3n2LVXbXQhPGcVitb88Cm2R8cxQwgB1VncM8yvmKhREo2tz 7Y+sUx6eIl4dlNl9kVrH1TD3EwwtGsjUNlFSZhg= -----END CERTIFICATE-----"),
					resource.TestCheckResourceAttr("ably_ingress_rule_postgres_outbox.rule0", "target.primary_site", "us-east-1-A"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
func testAccAblyIngressRulePostgresOutboxConfig(
	appName string,
	ruleStatus string,
	url string,
	outboxTableSchema string,
	outboxTableName string,
	nodesTableSchema string,
	nodesTableName string,
	sslMode string,
	sslRootCert string,
	primarySite string,
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

resource "ably_ingress_rule_postgres_outbox" "rule0" {
	app_id = ably_app.app0.id
	status = %[2]q

	target = {
		url = %[3]q
		outbox_table_schema = %[4]q
		outbox_table_name = %[5]q
		nodes_table_schema = %[6]q
		nodes_table_name = %[7]q
		ssl_mode = %[8]q
		ssl_root_cert = %[9]q
		primary_site = %[10]q
	}
  }
`, appName, ruleStatus, url, outboxTableSchema, outboxTableName, nodesTableSchema, nodesTableName, sslMode, sslRootCert, primarySite)
}
