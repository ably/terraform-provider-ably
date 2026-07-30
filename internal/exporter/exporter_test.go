package exporter

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/ably/terraform-provider-ably/internal/provider"
)

// runExport exports the fake account and returns the generated files by name.
func runExport(t *testing.T, config Config) (*Result, map[string]string) {
	t.Helper()

	fake := newFakeControlAPI(t)
	config.Token = "fake-token"
	config.URL = fake.url()

	result, err := Run(context.Background(), config)
	if err != nil {
		t.Fatalf("Run: %s", err)
	}

	files := map[string]string{}
	for _, file := range result.Files {
		files[file.Name] = string(file.Content)
	}
	return result, files
}

func TestRunExportsEveryResourceInTheAccount(t *testing.T) {
	result, files := runExport(t, Config{Imports: true, ProviderVersion: DefaultProviderVersion})

	// One app, one active key, one namespace, one queue and three supported
	// rules. The fourth rule has a ruleType the provider does not implement.
	want := map[string]int{
		"ably_app":            1,
		"ably_api_key":        1,
		"ably_namespace":      1,
		"ably_queue":          1,
		"ably_rule_http":      1,
		"ably_rule_amqp":      1,
		"ably_rule_bodyguard": 1,
		"ably_rule_kafka":     1,
	}
	if result.Total != 8 {
		t.Errorf("exported %d resources, want 8:\n%s", result.Total, result.Summary())
	}
	for resourceType, count := range want {
		if result.Counts[resourceType] != count {
			t.Errorf("exported %d of %s, want %d", result.Counts[resourceType], resourceType, count)
		}
	}

	for name := range files {
		t.Logf("generated %s", name)
	}
	config, ok := files["app_chat_service.tf"]
	if !ok {
		t.Fatalf("no per-app file was generated, got %v", fileNames(files))
	}

	// Labels come from names, so the app's resources are readable.
	for _, want := range []string{
		`resource "ably_app" "chat_service" {`,
		`resource "ably_api_key" "chat_service_root_key" {`,
		`resource "ably_namespace" "chat_service_chat" {`,
		`resource "ably_queue" "chat_service_orders" {`,
		`resource "ably_rule_http" "chat_service"`,
	} {
		if !strings.Contains(config, want) {
			t.Errorf("generated config is missing %s:\n%s", want, config)
		}
	}
}

func TestRunWritesReferencesRatherThanIDs(t *testing.T) {
	_, files := runExport(t, Config{})
	config := files["app_chat_service.tf"]

	assertAssigns(t, config, "app_id", "ably_app.chat_service.id")
	if strings.Contains(config, `"app1"`) {
		t.Errorf("an app ID was written as a literal:\n%s", config)
	}
	// The AMQP rule targets the exported queue, and the HTTP rule signs with the
	// exported key, so both should point at them.
	assertAssigns(t, config, "queue_id", "ably_queue.chat_service_orders.id")
	assertAssigns(t, config, "signing_key_id", "ably_api_key.chat_service_root_key.id")
	if strings.Contains(config, `"key1"`) {
		t.Errorf("a key ID was written as a literal:\n%s", config)
	}
}

// assertAssigns checks an assignment, tolerating the whitespace hclwrite inserts
// to align equals signs.
func assertAssigns(t *testing.T, config, attribute, value string) {
	t.Helper()
	pattern := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(attribute) + `\s*=\s*` + regexp.QuoteMeta(value) + `\s*$`)
	if !pattern.MatchString(config) {
		t.Errorf("expected %s = %s in:\n%s", attribute, value, config)
	}
}

