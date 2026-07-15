// Package planmodifiers holds plan modifiers shared by generated schemas and
// hand-written resources. Generated code references these via the overrides
// table in codegen/ruletypesgen, so behaviour that the Provider Code Spec
// cannot express inline lives here once instead of being pasted into every
// generated file.
package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// RequiresReplaceWhenCleared forces resource replacement when a configured
// string attribute is removed from the configuration.
//
// The Control API's rule PATCH endpoints validate against a schema that
// neither accepts null nor lets a pattern-constrained field (e.g.
// chatRoomFilter, pattern ^/.*/$) be set to "" — verified against the live
// API, 2026-07-08. Omitting the field keeps the stored value, so an in-place
// update can never unset it: recreating the rule without the field is the
// only mechanism the API offers, and this modifier makes Terraform plan
// exactly that. Changing the value (rather than removing it) stays an
// in-place update.
func RequiresReplaceWhenCleared() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			if !req.StateValue.IsNull() && req.ConfigValue.IsNull() {
				resp.RequiresReplace = true
			}
		},
		"Removing this attribute requires replacing the resource.",
		"The Control API cannot unset this attribute in an update, so removing it recreates the rule without it.",
	)
}
