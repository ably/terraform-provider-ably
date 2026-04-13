package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// reconciler accumulates diagnostics so callers can reconcile many fields
// without checking errors after each one.
//
// When reading is true (set via forRead), every field is treated as computed
// so the API response is always accepted. This is necessary because during
// Read (including import), the prior state may be empty — there is no user
// plan to compare against.
type reconciler struct {
	diags   *diag.Diagnostics
	reading bool
}

func newReconciler(diags *diag.Diagnostics) *reconciler {
	return &reconciler{diags: diags}
}

func (r *reconciler) forRead() *reconciler {
	r.reading = true
	return r
}

func (r *reconciler) str(field string, input, output types.String, computed bool) types.String {
	v, err := reconcileString(field, input, output, computed || r.reading)
	if err != nil {
		r.diags.AddError("State reconciliation error", err.Error())
	}
	return v
}

func (r *reconciler) boolean(field string, input, output types.Bool, computed bool) types.Bool {
	v, err := reconcileBool(field, input, output, computed || r.reading)
	if err != nil {
		r.diags.AddError("State reconciliation error", err.Error())
	}
	return v
}

func (r *reconciler) int64val(field string, input, output types.Int64, computed bool) types.Int64 {
	v, err := reconcileInt64(field, input, output, computed || r.reading)
	if err != nil {
		r.diags.AddError("State reconciliation error", err.Error())
	}
	return v
}

// reconcileSlice reconciles a plan/state slice with an API response slice.
// Go does not allow generic methods on structs, so this is a standalone helper
// that takes the reconciler for diagnostics and reading mode.
func reconcileSlice[T any](field string, input, output []T, computed bool) ([]T, error) {
	inputEmpty := len(input) == 0
	outputEmpty := len(output) == 0

	switch {
	case !inputEmpty && !outputEmpty:
		return output, nil
	case !inputEmpty && outputEmpty:
		return input, nil
	case inputEmpty && !outputEmpty:
		if computed {
			return output, nil
		}
		return nil, fmt.Errorf(
			"reconcile %q: API returned %d elements but field was not set in config and is not computed",
			field, len(output),
		)
	default:
		return nil, nil
	}
}

// rcSlice is the reconciler-aware wrapper for reconcileSlice.
func rcSlice[T any](rc *reconciler, field string, input, output []T, computed bool) []T {
	v, err := reconcileSlice(field, input, output, computed || rc.reading)
	if err != nil {
		rc.diags.AddError("State reconciliation error", err.Error())
	}
	return v
}

// reconcileMapSet reconciles a plan/state map[string]types.Set with an API response map.
func reconcileMapSet(field string, input, output map[string]types.Set, computed bool) (map[string]types.Set, error) {
	inputEmpty := len(input) == 0
	outputEmpty := len(output) == 0

	switch {
	case !inputEmpty && !outputEmpty:
		return output, nil
	case !inputEmpty && outputEmpty:
		return input, nil
	case inputEmpty && !outputEmpty:
		if computed {
			return output, nil
		}
		return nil, fmt.Errorf(
			"reconcile %q: API returned %d entries but field was not set in config and is not computed",
			field, len(output),
		)
	default:
		return nil, nil
	}
}

func (r *reconciler) mapSet(field string, input, output map[string]types.Set, computed bool) map[string]types.Set {
	v, err := reconcileMapSet(field, input, output, computed || r.reading)
	if err != nil {
		r.diags.AddError("State reconciliation error", err.Error())
	}
	return v
}

// --- Plan field accessors ---
// These extract a typed field from a possibly-nil plan target pointer.
// When the plan target is nil (e.g. during import or type mismatch),
// they return the zero value which reconcile treats as empty/null.

func planStr[T any](pt *T, fn func(*T) types.String) types.String {
	if pt == nil {
		return types.StringNull()
	}
	return fn(pt)
}

func planBool[T any](pt *T, fn func(*T) types.Bool) types.Bool {
	if pt == nil {
		return types.BoolNull()
	}
	return fn(pt)
}

func planInt64[T any](pt *T, fn func(*T) types.Int64) types.Int64 {
	if pt == nil {
		return types.Int64Null()
	}
	return fn(pt)
}

func planSlice[T any, E any](pt *T, fn func(*T) []E) []E {
	if pt == nil {
		return nil
	}
	return fn(pt)
}

// optIntFromIntPtr converts a *int to types.Int64, returning types.Int64Null() when nil.
func optIntFromIntPtr(i *int) types.Int64 {
	if i == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*i))
}

// optStringValue converts a non-pointer string to types.String, treating "" as null.
// For pointer strings, use the existing optStringValue in helpers.go.
func optStringFromString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// Reconcile functions handle the four-way matrix of input (plan/state) vs
// output (API response) emptiness:
//
//  1. Input non-empty, Output non-empty → return output (API is authoritative)
//  2. Input non-empty, Output empty     → return input  (write-only / sensitive field)
//  3. Input empty,     Output non-empty:
//     - computed=true  → return output (server-owned or defaulted field)
//     - computed=false → return error  (unexpected value for optional-only field)
//  4. Input empty,     Output empty     → return null
//
// "Empty" means null, unknown, or (for strings) the zero-length string "".

// reconcileString reconciles a plan/state string with an API response string.
func reconcileString(field string, input, output types.String, computed bool) (types.String, error) {
	inputEmpty := input.IsNull() || input.IsUnknown() || input.ValueString() == ""
	outputEmpty := output.IsNull() || output.IsUnknown() || output.ValueString() == ""

	switch {
	case !inputEmpty && !outputEmpty:
		return output, nil
	case !inputEmpty && outputEmpty:
		return input, nil
	case inputEmpty && !outputEmpty:
		if computed {
			return output, nil
		}
		return types.StringNull(), fmt.Errorf(
			"reconcile %q: API returned %q but field was not set in config and is not computed",
			field, output.ValueString(),
		)
	default:
		return types.StringNull(), nil
	}
}

// reconcileBool reconciles a plan/state bool with an API response bool.
func reconcileBool(field string, input, output types.Bool, computed bool) (types.Bool, error) {
	inputEmpty := input.IsNull() || input.IsUnknown()
	outputEmpty := output.IsNull() || output.IsUnknown()

	switch {
	case !inputEmpty && !outputEmpty:
		return output, nil
	case !inputEmpty && outputEmpty:
		return input, nil
	case inputEmpty && !outputEmpty:
		if computed {
			return output, nil
		}
		return types.BoolNull(), fmt.Errorf(
			"reconcile %q: API returned %t but field was not set in config and is not computed",
			field, output.ValueBool(),
		)
	default:
		return types.BoolNull(), nil
	}
}

// reconcileInt64 reconciles a plan/state int64 with an API response int64.
func reconcileInt64(field string, input, output types.Int64, computed bool) (types.Int64, error) {
	inputEmpty := input.IsNull() || input.IsUnknown()
	outputEmpty := output.IsNull() || output.IsUnknown()

	switch {
	case !inputEmpty && !outputEmpty:
		return output, nil
	case !inputEmpty && outputEmpty:
		return input, nil
	case inputEmpty && !outputEmpty:
		if computed {
			return output, nil
		}
		return types.Int64Null(), fmt.Errorf(
			"reconcile %q: API returned %d but field was not set in config and is not computed",
			field, output.ValueInt64(),
		)
	default:
		return types.Int64Null(), nil
	}
}