// TestRunOmitsComputedAttributes checks the invariant that makes the output
// appliable: Terraform rejects config that sets a computed attribute. The list
// comes from the provider's schemas rather than being written out here, so the
// check survives schema changes.
func TestRunOmitsComputedAttributes(t *testing.T) {
	_, files := runExport(t, Config{})
	config := files["app_chat_service.tf"]

	bridge, err := newBridge(context.Background(), "test")
	if err != nil {
		t.Fatalf("newBridge: %s", err)
	}

	// Attribute names are only meaningful within their own resource, so walk
	// the file block by block.
	resourceHeader := regexp.MustCompile(`^resource "([^"]+)" "([^"]+)" \{$`)
	var computedHere map[string]bool
	var currentType string

	for _, line := range strings.Split(config, "\n") {
		if match := resourceHeader.FindStringSubmatch(line); match != nil {
			currentType = match[1]
			schema, err := bridge.schema(currentType)
			if err != nil {
				t.Fatal(err)
			}
			computedHere = map[string]bool{}
			for _, attribute := range computedOnlyAttributes(schema.Block) {
				computedHere[attribute] = true
			}
			continue
		}

		// hclwrite aligns the equals signs, so the name carries trailing space.
		name, _, isAssignment := strings.Cut(strings.TrimSpace(line), "=")
		name = strings.TrimSpace(name)
		if !isAssignment || currentType == "" {
			continue
		}
		if computedHere[name] {
			t.Errorf("generated config sets %s.%s, which the provider computes:\n%s", currentType, name, config)
		}
	}

	// A configurable attribute of the same shape has to survive, otherwise the
	// check above would pass on empty output.
	if !strings.Contains(config, "app_id =") {
		t.Errorf("app_id was dropped:\n%s", config)
	}
	// ably_namespace.id is required, not computed, so it must be written.
	assertAssigns(t, config, "id", `"chat"`)
}

// computedOnlyAttributes returns attributes a user cannot set, at any depth.
func computedOnlyAttributes(block *tfprotov6.SchemaBlock) []string {
	if block == nil {
		return nil
	}
	var names []string
	var walk func(attributes []*tfprotov6.SchemaAttribute)
	walk = func(attributes []*tfprotov6.SchemaAttribute) {
		for _, attribute := range attributes {
			if attribute == nil {
				continue
			}
			if isComputedOnly(attribute) {
				names = append(names, attribute.Name)
			}
			if attribute.NestedType != nil {
				walk(attribute.NestedType.Attributes)
			}
		}
	}
	walk(block.Attributes)
	return names
}

func TestRunRendersNestedAttributes(t *testing.T) {
	_, files := runExport(t, Config{})
	config := files["app_chat_service.tf"]

	for _, want := range []string{"source = {", "target = {", "headers = [", "before_publish_config = {"} {
		if !strings.Contains(config, want) {
			t.Errorf("generated config is missing %q:\n%s", want, config)
		}
	}

	assertAssigns(t, config, "url", `"https://example.com/hook?a=1"`)
	assertAssigns(t, config, "name", `"X-Ably"`)
	assertAssigns(t, config, "retry_timeout", "3000")
	assertAssigns(t, config, `"chat:*"`, `["publish", "subscribe"]`)
}

func TestRunSkipsUnsupportedRuleTypesWithAWarning(t *testing.T) {
	result, _ := runExport(t, Config{})

	var found bool
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "hive/text-moderation") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a warning about the unsupported ruleType, got %v", result.Warnings)
	}
}

func TestRunGeneratesImportBlocks(t *testing.T) {
	_, files := runExport(t, Config{Imports: true})

	imports, ok := files["imports.tf"]
	if !ok {
		t.Fatalf("no imports.tf was generated, got %v", fileNames(files))
	}

	// Account-level resources import by bare ID, app-scoped ones by
	// "appID,id", which is the format the provider's ImportState expects.
	assertAssigns(t, imports, "to", "ably_app.chat_service")
	assertAssigns(t, imports, "id", `"app1"`)
	assertAssigns(t, imports, "to", "ably_api_key.chat_service_root_key")
	assertAssigns(t, imports, "id", `"app1,key1"`)
	assertAssigns(t, imports, "to", "ably_rule_bodyguard.chat_service")
	assertAssigns(t, imports, "id", `"app1,rule3"`)
}

