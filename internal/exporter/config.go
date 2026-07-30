package exporter

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// maxRepairs bounds the validate-and-repair loop. Each pass drops one
// attribute, and no real resource needs anything like this many.
const maxRepairs = 8

// configValue turns a resource's Terraform state into the configuration value a
// user could have written for it: attributes the provider computes are nulled
// out, at every level of nesting.
//
// This is the value the exporter validates and renders, rather than the raw
// state, so what gets written is exactly what a user is allowed to write.
func configValue(schema *tfprotov6.Schema, state tftypes.Value) (tftypes.Value, error) {
	if schema == nil || schema.Block == nil {
		return state, fmt.Errorf("resource has no schema")
	}
	return configBlock(schema.Block, state)
}

func configBlock(block *tfprotov6.SchemaBlock, value tftypes.Value) (tftypes.Value, error) {
	if value.IsNull() || !value.IsKnown() {
		return value, nil
	}

	attributes := map[string]tftypes.Value{}
	if err := value.As(&attributes); err != nil {
		return value, err
	}

	for _, attribute := range block.Attributes {
		if attribute == nil {
			continue
		}
		current, ok := attributes[attribute.Name]
		if !ok {
			continue
		}

		if isComputedOnly(attribute) || attribute.WriteOnly {
			attributes[attribute.Name] = tftypes.NewValue(current.Type(), nil)
			continue
		}
		if attribute.NestedType != nil {
			nested, err := configNested(attribute.NestedType, current)
			if err != nil {
				return value, err
			}
			attributes[attribute.Name] = nested
		}
	}

	for _, nestedBlock := range block.BlockTypes {
		if nestedBlock == nil {
			continue
		}
		current, ok := attributes[nestedBlock.TypeName]
		if !ok {
			continue
		}
		rebuilt, err := configCollection(current, func(element tftypes.Value) (tftypes.Value, error) {
			return configBlock(nestedBlock.Block, element)
		})
		if err != nil {
			return value, err
		}
		attributes[nestedBlock.TypeName] = rebuilt
	}

	return tftypes.NewValue(value.Type(), attributes), nil
}

func configNested(nested *tfprotov6.SchemaObject, value tftypes.Value) (tftypes.Value, error) {
	object := &tfprotov6.SchemaBlock{Attributes: nested.Attributes}
	if nested.Nesting == tfprotov6.SchemaObjectNestingModeSingle {
		return configBlock(object, value)
	}
	return configCollection(value, func(element tftypes.Value) (tftypes.Value, error) {
		return configBlock(object, element)
	})
}

// configCollection applies transform to every element of a list, set, map or
// single object, rebuilding the collection around the results.
func configCollection(value tftypes.Value, transform func(tftypes.Value) (tftypes.Value, error)) (tftypes.Value, error) {
	if value.IsNull() || !value.IsKnown() {
		return value, nil
	}

	valueType := value.Type()
	switch {
	case valueType.Is(tftypes.List{}), valueType.Is(tftypes.Set{}), valueType.Is(tftypes.Tuple{}):
		elements := []tftypes.Value{}
		if err := value.As(&elements); err != nil {
			return value, err
		}
		for index, element := range elements {
			transformed, err := transform(element)
			if err != nil {
				return value, err
			}
			elements[index] = transformed
		}
		return tftypes.NewValue(valueType, elements), nil

	// A single nested object or block arrives as an Object; transform it whole.
	case valueType.Is(tftypes.Object{}):
		return transform(value)

	case valueType.Is(tftypes.Map{}):
		elements := map[string]tftypes.Value{}
		if err := value.As(&elements); err != nil {
			return value, err
		}
		for key, element := range elements {
			transformed, err := transform(element)
			if err != nil {
				return value, err
			}
			elements[key] = transformed
		}
		return tftypes.NewValue(valueType, elements), nil

	default:
		return value, nil
	}
}

func isComputedOnly(attribute *tfprotov6.SchemaAttribute) bool {
	return attribute.Computed && !attribute.Optional && !attribute.Required
}

