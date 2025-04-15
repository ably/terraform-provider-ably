package ably_control

import (
	"fmt"
	"testing"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAblyNamespace(t *testing.T) {
	app_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	namespace_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyNamespaceConfig(app_name, control.Namespace{
					ID:               namespace_name,
					Authenticated:    true,
					Persisted:        true,
					PersistLast:      true,
					PushEnabled:      true,
					TlsOnly:          true,
					ExposeTimeserial: true,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespace_name),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "authenticated", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persisted", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persist_last", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "push_enabled", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "tls_only", "true"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "expose_timeserial", "true"),
				),
			},
			// Update and Read testing of ably_app.app0
			{
				Config: testAccAblyNamespaceConfig(app_name, control.Namespace{
					ID:               namespace_name,
					Authenticated:    false,
					Persisted:        false,
					PersistLast:      false,
					PushEnabled:      false,
					TlsOnly:          false,
					ExposeTimeserial: false,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespace_name),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "authenticated", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persisted", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persist_last", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "push_enabled", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "tls_only", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "expose_timeserial", "false"),
				),
			},
			{
				Config: testAccAblyNamespaceConfig(app_name, control.Namespace{
					ID:               namespace_name + "new",
					Authenticated:    false,
					Persisted:        false,
					PersistLast:      false,
					PushEnabled:      false,
					TlsOnly:          false,
					ExposeTimeserial: false,
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespace_name+"new"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "authenticated", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persisted", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "persist_last", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "push_enabled", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "tls_only", "false"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "expose_timeserial", "false"),
				),
			},
			{
				Config: testAccAblyNamespaceBatchingConfig(app_name, control.Namespace{
					ID:               namespace_name + "batching",
					Authenticated:    false,
					Persisted:        false,
					PersistLast:      false,
					PushEnabled:      false,
					TlsOnly:          false,
					ExposeTimeserial: false,
					BatchingEnabled:  true,
					BatchingInterval: control.Interval(100),
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespace_name+"batching"),
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
				Config: testAccAblyNamespaceConflationConfig(app_name, control.Namespace{
					ID:                 namespace_name + "conflation",
					Authenticated:      false,
					Persisted:          false,
					PersistLast:        false,
					PushEnabled:        false,
					TlsOnly:            false,
					ExposeTimeserial:   false,
					ConflationEnabled:  true,
					ConflationInterval: control.Interval(1000),
					ConflationKey:      "test",
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("ably_app.app0", "name", app_name),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "id", namespace_name+"conflation"),
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
func testAccAblyNamespaceConfig(appName string, namespace control.Namespace) string {
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

resource "ably_namespace" "namespace0" {
  app_id            = ably_app.app0.id
  id                = %[2]q
  authenticated     = %[3]t
  persisted         = %[4]t
  persist_last      = %[5]t
  push_enabled      = %[6]t
  tls_only          = %[7]t
  expose_timeserial = %[8]t
}

`,
		appName,
		namespace.ID,
		namespace.Authenticated,
		namespace.Persisted,
		namespace.PersistLast,
		namespace.PushEnabled,
		namespace.TlsOnly,
		namespace.ExposeTimeserial,
	)
}

func testAccAblyNamespaceBatchingConfig(appName string, namespace control.Namespace) string {
	return fmt.Sprintf(`
terraform {
	required_providers {
		ably = {
			source =  "github.com/ably/ably"
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
		namespace.TlsOnly,
		namespace.ExposeTimeserial,
		namespace.BatchingEnabled,
		*namespace.BatchingInterval,
	)
}

func testAccAblyNamespaceConflationConfig(appName string, namespace control.Namespace) string {
	return fmt.Sprintf(`
terraform {
	required_providers {
		ably = {
			source =  "github.com/ably/ably"
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
		namespace.TlsOnly,
		namespace.ExposeTimeserial,
		namespace.ConflationEnabled,
		*namespace.ConflationInterval,
		namespace.ConflationKey,
	)
}