func TestRunWithoutImports(t *testing.T) {
	_, files := runExport(t, Config{Imports: false})
	if _, ok := files["imports.tf"]; ok {
		t.Error("imports.tf was generated despite Imports being false")
	}
}

func TestRunSecretsInline(t *testing.T) {
	result, files := runExport(t, Config{Secrets: SecretsInline})
	config := files["app_chat_service.tf"]

	assertAssigns(t, config, "api_key", `"bodyguard-secret"`)
	if len(result.Sensitive) == 0 {
		t.Error("inline mode did not report which attributes were sensitive")
	}
}

func TestRunSecretsVars(t *testing.T) {
	result, files := runExport(t, Config{Secrets: SecretsVars})
	config := files["app_chat_service.tf"]

	if strings.Contains(config, "bodyguard-secret") {
		t.Errorf("vars mode leaked a sensitive value:\n%s", config)
	}
	assertAssigns(t, config, "api_key", "var.rule_bodyguard_chat_service_target_api_key")

	variables, ok := files["variables.tf"]
	if !ok {
		t.Fatalf("vars mode generated no variables.tf, got %v", fileNames(files))
	}
	if !strings.Contains(variables, `variable "rule_bodyguard_chat_service_target_api_key" {`) {
		t.Errorf("variables.tf is missing the declaration:\n%s", variables)
	}
	if !regexp.MustCompile(`sensitive\s*=\s*true`).MatchString(variables) {
		t.Errorf("variables.tf does not mark the variable sensitive:\n%s", variables)
	}
	if len(result.Sensitive) == 0 {
		t.Error("vars mode did not report which attributes were sensitive")
	}
}

func TestRunSecretsOmit(t *testing.T) {
	_, files := runExport(t, Config{Secrets: SecretsOmit})
	config := files["app_chat_service.tf"]

	if strings.Contains(config, "bodyguard-secret") {
		t.Errorf("omit mode leaked a sensitive value:\n%s", config)
	}
	if !strings.Contains(config, "# api_key omitted: sensitive value") {
		t.Errorf("omit mode did not leave a comment where the value was:\n%s", config)
	}
	if _, ok := files["variables.tf"]; ok {
		t.Error("omit mode generated variables.tf")
	}
}

func TestRunGeneratesParseableHCL(t *testing.T) {
	for _, secrets := range []SecretMode{SecretsInline, SecretsVars, SecretsOmit} {
		t.Run(string(secrets), func(t *testing.T) {
			_, files := runExport(t, Config{Secrets: secrets, Imports: true, ProviderVersion: DefaultProviderVersion})

			for name, content := range files {
				parser := hclparse.NewParser()
				_, diagnostics := parser.ParseHCL([]byte(content), name)
				if diagnostics.HasErrors() {
					t.Errorf("%s is not valid HCL: %s\n%s", name, diagnostics.Error(), content)
				}
			}
		})
	}
}

func TestRunSingleFile(t *testing.T) {
	_, files := runExport(t, Config{SingleFile: true})

	if _, ok := files["main.tf"]; !ok {
		t.Fatalf("single-file mode did not generate main.tf, got %v", fileNames(files))
	}
	for name := range files {
		if strings.HasPrefix(name, "app_") {
			t.Errorf("single-file mode also generated %s", name)
		}
	}
}

func TestRunProviderFile(t *testing.T) {
	_, files := runExport(t, Config{ProviderVersion: "~> 1.0"})

	providerFile, ok := files["provider.tf"]
	if !ok {
		t.Fatalf("no provider.tf was generated, got %v", fileNames(files))
	}
	assertAssigns(t, providerFile, "source", `"ably/ably"`)
	assertAssigns(t, providerFile, "version", `"~> 1.0"`)
	for _, want := range []string{`provider "ably" {`, "ABLY_ACCOUNT_TOKEN"} {
		if !strings.Contains(providerFile, want) {
			t.Errorf("provider.tf is missing %q:\n%s", want, providerFile)
		}
	}
	// A non-default Control API URL has to be carried into the config,
	// otherwise the generated config would point at production.
	if !strings.Contains(providerFile, "url =") {
		t.Errorf("provider.tf does not pin the Control API URL it was exported from:\n%s", providerFile)
	}
}

