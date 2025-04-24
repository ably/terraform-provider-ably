// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceNamespace struct {
	p *AblyProvider
}

var _ resource.Resource = &ResourceNamespace{}
var _ resource.ResourceWithImportState = &ResourceNamespace{}

// Schema defines the schema for the resource.
func (r ResourceNamespace) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app_id": schema.StringAttribute{
				Required:    true,
				Description: "The application ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The namespace or channel name that the channel rule will apply to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authenticated": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Require clients to be authenticated to use channels in this namespace.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"persisted": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, messages will be stored for 24 hours.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"persist_last": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, the last message on each channel will persist for 365 days.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"push_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, publishing messages with a push payload in the extras field is permitted.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"tls_only": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, only clients that are connected using TLS will be permitted to subscribe.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"expose_timeserial": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, messages received on a channel will contain a unique timeserial that can be referenced by later messages for use with message interactions.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"batching_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, channels within this namespace will start batching inbound messages instead of sending them out immediately to subscribers as per the configured policy.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"batching_interval": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "When configured, sets the maximium batching interval in the channel.",
				PlanModifiers: []planmodifier.Int64{
					DefaultInt64Attribute(types.Int64Null()),
				},
			},
			"conflation_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, enables conflation for channels within this namespace. Conflation reduces the number of messages sent to subscribers by combining multiple messages into a single message.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"conflation_interval": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The interval in milliseconds at which messages are conflated. This determines how frequently messages are combined into a single message.",
				PlanModifiers: []planmodifier.Int64{
					DefaultInt64Attribute(types.Int64Null()),
				},
			},
			"conflation_key": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The key used to determine which messages should be conflated. Messages with the same conflation key will be combined into a single message.",
				PlanModifiers: []planmodifier.String{
					DefaultStringAttribute(types.StringNull()),
				},
			},
		},
		MarkdownDescription: "The Ably namespace resource allows you to manage namespaces for channel rules in Ably. Read more in the Ably documentation: https://ably.com/docs/general/channel-rules-namespaces.",
	}
}

// Create creates a new resource.
func (r ResourceNamespace) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	namespaceValues := control.Namespace{
		ID:               plan.ID.ValueString(),
		Authenticated:    plan.Authenticated.ValueBool(),
		Persisted:        plan.Persisted.ValueBool(),
		PersistLast:      plan.PersistLast.ValueBool(),
		PushEnabled:      plan.PushEnabled.ValueBool(),
		TlsOnly:          plan.TlsOnly.ValueBool(),
		ExposeTimeserial: plan.ExposeTimeserial.ValueBool(),
	}

	if plan.BatchingEnabled.ValueBool() {
		namespaceValues.BatchingEnabled = true
		namespaceValues.BatchingInterval = control.Interval(int(plan.BatchingInterval.ValueInt64()))
	}

	if plan.ConflationEnabled.ValueBool() {
		namespaceValues.ConflationEnabled = true
		namespaceValues.ConflationInterval = control.Interval(int(plan.ConflationInterval.ValueInt64()))
		namespaceValues.ConflationKey = plan.ConflationKey.ValueString()
	}

	// Creates a new Ably namespace by invoking the CreateNamespace function from the Client Library
	ablyNamespace, err := r.p.client.CreateNamespace(plan.AppID.ValueString(), &namespaceValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	respApps := AblyNamespace{
		AppID:             types.StringValue(plan.AppID.ValueString()),
		ID:                types.StringValue(ablyNamespace.ID),
		Authenticated:     types.BoolValue(ablyNamespace.Authenticated),
		Persisted:         types.BoolValue(ablyNamespace.Persisted),
		PersistLast:       types.BoolValue(ablyNamespace.PersistLast),
		PushEnabled:       types.BoolValue(ablyNamespace.PushEnabled),
		TlsOnly:           types.BoolValue(ablyNamespace.TlsOnly),
		ExposeTimeserial:  types.BoolValue(namespaceValues.ExposeTimeserial),
		BatchingEnabled:   types.BoolValue(ablyNamespace.BatchingEnabled),
		ConflationEnabled: types.BoolValue(ablyNamespace.ConflationEnabled),
	}

	if ablyNamespace.BatchingEnabled {
		respApps.BatchingInterval = ptrValueInt(ablyNamespace.BatchingInterval)
	}

	if ablyNamespace.ConflationEnabled {
		respApps.ConflationInterval = ptrValueInt(ablyNamespace.ConflationInterval)
		respApps.ConflationKey = types.StringValue(ablyNamespace.ConflationKey)
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, respApps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r ResourceNamespace) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_namespace"
}

