package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// recordCreatedIdentity stores the identity attributes of a freshly created
// resource in the Create response state before any later step can fail.
//
// Once the remote create call has succeeded the resource exists whether or not
// Create returns cleanly. If Create then returns an error with no state (a
// read-back failure, or a reconciliation trip-wire), Terraform has no record
// of the resource, and every retry creates and leaks another one. With the
// identity recorded, Terraform instead marks the resource tainted and replaces
// it on the next apply. The identity must be enough for Delete to work.
func recordCreatedIdentity(ctx context.Context, resp *resource.CreateResponse, identity map[string]string) {
	for name, value := range identity {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(name), types.StringValue(value))...)
	}
}

// ensureConfigured checks that the provider has been configured and appends an
// error diagnostic if it has not. Returns true when the provider is ready,
// false otherwise.
func (p *AblyProvider) ensureConfigured(diags *diag.Diagnostics) bool {
	if !p.configured {
		diags.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply",
		)
		return false
	}
	return true
}

// --- Generic pointer helpers ---

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// deref safely dereferences a pointer, returning the zero value of T if nil.
func deref[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}

// --- Terraform types ↔ Go pointer conversions ---

// optStringValue converts a *string to types.String.
// Returns types.StringNull() when nil.
func optStringValue(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// stringOrNull converts a Go string to types.String, mapping "" to null so
// that unset optional attributes round-trip as null rather than empty string.
func stringOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// optBoolValue converts a *bool to types.Bool.
// Returns types.BoolNull() when nil.
func optBoolValue(b *bool) types.Bool {
	if b != nil {
		return types.BoolValue(*b)
	}
	return types.BoolNull()
}

// optFloat64Value converts a *float64 to types.Float64.
// Returns types.Float64Null() when nil.
func optFloat64Value(f *float64) types.Float64 {
	if f == nil {
		return types.Float64Null()
	}
	return types.Float64Value(*f)
}

// optIntValue converts a *int to types.Int64.
// Returns types.Int64Null() when nil.
func optIntValue(i *int) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
}

// optionalStringPtr converts a types.String to a *string.
// Returns nil if the value is null, unknown, or empty.
func optionalStringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	if s == "" {
		return nil
	}
	return &s
}

// optionalBoolPtr converts a types.Bool to a *bool.
// Returns nil if the value is null or unknown.
func optionalBoolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}