func TestRunFiltersByApp(t *testing.T) {
	fake := newFakeControlAPI(t)

	_, err := Run(context.Background(), Config{
		Token: "fake-token",
		URL:   fake.url(),
		Apps:  []string{"no-such-app"},
	})
	if err == nil {
		t.Fatal("filtering on an unknown app should fail rather than export nothing")
	}
	if !strings.Contains(err.Error(), "no-such-app") {
		t.Errorf("error should name the filter that matched nothing, got %s", err)
	}

	// Filtering by name, case-insensitively, keeps the app.
	result, err := Run(context.Background(), Config{
		Token: "fake-token",
		URL:   fake.url(),
		Apps:  []string{"chat service"},
	})
	if err != nil {
		t.Fatalf("Run with an app filter: %s", err)
	}
	if result.Counts["ably_app"] != 1 {
		t.Errorf("app filter by name exported %d apps, want 1", result.Counts["ably_app"])
	}
}

func TestRunRequiresAToken(t *testing.T) {
	_, err := Run(context.Background(), Config{})
	if err == nil {
		t.Fatal("Run without a token should fail")
	}
}

func TestRunOnlyReads(t *testing.T) {
	fake := newFakeControlAPI(t)

	if _, err := Run(context.Background(), Config{Token: "fake-token", URL: fake.url()}); err != nil {
		t.Fatalf("Run: %s", err)
	}

	for _, request := range fake.recorded() {
		if !strings.HasPrefix(request, "GET ") {
			t.Errorf("the exporter made a non-read request: %s", request)
		}
	}
}

func TestWriteFilesRefusesToClobber(t *testing.T) {
	result, _ := runExport(t, Config{})
	directory := t.TempDir()

	if err := WriteFiles(directory, result, false); err != nil {
		t.Fatalf("WriteFiles into an empty directory: %s", err)
	}
	if _, err := os.Stat(filepath.Join(directory, "provider.tf")); err != nil {
		t.Fatalf("provider.tf was not written: %s", err)
	}

	err := WriteFiles(directory, result, false)
	if err == nil {
		t.Fatal("WriteFiles overwrote an existing configuration without -force")
	}
	if !strings.Contains(err.Error(), "-force") {
		t.Errorf("error should point at -force, got %s", err)
	}

	if err := WriteFiles(directory, result, true); err != nil {
		t.Fatalf("WriteFiles with force: %s", err)
	}
}

// TestSupportedResourceTypesCoverage keeps the exporter in step with the
// provider. A resource it cannot find is a silent gap in an export, which is worse
// than a failure here.
func TestSupportedResourceTypesCoverage(t *testing.T) {
	supported := map[string]bool{}
	for _, resourceType := range SupportedResourceTypes() {
		supported[resourceType] = true
	}

	for _, resourceType := range provider.ResourceTypeNames(context.Background()) {
		if supported[resourceType] {
			continue
		}
		if provider.IsRuleResourceType(resourceType) {
			t.Errorf("the exporter cannot discover %s.\n"+
				"Add its Control API ruleType to RuleTypeResources in internal/provider/rule_types.go.", resourceType)
			continue
		}
		t.Errorf("the exporter cannot discover %s.\n"+
			"Add a lister for it to discoverApp (or discover, if it is account-level) "+
			"in internal/exporter/discover.go, and name it in SupportedResourceTypes.", resourceType)
	}
}

func TestParseSecretMode(t *testing.T) {
	for _, valid := range []string{"inline", "vars", "omit"} {
		if _, err := ParseSecretMode(valid); err != nil {
			t.Errorf("ParseSecretMode(%q): %s", valid, err)
		}
	}
	if _, err := ParseSecretMode("plaintext"); err == nil {
		t.Error("ParseSecretMode should reject unknown modes")
	}
}

