package provider

import (
	"context"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceNamespace struct {
	p *provider
}

// Get Namespace Resource schema
func (r resourceNamespace) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"app_id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The application ID.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.RequiresReplace(),
				},
			},
			"id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The namespace or channel name that the channel rule will apply to.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.RequiresReplace(),
				},
			},
			"authenticated": {
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				Description: "Require clients to be authenticated to use channels in this namespace.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.BoolValue(false)),
				},
			},
			"persisted": {
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				Description: "If true, messages will be stored for 24 hours.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.BoolValue(false)),
				},
			},
			"persist_last": {
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				Description: "If true, the last message on each channel will persist for 365 days.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.BoolValue(false)),
				},
			},
			"push_enabled": {
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				Description: "If true, publishing messages with a push payload in the extras field is permitted.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.BoolValue(false)),
				},
			},
			"tls_only": {
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				Description: "If true, only clients that are connected using TLS will be permitted to subscribe.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.BoolValue(false)),
				},
			},
			"expose_timeserial": {
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				Description: "If true, messages received on a channel will contain a unique timeserial that can be referenced by later messages for use with message interactions.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.BoolValue(false)),
				},
			},
			"batching_enabled": {
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				Description: "If true, channels within this namespace will start batching inbound messages instead of sending them out immediately to subscribers as per the configured policy.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.BoolValue(false)),
				},
			},
			"batching_policy": {
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "When configured, sets the policy for message batching.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.StringValue("")),
				},
			},
			"batching_interval": {
				Type:        types.Int64Type,
				Optional:    true,
				Computed:    true,
				Description: "When configured, sets the maximium batching interval in the channel.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.Int64Null()),
				},
			},
		},
		MarkdownDescription: "The ably_namespace resource allows you to manage namespaces for channel rules in Ably. Read more in the Ably documentation: https://ably.com/docs/general/channel-rules-namespaces.",
	}, nil
}

// Create a new resource
func (r resourceNamespace) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
	// Checks whether the provider and API Client are configured. If they are not, the provider responds with an error.
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply",
		)
		return
	}

	// Gets plan values
	var plan AblyNamespace
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generates an API request body from the plan values
	namespace_values := ably_control_go.Namespace{
		ID:               plan.ID.ValueString(),
		Authenticated:    plan.Authenticated.ValueBool(),
		Persisted:        plan.Persisted.ValueBool(),
		PersistLast:      plan.PersistLast.ValueBool(),
		PushEnabled:      plan.PushEnabled.ValueBool(),
		TlsOnly:          plan.TlsOnly.ValueBool(),
		ExposeTimeserial: plan.ExposeTimeserial.ValueBool(),
	}

	if plan.BatchingEnabled.ValueBool() {
		namespace_values.BatchingEnabled = true
		namespace_values.BatchingPolicy = plan.BatchingPolicy.ValueString()
		namespace_values.BatchingInterval = ably_control_go.BatchingInterval(int(plan.BatchingInterval.ValueInt64()))
	}

	// Creates a new Ably namespace by invoking the CreateNamespace function from the Client Library
	ably_namespace, err := r.p.client.CreateNamespace(plan.AppID.ValueString(), &namespace_values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Handle the pointer gracefully
	batchingInterval := types.Int64Null()
	if ably_namespace.BatchingInterval != nil {
		batchingInterval = types.Int64Value(int64(*ably_namespace.BatchingInterval))
	}

	// Maps response body to resource schema attributes.
	resp_apps := AblyNamespace{
		AppID:            types.StringValue(plan.AppID.ValueString()),
		ID:               types.StringValue(ably_namespace.ID),
		Authenticated:    types.BoolValue(ably_namespace.Authenticated),
		Persisted:        types.BoolValue(ably_namespace.Persisted),
		PersistLast:      types.BoolValue(ably_namespace.PersistLast),
		PushEnabled:      types.BoolValue(ably_namespace.PushEnabled),
		TlsOnly:          types.BoolValue(ably_namespace.TlsOnly),
		ExposeTimeserial: types.BoolValue(namespace_values.ExposeTimeserial),
		BatchingEnabled:  types.BoolValue(ably_namespace.BatchingEnabled),
		BatchingPolicy:   types.StringValue(ably_namespace.BatchingPolicy),
		BatchingInterval: batchingInterval,
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, resp_apps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceNamespace) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_namespace"
}

