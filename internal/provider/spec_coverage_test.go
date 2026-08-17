// Package provider implements the Ably provider for Terraform
package provider

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"gopkg.in/yaml.v3"
)

// The tests in this file answer one question: does the provider still cover the
// Control API? They compare the API's own OpenAPI spec against what the provider
// implements, and fail with the list of what is missing.
//
// They read the vendored spec by default, so they are hermetic and run in the
// normal `make test` loop. Point ABLY_SPEC_PATH at a freshly-fetched upstream
// spec to ask the same question of the API as it is today, which is what the
// scheduled spec-drift workflow does. That is the difference between "we vendored
// new surface and haven't implemented it" and "the API grew and we haven't
// noticed".

// specPathEnv overrides the spec these tests read.
const specPathEnv = "ABLY_SPEC_PATH"

// loadSpec reads the OpenAPI spec under test.
func loadSpec(t *testing.T) map[string]any {
	t.Helper()

	path := os.Getenv(specPathEnv)
	if path == "" {
		path = filepath.Join("..", "..", "codegen", "control-api.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read the Control API spec at %s: %s", path, err)
	}

	var doc map[string]any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("could not parse the Control API spec at %s: %s", path, err)
	}
	return doc
}

// TestSpecRuleTypeCoverage fails when the spec knows about a rule type the
// provider does not, or when the provider claims one the spec has never heard
// of. The spec's rule_post schema is a oneOf with a ruleType discriminator, and
// its mapping is the authoritative list of rule types.
func TestSpecRuleTypeCoverage(t *testing.T) {
	t.Parallel()

	specTypes := specRuleTypes(t)
	if len(specTypes) == 0 {
		t.Fatal("found no rule types in the spec; the rule_post discriminator has moved and this test needs updating")
	}

	var missing, unknown []string
	for _, ruleType := range specTypes {
		if _, ok := RuleTypeResources[ruleType]; !ok {
			missing = append(missing, ruleType)
		}
	}
	inSpec := make(map[string]bool, len(specTypes))
	for _, ruleType := range specTypes {
		inSpec[ruleType] = true
	}
	for ruleType := range RuleTypeResources {
		if !inSpec[ruleType] {
			unknown = append(unknown, ruleType)
		}
	}
	sort.Strings(unknown)

	if len(missing) > 0 {
		t.Errorf("the Control API has %d rule type(s) the provider does not implement: %v\n"+
			"Each needs a resource (see DEVELOPMENT.md \"Adding a new integration rule\") and an entry in RuleTypeResources.",
			len(missing), missing)
	}
	if len(unknown) > 0 {
		t.Errorf("RuleTypeResources names %d rule type(s) that are not in the spec: %v\n"+
			"Either the spec renamed them or the mapping has a typo; a wrong discriminator is rejected by the API at apply time.",
			len(unknown), unknown)
	}
}

// specRuleTypes returns every ruleType in the spec's rule_post discriminator.
func specRuleTypes(t *testing.T) []string {
	t.Helper()

	doc := loadSpec(t)
	components, _ := doc["components"].(map[string]any)
	schemas, _ := components["schemas"].(map[string]any)
	rulePost, _ := schemas["rule_post"].(map[string]any)
	discriminator, _ := rulePost["discriminator"].(map[string]any)
	mapping, _ := discriminator["mapping"].(map[string]any)

	types := make([]string, 0, len(mapping))
	for ruleType := range mapping {
		types = append(types, ruleType)
	}
	sort.Strings(types)
	return types
}