func fileNames(files map[string]string) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}

// TestRunRepairsConflictingAttributes covers what the protocol schema can't
// describe: ably_namespace's identified attribute and its deprecated
// authenticated alias conflict, a read returns both, and writing both is invalid.
// The exporter learns this from the provider and drops the deprecated spelling.
func TestRunRepairsConflictingAttributes(t *testing.T) {
	result, files := runExport(t, Config{})
	config := files["app_chat_service.tf"]

	var repaired bool
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "dropped the deprecated ably_namespace.chat_service_chat.authenticated") {
			repaired = true
		}
	}
	if !repaired {
		t.Errorf("expected the deprecated authenticated attribute to be dropped, warnings were %v", result.Warnings)
	}

	if strings.Contains(config, "authenticated") {
		t.Errorf("the deprecated attribute survived into the config:\n%s", config)
	}
	assertAssigns(t, config, "identified", "false")
}

// TestRunGeneratesConfigTheProviderAccepts checks every resource was validated by
// the provider, so nothing needs a human except values the API withholds.
func TestRunGeneratesConfigTheProviderAccepts(t *testing.T) {
	result, files := runExport(t, Config{Imports: true})

	for _, warning := range result.Warnings {
		for _, unresolved := range []string{"review it by hand", "gave up repairing", "could not validate"} {
			if strings.Contains(warning, unresolved) {
				t.Errorf("the provider rejected generated config: %s", warning)
			}
		}
	}

	// The only TODOs allowed are for required values the Control API does not
	// return, and each one has to be reported in Missing so it reaches the user.
	for name, content := range files {
		for _, line := range strings.Split(content, "\n") {
			if !strings.Contains(line, "# TODO:") {
				continue
			}
			if !strings.Contains(line, "the Control API does not return it") {
				t.Errorf("%s has an unexplained TODO: %s", name, strings.TrimSpace(line))
			}
		}
	}
	if len(result.Missing) == 0 {
		t.Error("the Kafka rule's withheld credentials were not reported in Missing")
	}
}

// TestRunFlagsWithheldRequiredValues covers the Kafka SASL credentials. The
// Control API takes them on write and returns empty strings, so a literal export
// would write password = "" and wipe the live credential on apply.
func TestRunFlagsWithheldRequiredValues(t *testing.T) {
	result, files := runExport(t, Config{})
	config := files["app_chat_service.tf"]

	if regexp.MustCompile(`(?m)^\s*(password|username)\s*=\s*""`).MatchString(config) {
		t.Errorf("a withheld credential was written as an empty string:\n%s", config)
	}
	if !strings.Contains(config, "# TODO: password is required but the Control API does not return it") {
		t.Errorf("the withheld password was not flagged:\n%s", config)
	}

	var reported bool
	for _, missing := range result.Missing {
		if strings.Contains(missing, "ably_rule_kafka.chat_service.target.auth.sasl.password") {
			reported = true
		}
	}
	if !reported {
		t.Errorf("the withheld password was not reported, Missing was %v", result.Missing)
	}
}

// TestRunWiresWithheldRequiredValuesToVariables checks vars mode turns a withheld
// credential into a variable, the only mode that can yield config which applies
// once the value is supplied.
func TestRunWiresWithheldRequiredValuesToVariables(t *testing.T) {
	_, files := runExport(t, Config{Secrets: SecretsVars})
	config := files["app_chat_service.tf"]

	assertAssigns(t, config, "password", "var.rule_kafka_chat_service_target_auth_sasl_password")

	variables, ok := files["variables.tf"]
	if !ok {
		t.Fatalf("no variables.tf was generated, got %v", fileNames(files))
	}
	if !strings.Contains(variables, `variable "rule_kafka_chat_service_target_auth_sasl_password" {`) {
		t.Errorf("no variable was declared for the withheld password:\n%s", variables)
	}
	if !strings.Contains(variables, "Required, but the Control API does not return it.") {
		t.Errorf("the variable does not explain why it exists:\n%s", variables)
	}
}

