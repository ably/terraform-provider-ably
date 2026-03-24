// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// defaultBoolModifier implements planmodifier.Bool using the built-in pattern.
type defaultBoolModifier struct {
	value types.Bool
}

func (m defaultBoolModifier) Description(_ context.Context) string {
	return "If the config does not contain a value, a default will be set using val."
}

func (m defaultBoolModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m defaultBoolModifier) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if resp.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}
	if !req.ConfigValue.IsNull() || !req.PlanValue.IsNull() {
		return
	}
	resp.PlanValue = m.value
}

// DefaultBoolAttribute returns a plan modifier that sets a default bool value.
func DefaultBoolAttribute(value types.Bool) planmodifier.Bool {
	return defaultBoolModifier{value: value}
}

// defaultInt64Modifier implements planmodifier.Int64.
type defaultInt64Modifier struct {
	value types.Int64
}

func (m defaultInt64Modifier) Description(_ context.Context) string {
	return "If the config does not contain a value, a default will be set using val."
}

func (m defaultInt64Modifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m defaultInt64Modifier) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if resp.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}
	if !req.ConfigValue.IsNull() || !req.PlanValue.IsNull() {
		return
	}
	resp.PlanValue = m.value
}

// DefaultInt64Attribute returns a plan modifier that sets a default int64 value.
func DefaultInt64Attribute(value types.Int64) planmodifier.Int64 {
	return defaultInt64Modifier{value: value}
}

// defaultStringModifier implements planmodifier.String.
type defaultStringModifier struct {
	value types.String
}

func (m defaultStringModifier) Description(_ context.Context) string {
	return "If the config does not contain a value, a default will be set using val."
}

func (m defaultStringModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m defaultStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if resp.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}
	if !req.ConfigValue.IsNull() || !req.PlanValue.IsNull() {
		return
	}
	resp.PlanValue = m.value
}

// DefaultStringAttribute returns a plan modifier that sets a default string value.
func DefaultStringAttribute(value types.String) planmodifier.String {
	return defaultStringModifier{value: value}
}

// sortSetsInMapModifier normalizes set element ordering within a map attribute
// so that the planned value and the apply result use the same element order.
// This works around a Terraform core issue where set elements inside map values
// are compared positionally rather than as true sets.
type sortSetsInMapModifier struct{}

func (m sortSetsInMapModifier) Description(_ context.Context) string {
	return "Sorts set elements within map values to ensure consistent ordering."
}

func (m sortSetsInMapModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m sortSetsInMapModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	elements := req.PlanValue.Elements()
	normalized := make(map[string]attr.Value, len(elements))
	for k, v := range elements {
		setVal, ok := v.(types.Set)
		if !ok || setVal.IsNull() || setVal.IsUnknown() {
			normalized[k] = v
			continue
		}
		var elems []types.String
		diags := setVal.ElementsAs(ctx, &elems, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		strs := make([]string, len(elems))
		for i, e := range elems {
			strs[i] = e.ValueString()
		}
		sort.Strings(strs)

		attrVals := make([]attr.Value, len(strs))
		for i, s := range strs {
			attrVals[i] = types.StringValue(s)
		}
		normalized[k] = types.SetValueMust(types.StringType, attrVals)
	}

	mapVal, diags := types.MapValue(
		types.SetType{ElemType: types.StringType},
		normalized,
	)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.PlanValue = mapVal
}

// SortSetsInMap returns a plan modifier that normalizes set element ordering
// within a map attribute.
func SortSetsInMap() planmodifier.Map {
	return sortSetsInMapModifier{}
}
