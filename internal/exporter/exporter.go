// Package exporter generates Terraform configuration for the resources in an
// Ably account, plus the import blocks to adopt them.
//
// It walks the account over the Control API, asks the provider to import and
// read each resource it finds, and renders the resulting state as HCL.
//
// Schemas come from the provider over protocol v6 (see bridge.go), so new
// attributes need no change here. A new rule type needs a line in the provider's
// RuleTypeResources; a new resource family needs a lister in discover.go. Tests
// fail if either is missed.
package exporter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"

	"github.com/ably/terraform-provider-ably/control"
)

// DefaultURL is the Control API the exporter talks to unless told otherwise,
// matching the provider's own default.
const DefaultURL = "https://control.ably.net/v1"

// DefaultProviderVersion is the version constraint written into the generated
// required_providers block.
const DefaultProviderVersion = "~> 1.0"

// Config is the input to an export run.
type Config struct {
	// Token is the Ably account token. Required.
	Token string
	// URL is the Control API base URL. Defaults to DefaultURL.
	URL string
	// Apps limits the export to apps matching these IDs or names. Empty
	// exports every app in the account.
	Apps []string
	// Secrets decides what happens to sensitive values. Defaults to
	// SecretsInline.
	Secrets SecretMode
	// Imports controls whether import blocks are generated.
	Imports bool
	// SingleFile writes everything to main.tf instead of one file per app.
	SingleFile bool
	// ProviderVersion is the version constraint for the generated
	// required_providers block. Empty omits the constraint.
	ProviderVersion string
	// Version is the provider version the exporter reports to the Control API
	// in its User-Agent.
	Version string
}

// File is one generated file.
type File struct {
	Name    string
	Content []byte
}

// Result is what an export produced.
type Result struct {
	// Files are the generated files, in the order they should be written.
	Files []File
	// Counts is the number of resources exported per resource type.
	Counts map[string]int
	// Total is the number of resources exported.
	Total int
	// Sensitive lists the sensitive attributes the export touched, as
	// `resource.address.attribute.path`.
	Sensitive []string
	// Missing lists required attributes the Control API did not return, which
	// have to be filled in before the config will apply.
	Missing []string
	// SecretsInline reports whether any secret was written into the output as a
	// literal, which decides how the files are permissioned.
	SecretsInline bool
	// Variables reports whether a variables.tf was generated.
	Variables bool
	// Warnings describes anything skipped or worth a second look.
	Warnings []string
}

// exported is a resource that has been read and rendered.
type exported struct {
	target   Target
	label    string
	importID string
	hcl      string
	// appLabel is the label of the app this resource belongs to, which decides
	// the file it lands in.
	appLabel string
	// notes are written above the resource as comments, for anything the
	// exporter could not resolve on its own.
	notes []string
}

// Run exports an account. It performs read-only Control API calls.
func Run(ctx context.Context, config Config) (*Result, error) {
	if config.Token == "" {
		return nil, errors.New("an Ably account token is required")
	}
	if config.URL == "" {
		config.URL = DefaultURL
	}
	if config.Secrets == "" {
		config.Secrets = SecretsInline
	}
	if config.Version == "" {
		config.Version = "dev"
	}

	client := control.NewClient(config.Token)
	client.BaseURL = config.URL
	client.UserAgent += " terraform-provider-ably-exporter/" + config.Version

	me, err := client.Me(ctx)
	if err != nil {
		return nil, fmt.Errorf("looking up the account for this token: %w", err)
	}
	if me.Account == nil || me.Account.ID == "" {
		return nil, errors.New("could not determine the account for this token; check it has account-level access")
	}
	accountID := me.Account.ID

	bridge, err := newBridge(ctx, config.Version)
	if err != nil {
		return nil, err
	}
	if err := bridge.configure(ctx, config.Token, config.URL); err != nil {
		return nil, err
	}

	found, err := discover(ctx, client, accountID, config.Apps)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Counts:   map[string]int{},
		Warnings: found.Warnings,
	}

	labels, appLabels, references := assignLabels(found.Targets)
	renderer := newRenderer(config.Secrets, references)

	var exports []exported
	for index, target := range found.Targets {
		schema, err := bridge.schema(target.ResourceType)
		if err != nil {
			return nil, err
		}

		importID, err := bridge.importID(target.ResourceType, target.AppID, target.ID)
		if err != nil {
			return nil, err
		}

		state, warnings, err := bridge.read(ctx, target.ResourceType, importID)
		result.Warnings = append(result.Warnings, prefixWarnings(warnings, target, labels[index])...)
		if err != nil {
			if errors.Is(err, errResourceGone) {
				result.Warnings = append(result.Warnings, fmt.Sprintf(
					"skipped %s %q: it was deleted while the export was running", target.ResourceType, importID))
				continue
			}
			return nil, err
		}

		// Render from the configuration the provider would accept rather than
		// from raw state: computed values dropped, and anything the provider's
		// validators reject taken back out.
		resourceConfig, err := configValue(schema, state)
		if err != nil {
			return nil, fmt.Errorf("deriving config for %s %q: %w", target.ResourceType, importID, err)
		}
		resourceConfig, dropped, unresolved := repairConfig(ctx, bridge, target.ResourceType, labels[index], schema, resourceConfig)
		result.Warnings = append(result.Warnings, dropped...)
		result.Warnings = append(result.Warnings, unresolved...)

		block, err := renderer.resourceBlock(target.ResourceType, labels[index], schema, resourceConfig)
		if err != nil {
			return nil, fmt.Errorf("rendering %s %q: %w", target.ResourceType, importID, err)
		}

		// An app owns the file its own resources go in.
		appLabel := appLabels[target.AppID]
		if target.ResourceType == resourceTypeApp {
			appLabel = labels[index]
		}

		exports = append(exports, exported{
			target:   target,
			label:    labels[index],
			importID: importID,
			hcl:      block,
			appLabel: appLabel,
			notes:    unresolved,
		})
		result.Counts[target.ResourceType]++
		result.Total++
	}

	result.Sensitive = renderer.sensitive
	result.Missing = renderer.missing
	result.SecretsInline = config.Secrets == SecretsInline && len(renderer.sensitive) > 0
	result.Warnings = append(result.Warnings, renderer.skipped...)
	result.Warnings = append(result.Warnings, renderer.missing...)

	result.Files = buildFiles(config, exports, renderer.variables)
	result.Variables = len(renderer.variables) > 0

	return result, nil
}

