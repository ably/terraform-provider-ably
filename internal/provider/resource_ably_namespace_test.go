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

func TestAccAblyNamespace(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	namespaceName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyNamespaceConfig(appName, control.NamespacePost{
					ID:                      namespaceName,
					Authenticated:           true,
					Persisted:               true,
					PersistLast:             true,
					PushEnabled:             true,
					TLSOnly:                 true,
					ExposeTimeserial:        true,
					MutableMessages:         true,
					PopulateChannelRegistry: true,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespaceName),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "authenticated", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persisted", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persist_last", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "push_enabled", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "tls_only", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "expose_timeserial", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "mutable_messages", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "populate_channel_registry", "true"),
					resource.TestCheckResourceAttrSet("ably_namespace.namespace0", "app_id"),
				),
			},
			// ImportState testing of ably_namespace.namespace0
			{
				ResourceName:      "ably_namespace.namespace0",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_namespace.namespace0"]
					if !ok {
						return "", fmt.Errorf("resource not found")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyNamespaceConfig(appName, control.NamespacePost{
					ID:                      namespaceName,
					Authenticated:           false,
					Persisted:               false,
					PersistLast:             false,
					PushEnabled:             false,
					TLSOnly:                 false,
					ExposeTimeserial:        false,
					MutableMessages:         false,
					PopulateChannelRegistry: false,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespaceName),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "authenticated", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persisted", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persist_last", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "push_enabled", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "tls_only", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "expose_timeserial", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "mutable_messages", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "populate_channel_registry", "false"),
				),
			},
			{
				Config: testAccAblyNamespaceConfig(appName, control.NamespacePost{
					ID:               namespaceName + "new",
					Authenticated:    false,
					Persisted:        false,
					PersistLast:      false,
					PushEnabled:      false,
					TLSOnly:          false,
					ExposeTimeserial: false,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespaceName+"new"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "authenticated", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persisted", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persist_last", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "push_enabled", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "tls_only", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "expose_timeserial", "false"),
				),
			},
			{
				Config: testAccAblyNamespaceBatchingConfig(appName, control.NamespacePost{
					ID:               namespaceName + "batching",
					Authenticated:    false,
					Persisted:        false,
					PersistLast:      false,
					PushEnabled:      false,
					TLSOnly:          false,
					ExposeTimeserial: false,
					BatchingEnabled:  ptr(true),
					BatchingInterval: ptr(100),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespaceName+"batching"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "authenticated", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persisted", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persist_last", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "push_enabled", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "tls_only", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "expose_timeserial", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "batching_enabled", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "batching_interval", "100"),
				),
			},
			{
				Config: testAccAblyNamespaceConflationConfig(appName, control.NamespacePost{
					ID:                 namespaceName + "conflation",
					Authenticated:      false,
					Persisted:          false,
					PersistLast:        false,
					PushEnabled:        false,
					TLSOnly:            false,
					ExposeTimeserial:   false,
					ConflationEnabled:  ptr(true),
					ConflationInterval: ptr(1000),
					ConflationKey:      ptr("test"),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", appName),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespaceName+"conflation"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "authenticated", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persisted", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persist_last", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "push_enabled", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "tls_only", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "expose_timeserial", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "conflation_enabled", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "conflation_interval", "1000"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "conflation_key", "test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
// Takes App name, status and tls_only status as function params.
func testAccAblyNamespaceConfig(appName string, namespace control.NamespacePost) string {
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

resource "ably_namespace" "namespace0" {
  app_id                    = ably_app.app0.id
  id                        = %[2]q
  authenticated             = %[3]t
  persisted                 = %[4]t
  persist_last              = %[5]t
  push_enabled              = %[6]t
  tls_only                  = %[7]t
  expose_timeserial         = %[8]t
  mutable_messages          = %[9]t
  populate_channel_registry = %[10]t
}

`,
		appName,
		namespace.ID,
		namespace.Authenticated,
		namespace.Persisted,
		namespace.PersistLast,
		namespace.PushEnabled,
		namespace.TLSOnly,
		namespace.ExposeTimeserial,
		namespace.MutableMessages,
		namespace.PopulateChannelRegistry,
	)
}

func testAccAblyNamespaceBatchingConfig(appName string, namespace control.NamespacePost) string {
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

resource "ably_namespace" "namespace0" {
  app_id              = ably_app.app0.id
  id                  = %[2]q
  authenticated       = %[3]t
  persisted           = %[4]t
  persist_last        = %[5]t
  push_enabled        = %[6]t
  tls_only            = %[7]t
  expose_timeserial   = %[8]t
  batching_enabled    = %[9]t
  batching_interval   = %[10]d
}

`,
		appName,
		namespace.ID,
		namespace.Authenticated,
		namespace.Persisted,
		namespace.PersistLast,
		namespace.PushEnabled,
		namespace.TLSOnly,
		namespace.ExposeTimeserial,
		*namespace.BatchingEnabled,
		*namespace.BatchingInterval,
	)
}

func testAccAblyNamespaceConflationConfig(appName string, namespace control.NamespacePost) string {
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

resource "ably_namespace" "namespace0" {
  app_id              = ably_app.app0.id
  id                  = %[2]q
  authenticated       = %[3]t
  persisted           = %[4]t
  persist_last        = %[5]t
  push_enabled        = %[6]t
  tls_only            = %[7]t
  expose_timeserial   = %[8]t
  conflation_enabled  = %[9]t
  conflation_interval = %[10]d
  conflation_key      = %[11]q
}

`,
		appName,
		namespace.ID,
		namespace.Authenticated,
		namespace.Persisted,
		namespace.PersistLast,
		namespace.PushEnabled,
		namespace.TLSOnly,
		namespace.ExposeTimeserial,
		*namespace.ConflationEnabled,
		*namespace.ConflationInterval,
		*namespace.ConflationKey,
	)
}

func TestAccAblyNamespace_Minimal(t *testing.T) {
	appName := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	config := fmt.Sprintf(`%s
resource "ably_app" "app0" {
	name = %q
}

resource "ably_namespace" "ns0" {
	app_id = ably_app.app0.id
	id     = "minimal-ns"
}
`, tfProvider, appName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("ably_namespace.ns0", "app_id"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "id", "minimal-ns"),
					// Optional+Computed defaults (all false)
					resource.TestCheckResourceAttr("ably_namespace.ns0", "authenticated", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "persisted", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "persist_last", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "push_enabled", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "tls_only", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "expose_timeserial", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "mutable_messages", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "populate_channel_registry", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "batching_enabled", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns0", "conflation_enabled", "false"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func TestAccAblyNamespace_InvalidBatchingInterval(t *testing.T) {
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
resource "ably_namespace" "ns0" {
	app_id            = ably_app.app0.id
	id                = "test-ns"
	batching_enabled  = true
	batching_interval = -1
}
`, appName),
				ExpectError: regexp.MustCompile(`.*must be at least 0.*`),
			},
		},
	})
}