// repairConfig asks the provider to validate a generated configuration and
// removes redundant attributes the provider will not accept.
//
// This exists because validators are not part of the protocol schema: the
// provider can reject a combination of attributes that nothing in the schema
// warns about. ably_namespace is the case in hand. It has an `identified`
// attribute and a deprecated `authenticated` alias which conflict in config,
// and a read returns values for both.
//
// It only ever removes *deprecated* attributes. A deprecated attribute has a
// modern equivalent carrying the same information, so taking it out loses
// nothing. Anything else the provider objects to is left exactly as the Control
// API returned it and reported instead: if the provider rejects a value, that is
// a genuine disagreement between the provider and the API, and quietly deleting
// the user's configuration to make the validator happy would hide it.
//
// It returns the config, notes describing what it removed, and notes describing
// what it could not fix.
func repairConfig(ctx context.Context, bridge *bridge, resourceType, label string, schema *tfprotov6.Schema, config tftypes.Value) (tftypes.Value, []string, []string) {
	var dropped, unresolved []string

	for range maxRepairs {
		diagnostics, err := bridge.validate(ctx, resourceType, config)
		if err != nil {
			return config, dropped, append(unresolved, fmt.Sprintf(
				"could not validate the generated config for %s.%s: %s", resourceType, label, err))
		}

		problems := errorDiagnostics(diagnostics)
		if len(problems) == 0 {
			return config, dropped, unresolved
		}

		path := chooseRedundant(schema, config, problems)
		if path == nil {
			return config, dropped, append(unresolved, fmt.Sprintf(
				"the provider rejected the generated config for %s.%s, review it by hand: %s",
				resourceType, label, strings.Join(summaries(problems), "; ")))
		}

		repaired, err := nullifyPath(config, path.Steps())
		if err != nil {
			return config, dropped, append(unresolved, fmt.Sprintf(
				"could not drop %s from %s.%s: %s", formatPath(path), resourceType, label, err))
		}
		config = repaired
		dropped = append(dropped, fmt.Sprintf("dropped the deprecated %s.%s.%s: %s",
			resourceType, label, formatPath(path), firstSummary(problems, path)))
	}

	// The last pass may have fixed the last problem, so check before declaring
	// failure. Report the outstanding diagnostics when it didn't: a bare "gave
	// up" leaves nothing to act on.
	diagnostics, err := bridge.validate(ctx, resourceType, config)
	if err == nil {
		problems := errorDiagnostics(diagnostics)
		if len(problems) == 0 {
			return config, dropped, unresolved
		}
		return config, dropped, append(unresolved, fmt.Sprintf(
			"gave up repairing the generated config for %s.%s after %d attempts, review it by hand: %s",
			resourceType, label, maxRepairs, strings.Join(summaries(problems), "; ")))
	}
	return config, dropped, append(unresolved, fmt.Sprintf(
		"gave up repairing the generated config for %s.%s after %d attempts, and could not re-read why: %s",
		resourceType, label, maxRepairs, err))
}

// chooseRedundant picks the next deprecated attribute to remove: one the
// provider complained about, that a user could have set, and that currently
// carries a value. Ties break on the path so a re-run makes the same choice.
func chooseRedundant(schema *tfprotov6.Schema, config tftypes.Value, problems []*tfprotov6.Diagnostic) *tftypes.AttributePath {
	var candidates []*tftypes.AttributePath

	for _, problem := range problems {
		if problem.Attribute == nil {
			continue
		}
		attribute := attributeAtPath(schema, problem.Attribute.Steps())
		if attribute == nil || !attribute.Deprecated || attribute.Required || !attribute.Optional {
			continue
		}
		current, err := valueAtPath(config, problem.Attribute.Steps())
		if err != nil || current.IsNull() {
			continue
		}
		candidates = append(candidates, problem.Attribute)
	}

	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return formatPath(candidates[i]) < formatPath(candidates[j])
	})
	return candidates[0]
}

