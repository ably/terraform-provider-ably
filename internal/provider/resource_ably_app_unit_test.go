// Package provider implements the Ably provider for Terraform
package provider

import (
	"testing"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestBuildAppState_FCMProjectID pins the INF-8172 behaviour: the Control API
// returns "fcmProjectId": "" for apps last written by the pre-1.0 provider, and
// an unset fcm_project_id must come back as null from Create, Update and Read
// alike, including when the prior state itself still holds the "" that v1.0.0
// wrote.
func TestBuildAppState_FCMProjectID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input types.String // plan on Create/Update, prior state on Read
		api   *string
		read  bool
		want  types.String
	}{
		{name: "create/update, unset in plan, API returns empty string", input: types.StringNull(), api: ptr(""), want: types.StringNull()},
		{name: "create/update, unset in plan, API omits the field", input: types.StringNull(), api: nil, want: types.StringNull()},
		{name: "read, unset in prior state, API returns empty string", input: types.StringNull(), api: ptr(""), read: true, want: types.StringNull()},
		{name: "read, v1.0.0 state holds empty string, API returns empty string", input: types.StringValue(""), api: ptr(""), read: true, want: types.StringNull()},
		{name: "read, v1.0.0 state holds empty string, API omits the field", input: types.StringValue(""), api: nil, read: true, want: types.StringNull()},
		{name: "create/update, set in plan, API echoes it", input: types.StringValue("project-a"), api: ptr("project-a"), want: types.StringValue("project-a")},
		{name: "read, set in prior state, API returns it", input: types.StringValue("project-a"), api: ptr("project-a"), read: true, want: types.StringValue("project-a")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var diags diag.Diagnostics
			rc := newReconciler(&diags)
			if tc.read {
				rc = rc.forRead()
			}
			input := AblyAppState{
				Name:         types.StringValue("my-app"),
				FcmProjectId: tc.input,
			}
			api := control.AppResponse{
				ID:           "app-1",
				AccountID:    "acct-1",
				Name:         "my-app",
				Status:       "enabled",
				FCMProjectID: tc.api,
			}
			got := buildAppState(rc, input, api)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if !got.FcmProjectId.Equal(tc.want) {
				t.Fatalf("fcm_project_id = %s, want %s", got.FcmProjectId, tc.want)
			}
		})
	}
}
