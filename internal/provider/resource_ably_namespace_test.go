package provider

import (
	"fmt"
	"testing"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAblyNamespace(t *testing.T) {
	app_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	namespace_name := acctest.RandStringFromCharSet(15, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Create and Read testing of ably_app.app0
			{
				Config: testAccAblyNamespaceConfig(app_name, ably_control_go.Namespace{
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
				Config: testAccAblyNamespaceConfig(app_name, ably_control_go.Namespace{
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
				Config: testAccAblyNamespaceConfig(app_name, ably_control_go.Namespace{
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
				Config: testAccAblyNamespaceBatchingConfig(app_name, ably_control_go.Namespace{
					ID:               namespace_name + "batching",
					Authenticated:    false,
					Persisted:        false,
					PersistLast:      false,
					PushEnabled:      false,
					TlsOnly:          false,
					ExposeTimeserial: false,
					BatchingEnabled:  true,
					BatchingPolicy:   "simple",
					BatchingInterval: ably_control_go.BatchingInterval(100),
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
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "batching_policy", "simple"),
					resource.TestCheckResourceAttr("ably_namespace.namespace0", "batching_interval", "100"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Function with inline HCL to provision an ably_app resource
// Takes App name, status and tls_only status as function params.
func testAccAblyNamespaceConfig(appName string, namespace ably_control_go.Namespace) string {
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

`, appName, namespace.ID, namespace.Authenticated, namespace.Persisted, namespace.PersistLast, namespace.PushEnabled, namespace.TlsOnly, namespace.ExposeTimeserial)
}

func testAccAblyNamespaceBatchingConfig(appName string, namespace ably_control_go.Namespace) string {
	return fmt.Sprintf(`
terraform {
	required_providers {
		ably = {
			source =  "github.com/ably/ably"
			version = "0.4.3"
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
  batching_enabled  = %[9]t
  batching_policy   = "%[10]s"
  batching_interval = %[11]d
}

`, appName, namespace.ID, namespace.Authenticated, namespace.Persisted, namespace.PersistLast, namespace.PushEnabled, namespace.TlsOnly, namespace.ExposeTimeserial, namespace.BatchingEnabled, namespace.BatchingPolicy, *namespace.BatchingInterval)
}
