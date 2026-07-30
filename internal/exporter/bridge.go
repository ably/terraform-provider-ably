package exporter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/ably/terraform-provider-ably/internal/provider"
)

// terraformVersion is the version the bridge reports to the provider. It only
// ever reaches the provider's User-Agent and logs; nothing gates on it.
const terraformVersion = "1.9.0"

// errResourceGone reports that a resource vanished between being listed and
// being read. Accounts change under us, so this is expected, not exceptional.
var errResourceGone = errors.New("resource no longer exists")

// bridge runs the provider in-process and drives it over the protocol-v6 RPCs
// Terraform uses: GetProviderSchema, ConfigureProvider, ImportResourceState,
// ReadResource and ValidateResourceConfig.
//
// Using the protocol rather than the provider's Go types means schemas,
// sensitivity and the API-to-state mapping all come from the provider as
// compiled, and the state rendered to HCL is what Terraform would hold after
// terraform import.
type bridge struct {
	server    tfprotov6.ProviderServer
	provider  *tfprotov6.Schema
	resources map[string]*tfprotov6.Schema
}

// newBridge starts the provider and reads its schemas.
func newBridge(ctx context.Context, version string) (*bridge, error) {
	server := providerserver.NewProtocol6(provider.New(version)())()

	resp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		return nil, fmt.Errorf("reading provider schema: %w", err)
	}
	if _, err := splitDiagnostics(resp.Diagnostics); err != nil {
		return nil, fmt.Errorf("reading provider schema: %w", err)
	}

	return &bridge{
		server:    server,
		provider:  resp.Provider,
		resources: resp.ResourceSchemas,
	}, nil
}

// configure configures the provider with an account token and Control API URL.
// The config object is built from the provider's own schema, so attributes added
// to the provider block later need no change here.
func (b *bridge) configure(ctx context.Context, token, url string) error {
	configType := b.provider.ValueType()
	object, ok := configType.(tftypes.Object)
	if !ok {
		return fmt.Errorf("provider config type is %T, expected an object", configType)
	}

	attributes := map[string]tftypes.Value{}
	for name, attrType := range object.AttributeTypes {
		switch {
		case name == "token":
			attributes[name] = tftypes.NewValue(attrType, token)
		case name == "url" && url != "":
			attributes[name] = tftypes.NewValue(attrType, url)
		default:
			attributes[name] = tftypes.NewValue(attrType, nil)
		}
	}

	config, err := tfprotov6.NewDynamicValue(configType, tftypes.NewValue(configType, attributes))
	if err != nil {
		return fmt.Errorf("encoding provider config: %w", err)
	}

	resp, err := b.server.ConfigureProvider(ctx, &tfprotov6.ConfigureProviderRequest{
		TerraformVersion: terraformVersion,
		Config:           &config,
	})
	if err != nil {
		return fmt.Errorf("configuring provider: %w", err)
	}
	if _, err := splitDiagnostics(resp.Diagnostics); err != nil {
		return fmt.Errorf("configuring provider: %w", err)
	}
	return nil
}

// schema returns the schema for a resource type.
func (b *bridge) schema(resourceType string) (*tfprotov6.Schema, error) {
	schema, ok := b.resources[resourceType]
	if !ok {
		return nil, fmt.Errorf("provider does not serve resource type %q", resourceType)
	}
	return schema, nil
}

