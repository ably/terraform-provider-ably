package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// reconcileCase is one row of the four-way matrix for a scalar kind. want is
// ignored when wantErr is set: an error always yields the null value.
type reconcileCase[V scalarValue] struct {
	name     string
	input    V
	output   V
	computed bool
	want     V
	wantErr  bool
}

func runReconcileCases[V scalarValue](t *testing.T, cases []reconcileCase[V]) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := reconcileValue("my_field", tc.input, tc.output, tc.computed)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error for unexpected API value on non-computed field")
				}
				if !got.IsNull() {
					t.Fatalf("expected null alongside the error, got %s", got)
				}
				// The API value may be a secret; the error must never contain it.
				if leaked := strings.Trim(tc.output.String(), `"`); strings.Contains(err.Error(), leaked) {
					t.Fatalf("error must not leak the raw API value %q, got %q", leaked, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !got.Equal(tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestReconcileValue_String(t *testing.T) {
	t.Parallel()
	runReconcileCases(t, []reconcileCase[types.String]{
		{name: "both non-empty: output wins", input: types.StringValue("plan"), output: types.StringValue("api"), want: types.StringValue("api")},
		{name: "input non-empty, output null: echo input", input: types.StringValue("secret"), output: types.StringNull(), want: types.StringValue("secret")},
		{name: "input non-empty, output empty string: echo input", input: types.StringValue("secret"), output: types.StringValue(""), want: types.StringValue("secret")},
		{name: "input null, output non-empty, computed: output", input: types.StringNull(), output: types.StringValue("server-id"), computed: true, want: types.StringValue("server-id")},
		{name: "input unknown, output non-empty, computed: output", input: types.StringUnknown(), output: types.StringValue("resolved"), computed: true, want: types.StringValue("resolved")},
		{name: "input null, output non-empty, not computed: error", input: types.StringNull(), output: types.StringValue("surprise"), wantErr: true},
		{name: "input empty string, output non-empty, not computed: error", input: types.StringValue(""), output: types.StringValue("surprise"), wantErr: true},
		{name: "both null: null", input: types.StringNull(), output: types.StringNull(), want: types.StringNull()},
		{name: "both empty string: null", input: types.StringValue(""), output: types.StringValue(""), want: types.StringNull()},
	})
}

func TestReconcileValue_Bool(t *testing.T) {
	t.Parallel()
	runReconcileCases(t, []reconcileCase[types.Bool]{
		{name: "both non-empty: output wins", input: types.BoolValue(true), output: types.BoolValue(false), want: types.BoolValue(false)},
		{name: "input non-empty, output null: echo input", input: types.BoolValue(true), output: types.BoolNull(), want: types.BoolValue(true)},
		{name: "false is a value, not empty", input: types.BoolValue(false), output: types.BoolNull(), want: types.BoolValue(false)},
		{name: "input null, output non-empty, computed: output", input: types.BoolNull(), output: types.BoolValue(true), computed: true, want: types.BoolValue(true)},
		{name: "input null, output non-empty, not computed: error", input: types.BoolNull(), output: types.BoolValue(true), wantErr: true},
		{name: "both null: null", input: types.BoolNull(), output: types.BoolNull(), want: types.BoolNull()},
	})
}

func TestReconcileValue_Int64(t *testing.T) {
	t.Parallel()
	runReconcileCases(t, []reconcileCase[types.Int64]{
		{name: "both non-empty: output wins", input: types.Int64Value(10), output: types.Int64Value(20), want: types.Int64Value(20)},
		{name: "input non-empty, output null: echo input", input: types.Int64Value(42), output: types.Int64Null(), want: types.Int64Value(42)},
		{name: "zero is a value, not empty", input: types.Int64Value(0), output: types.Int64Null(), want: types.Int64Value(0)},
		{name: "input null, output non-empty, computed: output", input: types.Int64Null(), output: types.Int64Value(99), computed: true, want: types.Int64Value(99)},
		{name: "input null, output non-empty, not computed: error", input: types.Int64Null(), output: types.Int64Value(99), wantErr: true},
		{name: "both null: null", input: types.Int64Null(), output: types.Int64Null(), want: types.Int64Null()},
	})
}

func TestReconcileSlice(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    []string
		output   []string
		computed bool
		want     []string
		wantErr  bool
	}{
		{name: "both non-empty: output wins", input: []string{"a"}, output: []string{"b"}, want: []string{"b"}},
		{name: "input non-empty, output empty: echo input", input: []string{"a"}, output: nil, want: []string{"a"}},
		{name: "input empty, output non-empty, computed: output", input: nil, output: []string{"x"}, computed: true, want: []string{"x"}},
		{name: "input empty, output non-empty, not computed: error", input: nil, output: []string{"x"}, wantErr: true},
		{name: "both empty: nil", input: nil, output: nil, want: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := reconcileSlice("my_field", tc.input, tc.output, tc.computed)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error for unexpected API value on non-computed field")
				}
				if got != nil {
					t.Fatalf("expected nil alongside the error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("expected %v, got %v", tc.want, got)
				}
			}
		})
	}
}

func TestReconcileMapSet(t *testing.T) {
	t.Parallel()
	a := map[string]types.Set{"a": types.SetNull(types.StringType)}
	b := map[string]types.Set{"b": types.SetNull(types.StringType)}
	cases := []struct {
		name     string
		input    map[string]types.Set
		output   map[string]types.Set
		computed bool
		wantKey  string // "" means a nil result is expected
		wantErr  bool
	}{
		{name: "both non-empty: output wins", input: a, output: b, wantKey: "b"},
		{name: "input non-empty, output empty: echo input", input: a, output: nil, wantKey: "a"},
		{name: "input empty, output non-empty, computed: output", input: nil, output: b, computed: true, wantKey: "b"},
		{name: "input empty, output non-empty, not computed: error", input: nil, output: b, wantErr: true},
		{name: "both empty: nil", input: nil, output: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := reconcileMapSet("my_field", tc.input, tc.output, tc.computed)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error for unexpected API value on non-computed field")
				}
				if got != nil {
					t.Fatalf("expected nil alongside the error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantKey == "" {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if _, ok := got[tc.wantKey]; !ok || len(got) != 1 {
				t.Fatalf("expected a map with only key %q, got %v", tc.wantKey, got)
			}
		})
	}
}

// TestPlanTarget covers the import path (typed nil) and a type mismatch: both
// must yield a usable zero-value target whose fields reconcile as empty.
func TestPlanTarget(t *testing.T) {
	t.Parallel()

	set := &AblyRuleTargetHTTP{Url: types.StringValue("https://example.com")}
	if got := planTarget[AblyRuleTargetHTTP](set); got != set {
		t.Fatal("expected the plan target to be returned as-is")
	}

	var typedNil *AblyRuleTargetHTTP
	if got := planTarget[AblyRuleTargetHTTP](typedNil); got == nil || !got.Url.IsNull() {
		t.Fatalf("expected a zero-value target for a typed nil, got %v", got)
	}

	if got := planTarget[AblyRuleTargetHTTP](&AblyRuleTargetZapier{}); got == nil || !got.Url.IsNull() {
		t.Fatalf("expected a zero-value target for a type mismatch, got %v", got)
	}

	if got := planTarget[AblyRuleTargetHTTP](nil); got == nil || got.Headers != nil {
		t.Fatalf("expected a zero-value target for a nil interface, got %v", got)
	}
}