// TestRunFlagsConfigItCannotRepair covers the Control API returning a value the
// provider won't accept. Writing it out quietly would fail at plan time for no
// visible reason, so the exporter warns and leaves a TODO by the resource.
func TestRunFlagsConfigItCannotRepair(t *testing.T) {
	fake := newFakeControlAPIWith(t, func(rules map[string]map[string]any) {
		config := rules["rule3"]["beforePublishConfig"].(map[string]any)
		config["failedAction"] = "not-a-valid-action"
	})

	result, err := Run(context.Background(), Config{Token: "fake-token", URL: fake.url()})
	if err != nil {
		t.Fatalf("Run should still finish, got %s", err)
	}

	var reported bool
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "ably_rule_bodyguard.chat_service") && strings.Contains(warning, "review it by hand") {
			reported = true
		}
	}
	if !reported {
		t.Errorf("expected a warning about the rejected config, warnings were %v", result.Warnings)
	}

	// The resource is still written, with the invalid value visible, so the
	// user can see what the API returned and decide what to do.
	var config string
	for _, file := range result.Files {
		if file.Name == "app_chat_service.tf" {
			config = string(file.Content)
		}
	}
	if !strings.Contains(config, "# TODO:") {
		t.Errorf("no TODO was left next to the resource:\n%s", config)
	}
	assertAssigns(t, config, "failed_action", `"not-a-valid-action"`)
}

// TestRunDoesNotWriteTheAccountID checks nothing identifying the account reaches
// the output, which is meant to be committed.
func TestRunDoesNotWriteTheAccountID(t *testing.T) {
	_, files := runExport(t, Config{Imports: true, Secrets: SecretsVars, ProviderVersion: DefaultProviderVersion})

	for name, content := range files {
		if strings.Contains(content, fakeAccountID) {
			t.Errorf("%s names the account:\n%s", name, content)
		}
	}
}

// TestWriteFilesRemovesStaleExports covers re-running with different flags. A
// main.tf left by -single-file beside the app_*.tf that replaced it declares the
// same resources twice, and Terraform rejects the directory.
func TestWriteFilesRemovesStaleExports(t *testing.T) {
	directory := t.TempDir()

	single, _ := runExport(t, Config{SingleFile: true})
	if err := WriteFiles(directory, single, false); err != nil {
		t.Fatalf("WriteFiles: %s", err)
	}
	if _, err := os.Stat(filepath.Join(directory, "main.tf")); err != nil {
		t.Fatalf("main.tf was not written: %s", err)
	}

	perApp, _ := runExport(t, Config{})
	if err := WriteFiles(directory, perApp, true); err != nil {
		t.Fatalf("WriteFiles with force: %s", err)
	}
	if _, err := os.Stat(filepath.Join(directory, "main.tf")); !os.IsNotExist(err) {
		t.Error("the stale main.tf survived, so two files now declare the same resources")
	}
	if _, err := os.Stat(filepath.Join(directory, "app_chat_service.tf")); err != nil {
		t.Errorf("the new file was not written: %s", err)
	}
}