// assignLabels gives every target an HCL label and records the reference
// expression that points at it. Discovery emits an app before its children, so
// one pass can name a child after its app.
func assignLabels(targets []Target) (labels []string, appLabels map[string]string, references map[resourceRef]string) {
	labeller := newLabeller()
	labels = make([]string, len(targets))
	references = map[resourceRef]string{}
	appLabels = map[string]string{}

	for index, target := range targets {
		label := labeller.assign(target.ResourceType, labelFor(target, appLabels[target.AppID]))
		labels[index] = label

		if target.ResourceType == resourceTypeApp {
			appLabels[target.ID] = label
		}
		// referenceableAttributes decides which attributes get rewritten.
		references[resourceRef{target.ResourceType, target.ID}] = fmt.Sprintf("%s.%s.id", target.ResourceType, label)
	}

	return labels, appLabels, references
}

// buildFiles lays the rendered resources out into files.
func buildFiles(config Config, exports []exported, variables []variableDecl) []File {
	var files []File

	files = append(files, formatFile("provider.tf", providerFile(config)))

	grouped := map[string][]exported{}
	var order []string
	for _, export := range exports {
		// One file per app, named for the app's label. The labeller has already
		// made that unique, so two apps whose names sanitise alike keep separate
		// files.
		name := "main.tf"
		if !config.SingleFile {
			name = fmt.Sprintf("app_%s.tf", export.appLabel)
		}
		if _, seen := grouped[name]; !seen {
			order = append(order, name)
		}
		grouped[name] = append(grouped[name], export)
	}
	sort.Strings(order)

	for _, name := range order {
		var buf strings.Builder
		buf.WriteString(fileHeader())
		for _, export := range grouped[name] {
			buf.WriteString("\n")
			for _, note := range export.notes {
				fmt.Fprintf(&buf, "# TODO: %s\n", note)
			}
			buf.WriteString(export.hcl)
		}
		files = append(files, formatFile(name, buf.String()))
	}

	if len(variables) > 0 {
		files = append(files, formatFile("variables.tf", variablesFile(variables)))
	}

	if config.Imports && len(exports) > 0 {
		files = append(files, formatFile("imports.tf", importsFile(exports)))
	}

	return files
}

// generatedMarker opens every generated file, and is how WriteFiles recognises
// its own output.
const generatedMarker = "# Generated by the Ably Terraform exporter."

// fileHeader tops every generated file. It names no account: the output is meant
// to be committed.
func fileHeader() string {
	return generatedMarker + `
# Review before applying: values the Control API does not return, write-only
# secrets in particular, are absent and will be cleared on the first apply if
# you do not fill them in.
`
}

func providerFile(config Config) string {
	var buf strings.Builder
	buf.WriteString(fileHeader())
	buf.WriteString("\nterraform {\n  required_providers {\n    ably = {\n      source = \"ably/ably\"\n")
	if config.ProviderVersion != "" {
		fmt.Fprintf(&buf, "      version = %s\n", quoteString(config.ProviderVersion))
	}
	buf.WriteString("    }\n  }\n}\n\nprovider \"ably\" {\n")
	buf.WriteString("  # The account token is read from the ABLY_ACCOUNT_TOKEN environment variable.\n")
	if config.URL != DefaultURL {
		fmt.Fprintf(&buf, "  url = %s\n", quoteString(config.URL))
	}
	buf.WriteString("}\n")
	return buf.String()
}

