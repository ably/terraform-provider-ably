package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- reconcileString ---

func TestReconcileString_BothNonEmpty(t *testing.T) {
	t.Parallel()
	got, err := reconcileString("f", types.StringValue("plan"), types.StringValue("api"), false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueString() != "api" {
		t.Fatalf("expected output value 'api', got %q", got.ValueString())
	}
}

func TestReconcileString_InputNonEmpty_OutputNull(t *testing.T) {
	t.Parallel()
	got, err := reconcileString("f", types.StringValue("secret"), types.StringNull(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueString() != "secret" {
		t.Fatalf("expected echoed input 'secret', got %q", got.ValueString())
	}
}

func TestReconcileString_InputNonEmpty_OutputEmptyString(t *testing.T) {
	t.Parallel()
	got, err := reconcileString("f", types.StringValue("secret"), types.StringValue(""), false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueString() != "secret" {
		t.Fatalf("expected echoed input 'secret', got %q", got.ValueString())
	}
}

func TestReconcileString_InputNull_OutputNonEmpty_Computed(t *testing.T) {
	t.Parallel()
	got, err := reconcileString("f", types.StringNull(), types.StringValue("server-id"), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueString() != "server-id" {
		t.Fatalf("expected output value 'server-id', got %q", got.ValueString())
	}
}

func TestReconcileString_InputNull_OutputNonEmpty_NotComputed(t *testing.T) {
	t.Parallel()
	_, err := reconcileString("my_field", types.StringNull(), types.StringValue("surprise"), false)
	if err == nil {
		t.Fatal("expected error for unexpected API value on non-computed field")
	}
}

func TestReconcileString_InputEmpty_OutputNonEmpty_NotComputed(t *testing.T) {
	t.Parallel()
	_, err := reconcileString("my_field", types.StringValue(""), types.StringValue("surprise"), false)
	if err == nil {
		t.Fatal("expected error for unexpected API value on non-computed field (empty string input)")
	}
}

func TestReconcileString_BothNull(t *testing.T) {
	t.Parallel()
	got, err := reconcileString("f", types.StringNull(), types.StringNull(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsNull() {
		t.Fatalf("expected null, got %q", got.ValueString())
	}
}

func TestReconcileString_BothEmpty(t *testing.T) {
	t.Parallel()
	got, err := reconcileString("f", types.StringValue(""), types.StringValue(""), false)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsNull() {
		t.Fatalf("expected null, got %q", got.ValueString())
	}
}

func TestReconcileString_InputUnknown_OutputNonEmpty_Computed(t *testing.T) {
	t.Parallel()
	got, err := reconcileString("f", types.StringUnknown(), types.StringValue("resolved"), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueString() != "resolved" {
		t.Fatalf("expected 'resolved', got %q", got.ValueString())
	}
}

// --- reconcileBool ---

func TestReconcileBool_BothNonEmpty(t *testing.T) {
	t.Parallel()
	got, err := reconcileBool("f", types.BoolValue(true), types.BoolValue(false), false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueBool() != false {
		t.Fatal("expected output value false")
	}
}

func TestReconcileBool_InputNonEmpty_OutputNull(t *testing.T) {
	t.Parallel()
	got, err := reconcileBool("f", types.BoolValue(true), types.BoolNull(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueBool() != true {
		t.Fatal("expected echoed input true")
	}
}

func TestReconcileBool_InputNull_OutputNonEmpty_Computed(t *testing.T) {
	t.Parallel()
	got, err := reconcileBool("f", types.BoolNull(), types.BoolValue(true), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueBool() != true {
		t.Fatal("expected output value true")
	}
}

func TestReconcileBool_InputNull_OutputNonEmpty_NotComputed(t *testing.T) {
	t.Parallel()
	_, err := reconcileBool("my_field", types.BoolNull(), types.BoolValue(true), false)
	if err == nil {
		t.Fatal("expected error for unexpected API value on non-computed field")
	}
}

func TestReconcileBool_BothNull(t *testing.T) {
	t.Parallel()
	got, err := reconcileBool("f", types.BoolNull(), types.BoolNull(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsNull() {
		t.Fatal("expected null")
	}
}

func TestReconcileBool_FalseIsNotEmpty(t *testing.T) {
	t.Parallel()
	// false is a valid value, not "empty"
	got, err := reconcileBool("f", types.BoolValue(false), types.BoolNull(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueBool() != false {
		t.Fatal("expected echoed input false")
	}
	if got.IsNull() {
		t.Fatal("false should not be treated as empty")
	}
}

// --- reconcileInt64 ---

func TestReconcileInt64_BothNonEmpty(t *testing.T) {
	t.Parallel()
	got, err := reconcileInt64("f", types.Int64Value(10), types.Int64Value(20), false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueInt64() != 20 {
		t.Fatalf("expected output value 20, got %d", got.ValueInt64())
	}
}

func TestReconcileInt64_InputNonEmpty_OutputNull(t *testing.T) {
	t.Parallel()
	got, err := reconcileInt64("f", types.Int64Value(42), types.Int64Null(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueInt64() != 42 {
		t.Fatalf("expected echoed input 42, got %d", got.ValueInt64())
	}
}

func TestReconcileInt64_InputNull_OutputNonEmpty_Computed(t *testing.T) {
	t.Parallel()
	got, err := reconcileInt64("f", types.Int64Null(), types.Int64Value(99), true)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueInt64() != 99 {
		t.Fatalf("expected output value 99, got %d", got.ValueInt64())
	}
}

func TestReconcileInt64_InputNull_OutputNonEmpty_NotComputed(t *testing.T) {
	t.Parallel()
	_, err := reconcileInt64("my_field", types.Int64Null(), types.Int64Value(99), false)
	if err == nil {
		t.Fatal("expected error for unexpected API value on non-computed field")
	}
}

func TestReconcileInt64_BothNull(t *testing.T) {
	t.Parallel()
	got, err := reconcileInt64("f", types.Int64Null(), types.Int64Null(), false)
	if err != nil {
		t.Fatal(err)
	}
	if !got.IsNull() {
		t.Fatal("expected null")
	}
}

func TestReconcileInt64_ZeroIsNotEmpty(t *testing.T) {
	t.Parallel()
	// 0 is a valid value, not "empty"
	got, err := reconcileInt64("f", types.Int64Value(0), types.Int64Null(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got.ValueInt64() != 0 {
		t.Fatalf("expected echoed input 0, got %d", got.ValueInt64())
	}
	if got.IsNull() {
		t.Fatal("0 should not be treated as empty")
	}
}

// --- reconcileSlice ---

func TestReconcileSlice_BothNonEmpty(t *testing.T) {
	t.Parallel()
	got, err := reconcileSlice("f", []string{"a"}, []string{"b"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("expected output [b], got %v", got)
	}
}

func TestReconcileSlice_InputNonEmpty_OutputEmpty(t *testing.T) {
	t.Parallel()
	got, err := reconcileSlice("f", []string{"a"}, []string(nil), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("expected echoed input [a], got %v", got)
	}
}

func TestReconcileSlice_InputEmpty_OutputNonEmpty_Computed(t *testing.T) {
	t.Parallel()
	got, err := reconcileSlice("f", []string(nil), []string{"x"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "x" {
		t.Fatalf("expected output [x], got %v", got)
	}
}

func TestReconcileSlice_InputEmpty_OutputNonEmpty_NotComputed(t *testing.T) {
	t.Parallel()
	_, err := reconcileSlice("my_field", []string(nil), []string{"x"}, false)
	if err == nil {
		t.Fatal("expected error for unexpected API value on non-computed field")
	}
}

func TestReconcileSlice_BothEmpty(t *testing.T) {
	t.Parallel()
	got, err := reconcileSlice("f", []string(nil), []string(nil), false)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

// --- reconcileMapSet ---

func TestReconcileMapSet_BothNonEmpty(t *testing.T) {
	t.Parallel()
	input := map[string]types.Set{"a": types.SetNull(types.StringType)}
	output := map[string]types.Set{"b": types.SetNull(types.StringType)}
	got, err := reconcileMapSet("f", input, output, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["b"]; !ok {
		t.Fatal("expected output map with key 'b'")
	}
}

func TestReconcileMapSet_InputNonEmpty_OutputEmpty(t *testing.T) {
	t.Parallel()
	input := map[string]types.Set{"a": types.SetNull(types.StringType)}
	got, err := reconcileMapSet("f", input, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got["a"]; !ok {
		t.Fatal("expected echoed input map with key 'a'")
	}
}

func TestReconcileMapSet_InputEmpty_OutputNonEmpty_NotComputed(t *testing.T) {
	t.Parallel()
	output := map[string]types.Set{"b": types.SetNull(types.StringType)}
	_, err := reconcileMapSet("my_field", nil, output, false)
	if err == nil {
		t.Fatal("expected error for unexpected API value on non-computed field")
	}
}

func TestReconcileMapSet_BothEmpty(t *testing.T) {
	t.Parallel()
	got, err := reconcileMapSet("f", nil, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil")
	}
}
