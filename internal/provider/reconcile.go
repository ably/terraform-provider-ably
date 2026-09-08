package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Reconciliation decides, field by field, what to store in state after an API
// call by comparing the input (the plan on Create/Update, the prior state on
// Read) with the output (the API response). It resolves the four-way matrix of
// input/output emptiness:
//
//  1. Input non-empty, Output non-empty → output (the API is authoritative)
//  2. Input non-empty, Output empty     → input  (write-only / sensitive field)
//  3. Input empty,     Output non-empty:
//     - computed=true  → output (server-owned or defaulted field)
//     - computed=false → error  (unexpected value for an optional-only field)
//  4. Input empty,     Output empty     → null
//
// "Empty" means null or unknown, the zero-length string for strings, and zero
// length for slices and maps. false and 0 are real values, not empty.
//
// The matrix is implemented once, in reconcile; the typed helpers above it only
// decide emptiness for their kind.

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

// record folds a reconciliation error into the accumulated diagnostics.
func (r *reconciler) record(err error) {
	if err != nil {
		r.diags.AddError("State reconciliation error", err.Error())
	}
}

// Go does not allow generic methods, so the reconciler-aware entry points are
// free functions that take the reconciler as their first argument.

// rcVal reconciles a scalar (string, bool or int64) field.
func rcVal[V scalarValue](rc *reconciler, field string, input, output V, computed bool) V {
	v, err := reconcileValue(field, input, output, computed || rc.reading)
	rc.record(err)
	return v
}

// rcSlice reconciles a slice field.
func rcSlice[E any](rc *reconciler, field string, input, output []E, computed bool) []E {
	v, err := reconcileSlice(field, input, output, computed || rc.reading)
	rc.record(err)
	return v
}

// rcMapSet reconciles a map[string]types.Set field.
func rcMapSet(rc *reconciler, field string, input, output map[string]types.Set, computed bool) map[string]types.Set {
	v, err := reconcileMapSet(field, input, output, computed || rc.reading)
	rc.record(err)
	return v
}

// planTarget extracts the typed plan target of a rule, or a zero-value one when
// the plan carries none (import) or one of a different type. Every field of the
// zero value is null (or a nil slice), which reconciliation treats as empty, so
// callers can read fields directly without nil checks.
func planTarget[T any](target any) *T {
	if p, ok := target.(*T); ok && p != nil {
		return p
	}
	return new(T)
}

// scalarValue is the set of framework scalar types reconcileValue accepts.
type scalarValue interface {
	types.String | types.Bool | types.Int64
	attr.Value
}

// valueEmpty reports whether a scalar is empty: null, unknown, or (for
// strings) the zero-length string.
func valueEmpty[V scalarValue](v V) bool {
	if v.IsNull() || v.IsUnknown() {
		return true
	}
	s, ok := any(v).(types.String)
	return ok && s.ValueString() == ""
}

// reconcileValue reconciles a plan/state scalar with an API response scalar.
func reconcileValue[V scalarValue](field string, input, output V, computed bool) (V, error) {
	return reconcile(field, input, output, valueEmpty(input), valueEmpty(output), computed)
}

// reconcileSlice reconciles a plan/state slice with an API response slice.
func reconcileSlice[E any](field string, input, output []E, computed bool) ([]E, error) {
	return reconcile(field, input, output, len(input) == 0, len(output) == 0, computed)
}

// reconcileMapSet reconciles a plan/state map[string]types.Set with an API
// response map.
func reconcileMapSet(field string, input, output map[string]types.Set, computed bool) (map[string]types.Set, error) {
	return reconcile(field, input, output, len(input) == 0, len(output) == 0, computed)
}

// reconcile is the single implementation of the four-way matrix documented at
// the top of this file. Emptiness is decided by the typed callers; the zero
// value of T is null for framework scalars and nil for collections.
//
// The error deliberately omits the returned value: fields can carry secrets
// (keys, tokens, header values) that must not leak into diagnostics or logs.
func reconcile[T any](field string, input, output T, inputEmpty, outputEmpty, computed bool) (T, error) {
	var null T
	switch {
	case !inputEmpty && !outputEmpty:
		return output, nil
	case !inputEmpty && outputEmpty:
		return input, nil
	case inputEmpty && !outputEmpty:
		if computed {
			return output, nil
		}
		return null, fmt.Errorf(
			"reconcile %q: API returned a value but field was not set in config and is not computed",
			field,
		)
	default:
		return null, nil
	}
}
