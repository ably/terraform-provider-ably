package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveRetryInt(t *testing.T) {
	const attrName, envVar = "retry_max", "ABLY_RETRY_MAX"

	tests := []struct {
		name      string
		attr      types.Int64
		env       string // "" means unset
		wantValue int
		wantOk    bool
		errSubstr string // "" means no error expected
	}{
		{name: "attribute set", attr: types.Int64Value(3), wantValue: 3, wantOk: true},
		{name: "attribute zero", attr: types.Int64Value(0), wantValue: 0, wantOk: true},
		{name: "unset falls through", attr: types.Int64Null(), wantOk: false},
		{name: "unknown falls through", attr: types.Int64Unknown(), wantOk: false},
		{name: "env fallback", attr: types.Int64Null(), env: "5", wantValue: 5, wantOk: true},
		{name: "attribute wins over env", attr: types.Int64Value(2), env: "9", wantValue: 2, wantOk: true},
		{name: "non-integer env", attr: types.Int64Null(), env: "nope", errSubstr: envVar},
		{name: "negative attribute names attribute", attr: types.Int64Value(-1), errSubstr: attrName},
		{name: "negative env names env var", attr: types.Int64Null(), env: "-1", errSubstr: envVar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.env != "" {
				t.Setenv(envVar, tt.env)
			}

			value, ok, err := resolveRetryInt(tt.attr, attrName, envVar)

			if tt.errSubstr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error to mention %q, got %q", tt.errSubstr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok != tt.wantOk {
				t.Fatalf("ok: got %v, want %v", ok, tt.wantOk)
			}
			if ok && value != tt.wantValue {
				t.Fatalf("value: got %d, want %d", value, tt.wantValue)
			}
		})
	}
}