// read imports a resource by its import ID and reads it, returning the resulting
// state. These are the calls terraform import makes, so the import IDs written
// into imports.tf are ones the exporter has just proved work.
func (b *bridge) read(ctx context.Context, resourceType, importID string) (tftypes.Value, []string, error) {
	schema, err := b.schema(resourceType)
	if err != nil {
		return tftypes.Value{}, nil, err
	}
	stateType := schema.ValueType()

	var warnings []string

	importResp, err := b.server.ImportResourceState(ctx, &tfprotov6.ImportResourceStateRequest{
		TypeName: resourceType,
		ID:       importID,
	})
	if err != nil {
		return tftypes.Value{}, warnings, fmt.Errorf("importing %s %q: %w", resourceType, importID, err)
	}
	importWarnings, err := splitDiagnostics(importResp.Diagnostics)
	warnings = append(warnings, importWarnings...)
	if err != nil {
		return tftypes.Value{}, warnings, fmt.Errorf("importing %s %q: %w", resourceType, importID, err)
	}
	if len(importResp.ImportedResources) == 0 {
		return tftypes.Value{}, warnings, fmt.Errorf("importing %s %q: provider returned no state", resourceType, importID)
	}
	imported := importResp.ImportedResources[0]

	readResp, err := b.server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{
		TypeName:     resourceType,
		CurrentState: imported.State,
		Private:      imported.Private,
	})
	if err != nil {
		return tftypes.Value{}, warnings, fmt.Errorf("reading %s %q: %w", resourceType, importID, err)
	}
	readWarnings, err := splitDiagnostics(readResp.Diagnostics)
	warnings = append(warnings, readWarnings...)
	if err != nil {
		return tftypes.Value{}, warnings, fmt.Errorf("reading %s %q: %w", resourceType, importID, err)
	}
	if readResp.NewState == nil {
		return tftypes.Value{}, warnings, errResourceGone
	}

	state, err := readResp.NewState.Unmarshal(stateType)
	if err != nil {
		return tftypes.Value{}, warnings, fmt.Errorf("decoding %s %q state: %w", resourceType, importID, err)
	}
	// The provider signals "not found" by removing the resource from state,
	// which reaches us as a null object.
	if state.IsNull() {
		return tftypes.Value{}, warnings, errResourceGone
	}

	return state, warnings, nil
}

// validate asks the provider whether it would accept a configuration. It is how
// the exporter learns about anything the provider enforces with validators
// rather than schema flags, conflicting attributes in particular.
func (b *bridge) validate(ctx context.Context, resourceType string, config tftypes.Value) ([]*tfprotov6.Diagnostic, error) {
	schema, err := b.schema(resourceType)
	if err != nil {
		return nil, err
	}

	encoded, err := tfprotov6.NewDynamicValue(schema.ValueType(), config)
	if err != nil {
		return nil, fmt.Errorf("encoding %s config: %w", resourceType, err)
	}

	resp, err := b.server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{
		TypeName: resourceType,
		Config:   &encoded,
	})
	if err != nil {
		return nil, err
	}
	return resp.Diagnostics, nil
}

// hasAttribute reports whether a resource schema has a top-level attribute.
func hasAttribute(schema *tfprotov6.Schema, name string) bool {
	if schema == nil || schema.Block == nil {
		return false
	}
	for _, attribute := range schema.Block.Attributes {
		if attribute != nil && attribute.Name == name {
			return true
		}
	}
	return false
}

// importID builds the import identifier for a resource. Resources inside an app
// take "appID,id" (see ImportResource in the provider package), account-level
// ones the bare ID. Keying off the app_id attribute rather than a hardcoded list
// means a future account-level resource works untouched.
func (b *bridge) importID(resourceType, appID, id string) (string, error) {
	schema, err := b.schema(resourceType)
	if err != nil {
		return "", err
	}
	if hasAttribute(schema, "app_id") {
		if appID == "" {
			return "", fmt.Errorf("resource type %s is app-scoped but no app ID was given", resourceType)
		}
		return appID + "," + id, nil
	}
	return id, nil
}

// splitDiagnostics separates warnings from errors, returning the warning
// messages and a single error covering every error diagnostic.
func splitDiagnostics(diagnostics []*tfprotov6.Diagnostic) ([]string, error) {
	var warnings []string
	var errs []string

	for _, diagnostic := range diagnostics {
		if diagnostic == nil {
			continue
		}
		message := diagnostic.Summary
		if detail := strings.TrimSpace(diagnostic.Detail); detail != "" {
			message += ": " + detail
		}
		if diagnostic.Attribute != nil {
			message += fmt.Sprintf(" (at %s)", diagnostic.Attribute)
		}

		switch diagnostic.Severity {
		case tfprotov6.DiagnosticSeverityError:
			errs = append(errs, message)
		default:
			warnings = append(warnings, message)
		}
	}

	if len(errs) > 0 {
		return warnings, errors.New(strings.Join(errs, "; "))
	}
	return warnings, nil
}
