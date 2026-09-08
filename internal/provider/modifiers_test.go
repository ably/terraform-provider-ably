package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The default* modifiers must apply their default only when Terraform has no
// prior value for the attribute. On an update they keep the prior state value,
// because the CRUD code serialises plan values into requests and re-sending
// the default would overwrite a server-side or out-of-band value on every
// unrelated change.
func TestDefaultBoolAttribute(t *testing.T) {
	t.Parallel()

	def := types.BoolValue(true)
	cases := []struct {
		name   string
		config types.Bool
		state  types.Bool
		plan   types.Bool
		want   types.Bool
	}{
		{name: "create, unconfigured: default", config: types.BoolNull(), state: types.BoolNull(), plan: types.BoolUnknown(), want: def},
		{name: "update, unconfigured, unknown plan: prior state", config: types.BoolNull(), state: types.BoolValue(false), plan: types.BoolUnknown(), want: types.BoolValue(false)},
		{name: "no-op update, unconfigured: plan already carries state", config: types.BoolNull(), state: types.BoolValue(false), plan: types.BoolValue(false), want: types.BoolValue(false)},
		{name: "configured: config wins over default and state", config: types.BoolValue(false), state: types.BoolValue(true), plan: types.BoolValue(false), want: types.BoolValue(false)},
		{name: "unknown config: left unknown", config: types.BoolUnknown(), state: types.BoolValue(true), plan: types.BoolUnknown(), want: types.BoolUnknown()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := planmodifier.BoolRequest{ConfigValue: tc.config, StateValue: tc.state, PlanValue: tc.plan}
			resp := planmodifier.BoolResponse{PlanValue: tc.plan}
			DefaultBoolAttribute(def).PlanModifyBool(context.Background(), req, &resp)
			if !resp.PlanValue.Equal(tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, resp.PlanValue)
			}
		})
	}
}

func TestDefaultStringAttribute(t *testing.T) {
	t.Parallel()

	def := types.StringValue("single")
	cases := []struct {
		name   string
		config types.String
		state  types.String
		plan   types.String
		want   types.String
	}{
		{name: "create, unconfigured: default", config: types.StringNull(), state: types.StringNull(), plan: types.StringUnknown(), want: def},
		{name: "update, unconfigured, unknown plan: prior state", config: types.StringNull(), state: types.StringValue("batch"), plan: types.StringUnknown(), want: types.StringValue("batch")},
		{name: "configured: config wins", config: types.StringValue("batch"), state: types.StringValue("single"), plan: types.StringValue("batch"), want: types.StringValue("batch")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := planmodifier.StringRequest{ConfigValue: tc.config, StateValue: tc.state, PlanValue: tc.plan}
			resp := planmodifier.StringResponse{PlanValue: tc.plan}
			DefaultStringAttribute(def).PlanModifyString(context.Background(), req, &resp)
			if !resp.PlanValue.Equal(tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, resp.PlanValue)
			}
		})
	}
}

func TestDefaultInt64Attribute(t *testing.T) {
	t.Parallel()

	def := types.Int64Value(0)
	cases := []struct {
		name   string
		config types.Int64
		state  types.Int64
		plan   types.Int64
		want   types.Int64
	}{
		{name: "create, unconfigured: default", config: types.Int64Null(), state: types.Int64Null(), plan: types.Int64Unknown(), want: def},
		{name: "update, unconfigured, unknown plan: prior state", config: types.Int64Null(), state: types.Int64Value(1), plan: types.Int64Unknown(), want: types.Int64Value(1)},
		{name: "configured: config wins", config: types.Int64Value(7), state: types.Int64Value(1), plan: types.Int64Value(7), want: types.Int64Value(7)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := planmodifier.Int64Request{ConfigValue: tc.config, StateValue: tc.state, PlanValue: tc.plan}
			resp := planmodifier.Int64Response{PlanValue: tc.plan}
			DefaultInt64Attribute(def).PlanModifyInt64(context.Background(), req, &resp)
			if !resp.PlanValue.Equal(tc.want) {
				t.Fatalf("expected %s, got %s", tc.want, resp.PlanValue)
			}
		})
	}
}