// Read resource
func (r resourceNamespace) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyNamespace
	found := false
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID and namespace ID value for the resource
	app_id := state.AppID.ValueString()
	namespace_id := state.ID.ValueString()

	// Fetches all Ably Namespaces in the app. The function invokes the Client Library Namespaces() method.
	// NOTE: Control API & Client Lib do not currently support fetching single namespace given namespace id
	namespaces, err := r.p.client.Namespaces(app_id)
	if err != nil {
		if is_404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Loops through namespaces and if id matches, sets state.
	for _, v := range namespaces {
		if v.ID == namespace_id {
			// Handle the pointer gracefully
			batchingInterval := types.Int64Null()
			if v.BatchingInterval != nil {
				batchingInterval = types.Int64Value(int64(*v.BatchingInterval))
			}

			resp_namespaces := AblyNamespace{
				AppID:            types.StringValue(app_id),
				ID:               types.StringValue(namespace_id),
				Authenticated:    types.BoolValue(v.Authenticated),
				Persisted:        types.BoolValue(v.Persisted),
				PersistLast:      types.BoolValue(v.PersistLast),
				PushEnabled:      types.BoolValue(v.PushEnabled),
				TlsOnly:          types.BoolValue(v.TlsOnly),
				ExposeTimeserial: types.BoolValue(v.ExposeTimeserial),
				BatchingEnabled:  types.BoolValue(v.BatchingEnabled),
				BatchingPolicy:   types.StringValue(v.BatchingPolicy),
				BatchingInterval: batchingInterval,
			}
			// Sets state to namespace values.
			diags = resp.State.Set(ctx, &resp_namespaces)
			found = true

			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
	}
}

// Update resource
func (r resourceNamespace) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
	// Get plan values
	var plan AblyNamespace
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the app ID and ID
	app_id := plan.AppID.ValueString()
	namespace_id := plan.ID.ValueString()

	// Instantiates struct of type ably_control_go.Namespace and sets values to output of plan
	namespace_values := ably_control_go.Namespace{
		ID:               namespace_id,
		Authenticated:    plan.Authenticated.ValueBool(),
		Persisted:        plan.Persisted.ValueBool(),
		PersistLast:      plan.PersistLast.ValueBool(),
		PushEnabled:      plan.PushEnabled.ValueBool(),
		TlsOnly:          plan.TlsOnly.ValueBool(),
		ExposeTimeserial: plan.ExposeTimeserial.ValueBool(),
	}

	if plan.BatchingEnabled.ValueBool() {
		namespace_values.BatchingEnabled = true
		namespace_values.BatchingPolicy = plan.BatchingPolicy.ValueString()
		namespace_values.BatchingInterval = ably_control_go.BatchingInterval(int(plan.BatchingInterval.ValueInt64()))
	}

	// Updates an Ably Namespace. The function invokes the Client Library UpdateNamespace method.
	ably_namespace, err := r.p.client.UpdateNamespace(app_id, &namespace_values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Handle the pointer gracefully
	batchingInterval := types.Int64Null()
	if ably_namespace.BatchingInterval != nil {
		batchingInterval = types.Int64Value(int64(*ably_namespace.BatchingInterval))
	}

	resp_namespaces := AblyNamespace{
		AppID:            types.StringValue(app_id),
		ID:               types.StringValue(ably_namespace.ID),
		Authenticated:    types.BoolValue(ably_namespace.Authenticated),
		Persisted:        types.BoolValue(ably_namespace.Persisted),
		PersistLast:      types.BoolValue(ably_namespace.PersistLast),
		PushEnabled:      types.BoolValue(ably_namespace.PushEnabled),
		TlsOnly:          types.BoolValue(ably_namespace.TlsOnly),
		ExposeTimeserial: types.BoolValue(ably_namespace.ExposeTimeserial),
		BatchingEnabled:  types.BoolValue(ably_namespace.BatchingEnabled),
		BatchingPolicy:   types.StringValue(ably_namespace.BatchingPolicy),
		BatchingInterval: batchingInterval,
	}

	// Sets state to new namespace.
	diags = resp.State.Set(ctx, resp_namespaces)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceNamespace) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	// Get current state
	var state AblyNamespace
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	app_id := state.AppID.ValueString()
	namespace_id := state.ID.ValueString()

	err := r.p.client.DeleteNamespace(app_id, namespace_id)
	if err != nil {
		if is_404(err) {
			resp.Diagnostics.AddWarning(
				"Resource does not exist",
				"Resource does not exist, it may have already been deleted: "+err.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error deleting Resource",
				"Could not delete resource, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// Import resource
func (r resourceNamespace) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")

}