// TestWriteFilesLeavesHandWrittenFilesAlone checks -force only clears the
// exporter's own output.
func TestWriteFilesLeavesHandWrittenFilesAlone(t *testing.T) {
	directory := t.TempDir()
	handWritten := filepath.Join(directory, "mine.tf")
	if err := os.WriteFile(handWritten, []byte("# mine\nresource \"ably_app\" \"mine\" {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, _ := runExport(t, Config{})
	if err := WriteFiles(directory, result, true); err != nil {
		t.Fatalf("WriteFiles with force: %s", err)
	}

	if _, err := os.Stat(handWritten); err != nil {
		t.Errorf("a hand-written file was deleted: %s", err)
	}
}

// TestWriteFilesRestrictsPermissionsOnSecrets checks output holding credentials
// isn't left world-readable.
func TestWriteFilesRestrictsPermissionsOnSecrets(t *testing.T) {
	for _, test := range []struct {
		secrets SecretMode
		want    os.FileMode
	}{
		{SecretsInline, 0o600},
		{SecretsVars, 0o644},
		{SecretsOmit, 0o644},
	} {
		t.Run(string(test.secrets), func(t *testing.T) {
			directory := t.TempDir()
			result, _ := runExport(t, Config{Secrets: test.secrets})
			if err := WriteFiles(directory, result, false); err != nil {
				t.Fatalf("WriteFiles: %s", err)
			}

			info, err := os.Stat(filepath.Join(directory, "app_chat_service.tf"))
			if err != nil {
				t.Fatal(err)
			}
			if got := info.Mode().Perm(); got != test.want {
				t.Errorf("%s mode wrote %#o, want %#o", test.secrets, got, test.want)
			}
		})
	}
}

// TestRunWarnsAboutUnmatchedAppFilters covers a typo alongside a good filter.
// Exporting only the app that matched, and saying nothing, looks like it worked.
func TestRunWarnsAboutUnmatchedAppFilters(t *testing.T) {
	fake := newFakeControlAPI(t)

	result, err := Run(context.Background(), Config{
		Token: "fake-token",
		URL:   fake.url(),
		Apps:  []string{"Chat Service", "typo-app"},
	})
	if err != nil {
		t.Fatalf("Run: %s", err)
	}
	if result.Counts["ably_app"] != 1 {
		t.Errorf("exported %d apps, want the one that matched", result.Counts["ably_app"])
	}

	var warned bool
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "typo-app") {
			warned = true
		}
	}
	if !warned {
		t.Errorf("no warning named the filter that matched nothing, warnings were %v", result.Warnings)
	}
}

// TestWriteFilesDoesNotWriteSecretsThroughAWorldReadableFile covers re-exporting
// over output from a previous run: writing in place would put the secrets into
// the existing 0644 file and only then tighten it.
func TestWriteFilesDoesNotWriteSecretsThroughAWorldReadableFile(t *testing.T) {
	directory := t.TempDir()

	// Stand in for a previous export made without secrets.
	loose, _ := runExport(t, Config{Secrets: SecretsOmit})
	if err := WriteFiles(directory, loose, false); err != nil {
		t.Fatalf("WriteFiles: %s", err)
	}

	withSecrets, _ := runExport(t, Config{Secrets: SecretsInline})
	if err := WriteFiles(directory, withSecrets, true); err != nil {
		t.Fatalf("WriteFiles with force: %s", err)
	}

	path := filepath.Join(directory, "app_chat_service.tf")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("re-export left mode %#o, want 0600", got)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "bodyguard-secret") {
		t.Error("the re-export did not replace the file contents")
	}
}

// TestBuildFilesKeepsOneFilePerApp covers two apps whose names sanitise to the
// same label. Filenames come from the label the labeller assigned, which is
// already unique, so each app keeps its own file.
func TestBuildFilesKeepsOneFilePerApp(t *testing.T) {
	exports := []exported{
		{target: Target{ResourceType: resourceTypeApp, ID: "app1", AppName: "Chat Service"}, label: "chat_service", appLabel: "chat_service"},
		{target: Target{ResourceType: resourceTypeApp, ID: "app2", AppName: "chat-service"}, label: "chat_service_2", appLabel: "chat_service_2"},
	}

	var names []string
	for _, file := range buildFiles(Config{}, exports, nil) {
		names = append(names, file.Name)
	}

	for _, want := range []string{"app_chat_service.tf", "app_chat_service_2.tf"} {
		if !slices.Contains(names, want) {
			t.Errorf("expected %s among %v", want, names)
		}
	}
}