func variablesFile(variables []variableDecl) string {
	var buf strings.Builder
	buf.WriteString(fileHeader())
	buf.WriteString("\n# Sensitive values are not written by the exporter. Supply them before applying.\n")

	sorted := make([]variableDecl, len(variables))
	copy(sorted, variables)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	seen := map[string]bool{}
	for _, variable := range sorted {
		if seen[variable.Name] {
			continue
		}
		seen[variable.Name] = true
		fmt.Fprintf(&buf, "\nvariable %s {\n", quoteString(variable.Name))
		fmt.Fprintf(&buf, "  description = %s\n", quoteString(variable.Description))
		fmt.Fprintf(&buf, "  type        = %s\n  sensitive   = true\n}\n", variable.Type)
	}
	return buf.String()
}

func importsFile(exports []exported) string {
	var buf strings.Builder
	buf.WriteString(fileHeader())
	buf.WriteString(`
# Import blocks adopt the resources above into Terraform state. Run
# ` + "`terraform plan`" + ` first: it should report the imports and no other
# changes. Delete this file once the imports have been applied.
`)

	for _, export := range exports {
		fmt.Fprintf(&buf, "\nimport {\n  to = %s.%s\n  id = %s\n}\n",
			export.target.ResourceType, export.label, quoteString(export.importID))
	}
	return buf.String()
}

// formatFile runs generated HCL through the same formatter `terraform fmt`
// uses, so the output looks hand-written rather than machine-emitted.
func formatFile(name, content string) File {
	return File{Name: name, Content: hclwrite.Format([]byte(content))}
}

// prefixWarnings labels provider diagnostics with the resource they came from.
func prefixWarnings(warnings []string, target Target, label string) []string {
	if len(warnings) == 0 {
		return nil
	}
	prefixed := make([]string, 0, len(warnings))
	for _, warning := range warnings {
		prefixed = append(prefixed, fmt.Sprintf("%s.%s: %s", target.ResourceType, label, warning))
	}
	return prefixed
}

// WriteFiles writes an export to a directory, creating it if needed.
//
// Without force it refuses a directory that already holds Terraform files. With
// force it first removes files a previous export wrote, because re-running with
// different flags otherwise leaves a stale main.tf beside the app_*.tf that
// replaced it, and duplicate resource blocks make Terraform reject the
// directory. Files holding secrets are written 0600.
func WriteFiles(directory string, result *Result, force bool) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", directory, err)
	}

	existing, err := filepath.Glob(filepath.Join(directory, "*.tf"))
	if err != nil {
		return err
	}
	if len(existing) > 0 && !force {
		return fmt.Errorf("%s already contains Terraform files (%s); move them aside or pass -force",
			directory, strings.Join(baseNames(existing), ", "))
	}

	if force {
		if err := removeStaleExports(existing, result); err != nil {
			return err
		}
	}

	mode := os.FileMode(0o644)
	if result.SecretsInline {
		mode = 0o600
	}

	for _, file := range result.Files {
		if err := writeFile(filepath.Join(directory, file.Name), file.Content, mode); err != nil {
			return err
		}
	}
	return nil
}

// writeFile writes content through a temporary file created with mode, then
// renames it into place.
//
// Writing in place would put the content into whatever mode the existing file
// already has, so re-exporting over a world-readable file would expose secrets
// until the chmod landed. The rename also leaves the previous file intact if
// anything fails part-way.
func writeFile(path string, content []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	name := temporary.Name()
	// Unconditional cleanup: after a successful rename both calls fail
	// harmlessly, and on an early return they tidy the temporary file away.
	defer func() {
		_ = temporary.Close()
		_ = os.Remove(name)
	}()

	// CreateTemp makes the file 0600; widen it only if the content allows.
	if err := temporary.Chmod(mode); err != nil {
		return fmt.Errorf("setting permissions on %s: %w", path, err)
	}
	if _, err := temporary.Write(content); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := os.Rename(name, path); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// removeStaleExports deletes Terraform files a previous export wrote that this
// one won't replace. It only removes files carrying the exporter's own header, so
// hand-written config survives -force.
func removeStaleExports(existing []string, result *Result) error {
	writing := make(map[string]bool, len(result.Files))
	for _, file := range result.Files {
		writing[file.Name] = true
	}

	for _, path := range existing {
		if writing[filepath.Base(path)] {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		if !strings.HasPrefix(string(content), generatedMarker) {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing stale %s: %w", path, err)
		}
	}
	return nil
}

func baseNames(paths []string) []string {
	names := make([]string, 0, len(paths))
	for _, path := range paths {
		names = append(names, filepath.Base(path))
	}
	return names
}

// Summary describes an export in the form the CLI prints.
func (r *Result) Summary() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "Exported %d resources:\n", r.Total)

	types := make([]string, 0, len(r.Counts))
	for resourceType := range r.Counts {
		types = append(types, resourceType)
	}
	sort.Strings(types)
	for _, resourceType := range types {
		fmt.Fprintf(&buf, "  %-40s %d\n", resourceType, r.Counts[resourceType])
	}
	return buf.String()
}
