// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// This file holds the plumbing shared by the data sources.
//
// The Control API has no GET-by-ID for apps, keys, namespaces or queues; only
// rules have one. Everything else is a list endpoint, so the generated data
// sources are all plural: a parent path parameter plus a computed set of whole
// objects. The singular data sources are built from the same generated schema by
// lifting the set's element attributes to the top level, so there is one
// generated source of truth per entity and the two can't drift apart. Their Read
// lists and filters client side, which is what the resources already do (see
// resource_ably_app.go's Read and the NOTE on it).

// elementAttributes returns a copy of the element attributes of a generated
// plural data source's collection, with the generated CustomTypes stripped so a
// plain-struct tfsdk model reflects cleanly. This is the attribute set a singular
// data source serves.
func elementAttributes(s schema.Schema, collection string) map[string]schema.Attribute {
	set, ok := s.Attributes[collection].(schema.SetNestedAttribute)
	if !ok {
		return map[string]schema.Attribute{}
	}

	attributes := make(map[string]schema.Attribute, len(set.NestedObject.Attributes))
	for name, attribute := range set.NestedObject.Attributes {
		attributes[name] = attribute
	}
	stripDataSourceCustomTypes(attributes)
	return attributes
}

// stripSetNestedCustomTypes strips the generated CustomTypes from a plural data
// source's collection attribute and everything nested inside it.
func stripSetNestedCustomTypes(s *schema.Schema, collection string) {
	set, ok := s.Attributes[collection].(schema.SetNestedAttribute)
	if !ok {
		return
	}
	set.CustomType = nil
	set.NestedObject.CustomType = nil
	stripDataSourceCustomTypes(set.NestedObject.Attributes)
	s.Attributes[collection] = set
}

// stripDataSourceCustomTypes strips generated CustomTypes from an attribute map,
// recursing into nested blocks.
func stripDataSourceCustomTypes(attributes map[string]schema.Attribute) {
	for name, attribute := range attributes {
		switch typed := attribute.(type) {
		case schema.SingleNestedAttribute:
			typed.CustomType = nil
			stripDataSourceCustomTypes(typed.Attributes)
			attributes[name] = typed
		case schema.SetNestedAttribute:
			typed.CustomType = nil
			typed.NestedObject.CustomType = nil
			stripDataSourceCustomTypes(typed.NestedObject.Attributes)
			attributes[name] = typed
		case schema.ListNestedAttribute:
			typed.CustomType = nil
			typed.NestedObject.CustomType = nil
			stripDataSourceCustomTypes(typed.NestedObject.Attributes)
			attributes[name] = typed
		}
	}
}

// requireString makes a generated computed string attribute a required argument,
// for a parent identifier a singular lookup can't work without.
func requireString(attributes map[string]schema.Attribute, name, description string) {
	attribute, ok := attributes[name].(schema.StringAttribute)
	if !ok {
		return
	}
	attribute.Required = true
	attribute.Computed = false
	attribute.Optional = false
	if description != "" {
		attribute.Description = description
		attribute.MarkdownDescription = description
	}
	attributes[name] = attribute
}

// optionalString makes a generated computed string attribute an optional
// argument that is still computed, which is the shape of a lookup key: set it to
// search by it, leave it out and it comes back from the API.
func optionalString(attributes map[string]schema.Attribute, name, description string) {
	attribute, ok := attributes[name].(schema.StringAttribute)
	if !ok {
		return
	}
	attribute.Optional = true
	attribute.Computed = true
	// The generated plural schemas mark the parent path parameter Required, and
	// the framework rejects Required alongside Optional/Computed.
	attribute.Required = false
	if description != "" {
		attribute.Description = description
		attribute.MarkdownDescription = description
	}
	attributes[name] = attribute
}

// lookup identifies which record a singular data source wants. Exactly one of id
// or name must be set: the Control API has no lookup endpoint for these entities,
// so a data source lists them and matches locally.
type lookup struct {
	// dataSourceName is the Terraform type name, for diagnostics.
	dataSourceName string
	// id and name are the caller's arguments; exactly one carries a value.
	id, name types.String
}

// validate checks that exactly one lookup key was given.
func (l lookup) validate() diag.Diagnostics {
	var diags diag.Diagnostics

	hasID := !l.id.IsNull() && l.id.ValueString() != ""
	hasName := !l.name.IsNull() && l.name.ValueString() != ""

	switch {
	case hasID && hasName:
		diags.AddError(
			fmt.Sprintf("Ambiguous %s lookup", l.dataSourceName),
			"Set either id or name, not both.",
		)
	case !hasID && !hasName:
		diags.AddError(
			fmt.Sprintf("Incomplete %s lookup", l.dataSourceName),
			"Set either id or name to say which one you want.",
		)
	}

	return diags
}

// findOne picks the single record matching the lookup out of everything the API
// returned.
//
// Names are not unique in the Control API, so a name that matches more than once
// is an error rather than a silent pick: returning an arbitrary match would make
// the config's meaning depend on API ordering. The error lists the IDs so the
// caller can switch to looking up by id.
func findOne[T any](l lookup, records []T, idOf, nameOf func(T) string) (T, diag.Diagnostics) {
	var zero T

	diags := l.validate()
	if diags.HasError() {
		return zero, diags
	}

	if id := l.id.ValueString(); id != "" {
		for _, record := range records {
			if idOf(record) == id {
				return record, diags
			}
		}
		diags.AddError(
			fmt.Sprintf("No %s with id %q", l.dataSourceName, id),
			fmt.Sprintf("Found %d record(s), none with that id. Check the id, and that the token has access to it.", len(records)),
		)
		return zero, diags
	}

	name := l.name.ValueString()
	var matches []T
	var matchedIDs []string
	for _, record := range records {
		if nameOf(record) == name {
			matches = append(matches, record)
			matchedIDs = append(matchedIDs, idOf(record))
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], diags
	case 0:
		diags.AddError(
			fmt.Sprintf("No %s named %q", l.dataSourceName, name),
			fmt.Sprintf("Found %d record(s), none with that name.", len(records)),
		)
	default:
		sort.Strings(matchedIDs)
		diags.AddError(
			fmt.Sprintf("More than one %s named %q", l.dataSourceName, name),
			fmt.Sprintf("Names are not unique in the Control API and %d records share this one (ids: %v). Look it up by id instead.", len(matches), matchedIDs),
		)
	}

	return zero, diags
}

// readDataSourceError is the diagnostic for a failed Control API read, worded the
// same way for every data source.
func readDataSourceError(diags *diag.Diagnostics, dataSourceName string, err error) {
	diags.AddError(
		fmt.Sprintf("Error reading %s", dataSourceName),
		fmt.Sprintf("Could not read %s, unexpected error: %s", dataSourceName, err.Error()),
	)
}

// dataSourceAccountID resolves the account the caller means: the account_id
// argument when set, otherwise the one the provider is configured with.
func dataSourceAccountID(p *AblyProvider, argument types.String) string {
	if !argument.IsNull() && argument.ValueString() != "" {
		return argument.ValueString()
	}
	return p.accountID
}

// stringsToList converts a []string into a types.List of strings.
func stringsToList(ctx context.Context, values []string) (types.List, diag.Diagnostics) {
	if values == nil {
		return types.ListNull(types.StringType), nil
	}
	return types.ListValueFrom(ctx, types.StringType, values)
}
