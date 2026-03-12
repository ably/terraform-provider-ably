// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	control "github.com/ably/terraform-provider-ably/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccAblyE2ELifecycle exercises a realistic workflow where multiple
// resource types are created together and reference each other:
//
//	Stage 1: Create full infrastructure (app, keys, namespaces, queue, HTTP rule, AMQP rule)
//	Stage 2: Import all resources and verify no plan diff
//	Stage 3: Update multiple resources in place
//	Stage 4: Implicit destroy with CheckDestroy verification
func TestAccAblyE2ELifecycle(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	suffix := acctest.RandStringFromCharSet(12, acctest.CharSetAlphaNum)
	appName := "e2e-" + suffix
	updatedAppName := "e2e-upd-" + suffix
	pubKeyName := "pub-key-" + suffix
	subKeyName := "sub-key-" + suffix
	updatedPubKeyName := "pub-key-upd-" + suffix
	nsPersisted := "nspersist" + suffix
	nsBatching := "nsbatch" + suffix
	queueName := "queue" + suffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckE2EDestroy,
		Steps: []resource.TestStep{
			// -------------------------------------------------------
			// Stage 1: Create full infrastructure
			// -------------------------------------------------------
			{
				Config: testAccE2EConfig(
					appName,
					pubKeyName,
					subKeyName,
					nsPersisted,
					nsBatching,
					queueName,
					"https://example.com/webhooks",
					true,  // ns1 persisted
					false, // ns2 batching_enabled initially false
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// -- App --
					resource.TestCheckResourceAttr("ably_app.e2e_app", "name", appName),
					resource.TestCheckResourceAttr("ably_app.e2e_app", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_app.e2e_app", "tls_only", "true"),
					resource.TestCheckResourceAttrSet("ably_app.e2e_app", "id"),
					resource.TestCheckResourceAttrSet("ably_app.e2e_app", "account_id"),

					// -- Publish key --
					resource.TestCheckResourceAttr("ably_api_key.pub_key", "name", pubKeyName),
					resource.TestCheckResourceAttr("ably_api_key.pub_key", "revocable_tokens", "true"),
					resource.TestCheckTypeSetElemAttr("ably_api_key.pub_key", "capabilities.channel-pub.*", "publish"),
					resource.TestCheckResourceAttrSet("ably_api_key.pub_key", "id"),
					resource.TestCheckResourceAttrSet("ably_api_key.pub_key", "app_id"),
					resource.TestCheckResourceAttrSet("ably_api_key.pub_key", "key"),

					// -- Subscribe key --
					resource.TestCheckResourceAttr("ably_api_key.sub_key", "name", subKeyName),
					resource.TestCheckResourceAttr("ably_api_key.sub_key", "revocable_tokens", "false"),
					resource.TestCheckTypeSetElemAttr("ably_api_key.sub_key", "capabilities.channel-sub.*", "subscribe"),
					resource.TestCheckTypeSetElemAttr("ably_api_key.sub_key", "capabilities.channel-sub.*", "history"),
					resource.TestCheckResourceAttrSet("ably_api_key.sub_key", "id"),
					resource.TestCheckResourceAttrSet("ably_api_key.sub_key", "key"),

					// -- Namespace with persistence --
					resource.TestCheckResourceAttr("ably_namespace.ns_persisted", "id", nsPersisted),
					resource.TestCheckResourceAttr("ably_namespace.ns_persisted", "persisted", "true"),
					resource.TestCheckResourceAttr("ably_namespace.ns_persisted", "persist_last", "true"),
					resource.TestCheckResourceAttr("ably_namespace.ns_persisted", "authenticated", "true"),
					resource.TestCheckResourceAttr("ably_namespace.ns_persisted", "tls_only", "true"),
					resource.TestCheckResourceAttrSet("ably_namespace.ns_persisted", "app_id"),

					// -- Namespace with batching (initially disabled) --
					resource.TestCheckResourceAttr("ably_namespace.ns_batching", "id", nsBatching),
					resource.TestCheckResourceAttr("ably_namespace.ns_batching", "persisted", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns_batching", "batching_enabled", "false"),
					resource.TestCheckResourceAttrSet("ably_namespace.ns_batching", "app_id"),

					// -- Queue --
					resource.TestCheckResourceAttr("ably_queue.e2e_queue", "name", queueName),
					resource.TestCheckResourceAttr("ably_queue.e2e_queue", "ttl", "60"),
					resource.TestCheckResourceAttr("ably_queue.e2e_queue", "max_length", "10000"),
					resource.TestCheckResourceAttr("ably_queue.e2e_queue", "region", "us-east-1-a"),
					resource.TestCheckResourceAttrSet("ably_queue.e2e_queue", "id"),
					resource.TestCheckResourceAttrSet("ably_queue.e2e_queue", "app_id"),
					resource.TestCheckResourceAttrSet("ably_queue.e2e_queue", "amqp_uri"),
					resource.TestCheckResourceAttrSet("ably_queue.e2e_queue", "amqp_queue_name"),

					// -- HTTP rule (references pub_key for signing) --
					resource.TestCheckResourceAttrSet("ably_rule_http.http_rule", "id"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "source.channel_filter", "^my-channel.*"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "request_mode", "single"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "target.url", "https://example.com/webhooks"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "target.format", "json"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "target.headers.0.name", "X-Custom-Header"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "target.headers.0.value", "custom-value"),
					resource.TestCheckResourceAttrSet("ably_rule_http.http_rule", "target.signing_key_id"),

					// -- AMQP rule (references queue) --
					resource.TestCheckResourceAttrSet("ably_rule_amqp.amqp_rule", "id"),
					resource.TestCheckResourceAttr("ably_rule_amqp.amqp_rule", "status", "enabled"),
					resource.TestCheckResourceAttr("ably_rule_amqp.amqp_rule", "source.channel_filter", "^events.*"),
					resource.TestCheckResourceAttr("ably_rule_amqp.amqp_rule", "source.type", "channel.message"),
					resource.TestCheckResourceAttr("ably_rule_amqp.amqp_rule", "request_mode", "single"),
					resource.TestCheckResourceAttrSet("ably_rule_amqp.amqp_rule", "target.queue_id"),
					resource.TestCheckResourceAttr("ably_rule_amqp.amqp_rule", "target.enveloped", "true"),
					resource.TestCheckResourceAttr("ably_rule_amqp.amqp_rule", "target.format", "json"),

					// -- Cross-resource reference checks --
					// Verify the AMQP rule's queue_id matches the queue's id
					resource.TestCheckResourceAttrPair(
						"ably_rule_amqp.amqp_rule", "target.queue_id",
						"ably_queue.e2e_queue", "id",
					),
					// Verify all child resources share the same app_id
					resource.TestCheckResourceAttrPair(
						"ably_api_key.pub_key", "app_id",
						"ably_app.e2e_app", "id",
					),
					resource.TestCheckResourceAttrPair(
						"ably_api_key.sub_key", "app_id",
						"ably_app.e2e_app", "id",
					),
					resource.TestCheckResourceAttrPair(
						"ably_namespace.ns_persisted", "app_id",
						"ably_app.e2e_app", "id",
					),
					resource.TestCheckResourceAttrPair(
						"ably_namespace.ns_batching", "app_id",
						"ably_app.e2e_app", "id",
					),
					resource.TestCheckResourceAttrPair(
						"ably_queue.e2e_queue", "app_id",
						"ably_app.e2e_app", "id",
					),
					resource.TestCheckResourceAttrPair(
						"ably_rule_http.http_rule", "app_id",
						"ably_app.e2e_app", "id",
					),
					resource.TestCheckResourceAttrPair(
						"ably_rule_amqp.amqp_rule", "app_id",
						"ably_app.e2e_app", "id",
					),
				),
			},

			// -------------------------------------------------------
			// Stage 2: Import all resources
			// -------------------------------------------------------
			// Import app
			{
				ResourceName:      "ably_app.e2e_app",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_app.e2e_app"]
					if !ok {
						return "", fmt.Errorf("resource ably_app.e2e_app not found in state")
					}
					return rs.Primary.ID, nil
				},
			},
			// Import publish key
			{
				ResourceName:      "ably_api_key.pub_key",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_api_key.pub_key"]
					if !ok {
						return "", fmt.Errorf("resource ably_api_key.pub_key not found in state")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},
			// Import subscribe key
			{
				ResourceName:      "ably_api_key.sub_key",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_api_key.sub_key"]
					if !ok {
						return "", fmt.Errorf("resource ably_api_key.sub_key not found in state")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},
			// Import persisted namespace
			{
				ResourceName:      "ably_namespace.ns_persisted",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_namespace.ns_persisted"]
					if !ok {
						return "", fmt.Errorf("resource ably_namespace.ns_persisted not found in state")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},
			// Import batching namespace
			{
				ResourceName:      "ably_namespace.ns_batching",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_namespace.ns_batching"]
					if !ok {
						return "", fmt.Errorf("resource ably_namespace.ns_batching not found in state")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},
			// Import queue
			{
				ResourceName:      "ably_queue.e2e_queue",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_queue.e2e_queue"]
					if !ok {
						return "", fmt.Errorf("resource ably_queue.e2e_queue not found in state")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},
			// Import HTTP rule
			{
				ResourceName:      "ably_rule_http.http_rule",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_http.http_rule"]
					if !ok {
						return "", fmt.Errorf("resource ably_rule_http.http_rule not found in state")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
				ImportStateVerifyIgnore: []string{"target.signing_key_id"},
			},
			// Import AMQP rule
			{
				ResourceName:      "ably_rule_amqp.amqp_rule",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["ably_rule_amqp.amqp_rule"]
					if !ok {
						return "", fmt.Errorf("resource ably_rule_amqp.amqp_rule not found in state")
					}
					return fmt.Sprintf("%s,%s", rs.Primary.Attributes["app_id"], rs.Primary.ID), nil
				},
			},

			// -------------------------------------------------------
			// Stage 3: Update multiple resources
			// -------------------------------------------------------
			{
				Config: testAccE2EConfig(
					updatedAppName,
					updatedPubKeyName,
					subKeyName,
					nsPersisted,
					nsBatching,
					queueName,
					"https://example.com/webhooks/v2",
					false, // ns1 persistence disabled
					true,  // ns2 batching enabled
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// -- Updated app name --
					resource.TestCheckResourceAttr("ably_app.e2e_app", "name", updatedAppName),
					resource.TestCheckResourceAttr("ably_app.e2e_app", "status", "enabled"),

					// -- Updated publish key name --
					resource.TestCheckResourceAttr("ably_api_key.pub_key", "name", updatedPubKeyName),
					// Capabilities changed: now has publish + history
					resource.TestCheckTypeSetElemAttr("ably_api_key.pub_key", "capabilities.channel-pub.*", "publish"),
					resource.TestCheckTypeSetElemAttr("ably_api_key.pub_key", "capabilities.channel-pub.*", "history"),

					// -- Namespace persistence toggled --
					resource.TestCheckResourceAttr("ably_namespace.ns_persisted", "persisted", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns_persisted", "persist_last", "false"),
					resource.TestCheckResourceAttr("ably_namespace.ns_persisted", "authenticated", "true"),

					// -- Namespace batching enabled --
					resource.TestCheckResourceAttr("ably_namespace.ns_batching", "batching_enabled", "true"),
					resource.TestCheckResourceAttr("ably_namespace.ns_batching", "batching_interval", "100"),

					// -- Queue still exists with same attributes --
					resource.TestCheckResourceAttr("ably_queue.e2e_queue", "name", queueName),
					resource.TestCheckResourceAttr("ably_queue.e2e_queue", "ttl", "60"),

					// -- HTTP rule updated URL --
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "target.url", "https://example.com/webhooks/v2"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "request_mode", "batch"),
					resource.TestCheckResourceAttr("ably_rule_http.http_rule", "target.format", "msgpack"),

					// -- AMQP rule still references the queue --
					resource.TestCheckResourceAttrPair(
						"ably_rule_amqp.amqp_rule", "target.queue_id",
						"ably_queue.e2e_queue", "id",
					),
					resource.TestCheckResourceAttr("ably_rule_amqp.amqp_rule", "target.enveloped", "false"),
					resource.TestCheckResourceAttr("ably_rule_amqp.amqp_rule", "target.format", "msgpack"),
				),
			},

			// -------------------------------------------------------
			// Stage 4: Implicit destroy (handled by test framework)
			// CheckDestroy is set at the TestCase level above.
			// -------------------------------------------------------
		},
	})
}

// testAccE2EConfig generates the full Terraform HCL configuration for the e2e test.
// It creates an interconnected set of resources: app, 2 keys, 2 namespaces, 1 queue,
// 1 HTTP rule, and 1 AMQP rule. The parameters control the mutable attributes so
// the same function can produce both the initial and updated configs.
func testAccE2EConfig(
	appName string,
	pubKeyName string,
	subKeyName string,
	nsPersistedName string,
	nsBatchingName string,
	queueName string,
	httpTargetURL string,
	ns1Persisted bool,
	ns2BatchingEnabled bool,
) string {
	// Build the publish key capabilities block.
	// In the updated config, we add "history" alongside "publish".
	pubKeyCaps := `["publish"]`
	httpRequestMode := "single"
	httpFormat := "json"
	httpEnveloped := true
	amqpEnveloped := true
	amqpFormat := "json"
	if !ns1Persisted {
		// This is the "updated" scenario; tweak several things
		pubKeyCaps = `["publish", "history"]`
		httpRequestMode = "batch"
		httpFormat = "msgpack"
		httpEnveloped = false
		amqpEnveloped = false
		amqpFormat = "msgpack"
	}

	// Build batching block for ns2
	ns2BatchingBlock := ""
	if ns2BatchingEnabled {
		ns2BatchingBlock = `
  batching_enabled  = true
  batching_interval = 100`
	}

	return fmt.Sprintf(`
terraform {
	required_providers {
		ably = {
			source = "registry.terraform.io/ably/ably"
		}
	}
}
provider "ably" {}

# --- App ---
resource "ably_app" "e2e_app" {
	name     = %[1]q
	status   = "enabled"
	tls_only = true
}

# --- Publish Key ---
resource "ably_api_key" "pub_key" {
	app_id = ably_app.e2e_app.id
	name   = %[2]q
	capabilities = {
		"channel-pub" = %[3]s
	}
	revocable_tokens = true
}

# --- Subscribe Key ---
resource "ably_api_key" "sub_key" {
	app_id = ably_app.e2e_app.id
	name   = %[4]q
	capabilities = {
		"channel-sub" = ["subscribe", "history"]
	}
	revocable_tokens = false
}

# --- Namespace with persistence ---
resource "ably_namespace" "ns_persisted" {
	app_id            = ably_app.e2e_app.id
	id                = %[5]q
	authenticated     = true
	persisted         = %[7]t
	persist_last      = %[7]t
	push_enabled      = false
	tls_only          = true
	expose_timeserial = false
}

# --- Namespace with batching ---
resource "ably_namespace" "ns_batching" {
	app_id            = ably_app.e2e_app.id
	id                = %[6]q
	authenticated     = false
	persisted         = false
	persist_last      = false
	push_enabled      = false
	tls_only          = false
	expose_timeserial = false%[12]s
}

# --- Queue ---
resource "ably_queue" "e2e_queue" {
	app_id     = ably_app.e2e_app.id
	name       = %[8]q
	ttl        = 60
	max_length = 10000
	region     = "us-east-1-a"
}

# --- HTTP Rule (uses pub_key for signing) ---
resource "ably_rule_http" "http_rule" {
	app_id = ably_app.e2e_app.id
	status = "enabled"
	source = {
		channel_filter = "^my-channel.*"
		type           = "channel.message"
	}
	request_mode = %[9]q
	target = {
		url            = %[10]q
		signing_key_id = ably_api_key.pub_key.id
		format         = %[11]q
		enveloped      = %[13]t
		headers = [
			{
				name  = "X-Custom-Header"
				value = "custom-value"
			},
		]
	}
}

# --- AMQP Rule (targets the queue) ---
resource "ably_rule_amqp" "amqp_rule" {
	app_id = ably_app.e2e_app.id
	status = "enabled"
	source = {
		channel_filter = "^events.*"
		type           = "channel.message"
	}
	request_mode = "single"
	target = {
		queue_id  = ably_queue.e2e_queue.id
		enveloped = %[14]t
		format    = %[15]q
		headers = [
			{
				name  = "X-Source"
				value = "terraform-e2e"
			},
		]
	}
}
`,
		appName,          // %[1] app name
		pubKeyName,       // %[2] pub key name
		pubKeyCaps,       // %[3] pub key capabilities
		subKeyName,       // %[4] sub key name
		nsPersistedName,  // %[5] namespace persisted name
		nsBatchingName,   // %[6] namespace batching name
		ns1Persisted,     // %[7] persisted + persist_last bool
		queueName,        // %[8] queue name
		httpRequestMode,  // %[9] HTTP request mode
		httpTargetURL,    // %[10] HTTP target URL
		httpFormat,       // %[11] HTTP format
		ns2BatchingBlock, // %[12] batching block for ns2
		httpEnveloped,    // %[13] HTTP enveloped
		amqpEnveloped,    // %[14] AMQP enveloped
		amqpFormat,       // %[15] AMQP format
	)
}

// testAccCheckE2EDestroy verifies that the app (and by cascade all child resources)
// have been deleted after the test completes.
func testAccCheckE2EDestroy(s *terraform.State) error {
	token := os.Getenv("ABLY_ACCOUNT_TOKEN")
	if token == "" {
		return fmt.Errorf("ABLY_ACCOUNT_TOKEN not set")
	}

	client := control.NewClient(token)
	url := os.Getenv("ABLY_URL")
	if url != "" {
		client.BaseURL = url
	}

	ctx := context.Background()

	// Get account ID via /me
	me, err := client.Me(ctx)
	if err != nil {
		return fmt.Errorf("failed to call /me: %w", err)
	}
	if me.Account == nil || me.Account.ID == "" {
		return fmt.Errorf("could not determine account ID from /me")
	}
	accountID := me.Account.ID

	// Check that the app is gone
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ably_app" {
			continue
		}
		appID := rs.Primary.ID

		apps, err := client.ListApps(ctx, accountID)
		if err != nil {
			return fmt.Errorf("error listing apps during destroy check: %w", err)
		}
		for _, app := range apps {
			if app.ID == appID {
				return fmt.Errorf("ably_app %s still exists after destroy", appID)
			}
		}
	}

	return nil
}