// specOperations is every operation in the Control API spec, and what the
// provider does with it. It exists so that new API surface cannot land unnoticed:
// TestSpecOperationCoverage fails on any operation that is not listed here, which
// turns "the API grew" into a decision someone has to make and write down rather
// than something we find out about from a support ticket.
//
// Adding an entry is the deliberate act of recording that decision. Say what the
// provider does, or why it deliberately does not.
var specOperations = map[string]string{
	// Apps.
	"get /accounts/{account_id}/apps":  "ably_app read, plus the ably_app and ably_apps data sources (the API has no GET by app ID, so all three list and filter)",
	"post /accounts/{account_id}/apps": "ably_app create",
	"patch /apps/{id}":                 "ably_app update",
	"delete /apps/{id}":                "ably_app delete",
	"post /apps/{id}/pkcs12": "NOT IMPLEMENTED: uploads APNs config from a binary .p12 archive. " +
		"control.UpdateAppPKCS12 exists but nothing calls it; ably_app takes PEM apns_certificate/apns_private_key instead. " +
		"Terraform has no natural home for a binary upload side effect, so this stays out until someone needs it.",

	// Keys.
	"get /apps/{app_id}/keys":                  "ably_api_key read, plus the ably_api_key and ably_api_keys data sources (no GET by key ID; all three list and filter)",
	"post /apps/{app_id}/keys":                 "ably_api_key create",
	"patch /apps/{app_id}/keys/{key_id}":       "ably_api_key update",
	"post /apps/{app_id}/keys/{key_id}/revoke": "ably_api_key delete (keys are revoked, not deleted)",

	// Namespaces.
	"get /apps/{app_id}/namespaces":                   "ably_namespace read, plus the ably_namespace and ably_namespaces data sources (no GET by namespace ID; all three list and filter)",
	"post /apps/{app_id}/namespaces":                  "ably_namespace create",
	"patch /apps/{app_id}/namespaces/{namespace_id}":  "ably_namespace update",
	"delete /apps/{app_id}/namespaces/{namespace_id}": "ably_namespace delete",

	// Queues.
	"get /apps/{app_id}/queues":               "ably_queue read, plus the ably_queue and ably_queues data sources (no GET by queue ID; all three list and filter)",
	"post /apps/{app_id}/queues":              "ably_queue create",
	"delete /apps/{app_id}/queues/{queue_id}": "ably_queue delete (the API has no queue update, hence RequiresReplace on every attribute)",

	// Rules. Which rule types are covered is TestSpecRuleTypeCoverage's job.
	"get /apps/{app_id}/rules":              "rule resources read (also how the exporter walks an account)",
	"get /apps/{app_id}/rules/{rule_id}":    "rule resources read",
	"post /apps/{app_id}/rules":             "rule resources create",
	"patch /apps/{app_id}/rules/{rule_id}":  "rule resources update",
	"delete /apps/{app_id}/rules/{rule_id}": "rule resources delete",

	// Read-only surface.
	"get /me": "ably_me data source",
	"get /apps/{id}/stats": "NOT IMPLEMENTED: app statistics. Read-only time series whose entries map has no fixed shape, " +
		"so it does not map cleanly onto framework types. Deferred deliberately.",
	"get /accounts/{id}/stats": "NOT IMPLEMENTED: account statistics. Same shape problem as app stats.",
}

// TestSpecOperationCoverage fails when the spec gains an operation that
// specOperations does not account for, or when specOperations describes one the
// spec no longer has.
func TestSpecOperationCoverage(t *testing.T) {
	t.Parallel()

	operations := specOperationList(t)
	if len(operations) == 0 {
		t.Fatal("found no operations in the spec; its paths section has moved and this test needs updating")
	}

	inSpec := make(map[string]bool, len(operations))
	var undocumented []string
	for _, operation := range operations {
		inSpec[operation] = true
		if _, ok := specOperations[operation]; !ok {
			undocumented = append(undocumented, operation)
		}
	}

	var stale []string
	for operation := range specOperations {
		if !inSpec[operation] {
			stale = append(stale, operation)
		}
	}
	sort.Strings(stale)

	if len(undocumented) > 0 {
		t.Errorf("the Control API has %d operation(s) this provider has not accounted for: %v\n"+
			"Decide what the provider should do with each, then record it in specOperations in this file.",
			len(undocumented), undocumented)
	}
	if len(stale) > 0 {
		t.Errorf("specOperations describes %d operation(s) the spec no longer has: %v\n"+
			"The API changed shape; drop or update those entries, and check whether the resources relying on them still work.",
			len(stale), stale)
	}
}

// specOperationList returns every "method path" pair in the spec.
func specOperationList(t *testing.T) []string {
	t.Helper()

	doc := loadSpec(t)
	paths, _ := doc["paths"].(map[string]any)

	methods := map[string]bool{"get": true, "post": true, "patch": true, "put": true, "delete": true}

	var operations []string
	for path, item := range paths {
		operationsByMethod, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for method := range operationsByMethod {
			if methods[method] {
				operations = append(operations, method+" "+path)
			}
		}
	}
	sort.Strings(operations)
	return operations
}
