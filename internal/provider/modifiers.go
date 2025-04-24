// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DefaultAttributePlanModifier struct {
	Bool   types.Bool
	Int64  types.Int64
	String types.String
}

func (m DefaultAttributePlanModifier) Description(ctx context.Context) string {
	return "If the config does not contain a value, a default will be set using val."
}

func (m DefaultAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func DefaultBoolAttribute(value types.Bool) DefaultAttributePlanModifier {
	return DefaultAttributePlanModifier{Bool: value}
}

func (m DefaultAttributePlanModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if resp.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.IsNull() || !req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue = m.Bool
}

func DefaultInt64Attribute(value types.Int64) DefaultAttributePlanModifier {
	return DefaultAttributePlanModifier{Int64: value}
}

func (m DefaultAttributePlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if resp.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.IsNull() || !req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue = m.Int64
}

func DefaultStringAttribute(value types.String) DefaultAttributePlanModifier {
	return DefaultAttributePlanModifier{String: value}
}

func (m DefaultAttributePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if resp.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	if !req.ConfigValue.IsNull() || !req.PlanValue.IsNull() {
		return
	}

	resp.PlanValue = m.String
}