// attributeAtPath resolves an attribute path against a schema. It only follows
// attribute names, which is as deep as validator diagnostics reach in practice.
func attributeAtPath(schema *tfprotov6.Schema, steps []tftypes.AttributePathStep) *tfprotov6.SchemaAttribute {
	if schema == nil || schema.Block == nil || len(steps) == 0 {
		return nil
	}

	attributes := schema.Block.Attributes
	for index, step := range steps {
		name, ok := step.(tftypes.AttributeName)
		if !ok {
			return nil
		}
		var found *tfprotov6.SchemaAttribute
		for _, attribute := range attributes {
			if attribute != nil && attribute.Name == string(name) {
				found = attribute
				break
			}
		}
		if found == nil {
			return nil
		}
		if index == len(steps)-1 {
			return found
		}
		if found.NestedType == nil {
			return nil
		}
		attributes = found.NestedType.Attributes
	}
	return nil
}

// valueAtPath reads the value at an attribute path.
func valueAtPath(value tftypes.Value, steps []tftypes.AttributePathStep) (tftypes.Value, error) {
	if len(steps) == 0 {
		return value, nil
	}
	name, ok := steps[0].(tftypes.AttributeName)
	if !ok {
		return value, fmt.Errorf("unsupported path step %T", steps[0])
	}
	if value.IsNull() || !value.IsKnown() {
		return value, fmt.Errorf("no value at %s", name)
	}

	attributes := map[string]tftypes.Value{}
	if err := value.As(&attributes); err != nil {
		return value, err
	}
	child, ok := attributes[string(name)]
	if !ok {
		return value, fmt.Errorf("no attribute named %s", name)
	}
	return valueAtPath(child, steps[1:])
}

// nullifyPath returns a copy of value with the attribute at the given path set
// to null.
func nullifyPath(value tftypes.Value, steps []tftypes.AttributePathStep) (tftypes.Value, error) {
	if len(steps) == 0 {
		return tftypes.NewValue(value.Type(), nil), nil
	}
	name, ok := steps[0].(tftypes.AttributeName)
	if !ok {
		return value, fmt.Errorf("unsupported path step %T", steps[0])
	}

	attributes := map[string]tftypes.Value{}
	if err := value.As(&attributes); err != nil {
		return value, err
	}
	child, ok := attributes[string(name)]
	if !ok {
		return value, fmt.Errorf("no attribute named %s", name)
	}

	updated, err := nullifyPath(child, steps[1:])
	if err != nil {
		return value, err
	}
	attributes[string(name)] = updated

	return tftypes.NewValue(value.Type(), attributes), nil
}

// formatPath renders an attribute path the way it is written in HCL, rather
// than in the tftypes debug form.
func formatPath(path *tftypes.AttributePath) string {
	if path == nil {
		return "(root)"
	}

	var buf strings.Builder
	for _, step := range path.Steps() {
		switch step := step.(type) {
		case tftypes.AttributeName:
			if buf.Len() > 0 {
				buf.WriteString(".")
			}
			buf.WriteString(string(step))
		case tftypes.ElementKeyString:
			fmt.Fprintf(&buf, "[%q]", string(step))
		case tftypes.ElementKeyInt:
			fmt.Fprintf(&buf, "[%d]", int64(step))
		default:
			fmt.Fprintf(&buf, "[%s]", step)
		}
	}
	return buf.String()
}

func errorDiagnostics(diagnostics []*tfprotov6.Diagnostic) []*tfprotov6.Diagnostic {
	var problems []*tfprotov6.Diagnostic
	for _, diagnostic := range diagnostics {
		if diagnostic != nil && diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
			problems = append(problems, diagnostic)
		}
	}
	return problems
}

func summaries(diagnostics []*tfprotov6.Diagnostic) []string {
	messages := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		message := diagnostic.Summary
		if detail := strings.TrimSpace(diagnostic.Detail); detail != "" {
			message += ": " + detail
		}
		messages = append(messages, message)
	}
	return messages
}

// firstSummary returns the message for the diagnostic that named a path, so a
// dropped attribute can be explained in the exporter's own words.
func firstSummary(diagnostics []*tfprotov6.Diagnostic, path *tftypes.AttributePath) string {
	for _, diagnostic := range diagnostics {
		if diagnostic.Attribute != nil && diagnostic.Attribute.String() == path.String() {
			return strings.Join(summaries([]*tfprotov6.Diagnostic{diagnostic}), "")
		}
	}
	return "rejected by the provider"
}