// Read reads the resource.
func (r ResourceNamespace) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyNamespace
	found := false
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID and namespace ID value for the resource
	appID := state.AppID.ValueString()
	namespaceID := state.ID.ValueString()

	// Fetches all Ably Namespaces in the app. The function invokes the Client Library Namespaces() method.
	// NOTE: Control API & Client Lib do not currently support fetching single namespace given namespace id
	namespaces, err := r.p.client.Namespaces(appID)
	if err != nil {
		if is404(err) {
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
		if v.ID == namespaceID {
			respNamespaces := AblyNamespace{
				AppID:             types.StringValue(appID),
				ID:                types.StringValue(namespaceID),
				Authenticated:     types.BoolValue(v.Authenticated),
				Persisted:         types.BoolValue(v.Persisted),
				PersistLast:       types.BoolValue(v.PersistLast),
				PushEnabled:       types.BoolValue(v.PushEnabled),
				TlsOnly:           types.BoolValue(v.TlsOnly),
				ExposeTimeserial:  types.BoolValue(v.ExposeTimeserial),
				BatchingEnabled:   types.BoolValue(v.BatchingEnabled),
				ConflationEnabled: types.BoolValue(v.ConflationEnabled),
			}

			if v.BatchingEnabled {
				respNamespaces.BatchingInterval = ptrValueInt(v.BatchingInterval)
			}

			if v.ConflationEnabled {
				respNamespaces.ConflationInterval = ptrValueInt(v.ConflationInterval)
				respNamespaces.ConflationKey = types.StringValue(v.ConflationKey)
			}

			// Sets state to namespace values.
			diags = resp.State.Set(ctx, &respNamespaces)
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

// Update updates an existing resource.
func (r ResourceNamespace) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan AblyNamespace
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the app ID and ID
	appID := plan.AppID.ValueString()
	namespaceID := plan.ID.ValueString()

	// Instantiates struct of type control.Namespace and sets values to output of plan
	namespaceValues := control.Namespace{
		ID:               namespaceID,
		Authenticated:    plan.Authenticated.ValueBool(),
		Persisted:        plan.Persisted.ValueBool(),
		PersistLast:      plan.PersistLast.ValueBool(),
		PushEnabled:      plan.PushEnabled.ValueBool(),
		TlsOnly:          plan.TlsOnly.ValueBool(),
		ExposeTimeserial: plan.ExposeTimeserial.ValueBool(),
	}

	if plan.BatchingEnabled.ValueBool() {
		namespaceValues.BatchingEnabled = true
		namespaceValues.BatchingInterval = control.Interval(int(plan.BatchingInterval.ValueInt64()))
	}

	if plan.ConflationEnabled.ValueBool() {
		namespaceValues.ConflationEnabled = true
		namespaceValues.ConflationInterval = control.Interval(int(plan.ConflationInterval.ValueInt64()))
		namespaceValues.ConflationKey = plan.ConflationKey.ValueString()
	}

	// Updates an Ably Namespace. The function invokes the Client Library UpdateNamespace method.
	ablyNamespace, err := r.p.client.UpdateNamespace(appID, &namespaceValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	respNamespaces := AblyNamespace{
		AppID:             types.StringValue(appID),
		ID:                types.StringValue(ablyNamespace.ID),
		Authenticated:     types.BoolValue(ablyNamespace.Authenticated),
		Persisted:         types.BoolValue(ablyNamespace.Persisted),
		PersistLast:       types.BoolValue(ablyNamespace.PersistLast),
		PushEnabled:       types.BoolValue(ablyNamespace.PushEnabled),
		TlsOnly:           types.BoolValue(ablyNamespace.TlsOnly),
		ExposeTimeserial:  types.BoolValue(ablyNamespace.ExposeTimeserial),
		BatchingEnabled:   types.BoolValue(ablyNamespace.BatchingEnabled),
		ConflationEnabled: types.BoolValue(ablyNamespace.ConflationEnabled),
	}

	if ablyNamespace.BatchingEnabled {
		respNamespaces.BatchingInterval = ptrValueInt(ablyNamespace.BatchingInterval)
	}

	if ablyNamespace.ConflationEnabled {
		respNamespaces.ConflationInterval = ptrValueInt(ablyNamespace.ConflationInterval)
		respNamespaces.ConflationKey = types.StringValue(ablyNamespace.ConflationKey)
	}

	// Sets state to new namespace.
	diags = resp.State.Set(ctx, respNamespaces)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource.
func (r ResourceNamespace) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state AblyNamespace
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	appID := state.AppID.ValueString()
	namespaceID := state.ID.ValueString()

	err := r.p.client.DeleteNamespace(appID, namespaceID)
	if err != nil {
		if is404(err) {
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

// ImportState handles the import state functionality.
func (r ResourceNamespace) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ImportResource(ctx, req, resp, "app_id", "id")
}

// Safely return an interval as null if it's not set
func ptrValueInt(in *int) types.Int64 {
	res := types.Int64Null()
	if in != nil {
		res = types.Int64Value(int64(*in))
	}
	return res
}
