package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type DefaultAttributePlanModifier struct {
	Default attr.Value
}

func DefaultAttribute(value attr.Value) DefaultAttributePlanModifier {
	return DefaultAttributePlanModifier{Default: value}
}

func (m DefaultAttributePlanModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if resp.AttributePlan == nil || req.AttributeConfig == nil {
		return
	}

	if !req.AttributeConfig.IsNull() {
		return
	}

	resp.AttributePlan = m.Default
}

func (m DefaultAttributePlanModifier) Description(ctx context.Context) string {
	return "If the config does not contain a value, a default will be set using val."
}

func (m DefaultAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
