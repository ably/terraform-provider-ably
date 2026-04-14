// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
			"mutable_messages": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enables message editing and deletion on the namespace. When enabled, messages published to channels matching this namespace can be modified or deleted.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"populate_channel_registry": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When enabled, channels matching this namespace will appear in the channel registry, allowing channel enumeration.",
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
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
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
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
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

// buildNamespaceState reconciles plan/state input with an API response.
func buildNamespaceState(rc *reconciler, input AblyNamespace, api control.NamespaceResponse) AblyNamespace {
	ns := AblyNamespace{
		AppID:                   rc.str("app_id", input.AppID, types.StringValue(api.AppID), false),
		ID:                      rc.str("id", input.ID, types.StringValue(api.ID), false),
		Authenticated:           rc.boolean("authenticated", input.Authenticated, types.BoolValue(api.Authenticated), true),
		Persisted:               rc.boolean("persisted", input.Persisted, types.BoolValue(api.Persisted), true),
		PersistLast:             rc.boolean("persist_last", input.PersistLast, types.BoolValue(api.PersistLast), true),
		PushEnabled:             rc.boolean("push_enabled", input.PushEnabled, types.BoolValue(api.PushEnabled), true),
		TlsOnly:                 rc.boolean("tls_only", input.TlsOnly, types.BoolValue(api.TLSOnly), true),
		ExposeTimeserial:        rc.boolean("expose_timeserial", input.ExposeTimeserial, types.BoolValue(api.ExposeTimeserial), true),
		MutableMessages:         rc.boolean("mutable_messages", input.MutableMessages, types.BoolValue(api.MutableMessages), true),
		PopulateChannelRegistry: rc.boolean("populate_channel_registry", input.PopulateChannelRegistry, types.BoolValue(api.PopulateChannelRegistry), true),
		BatchingEnabled:         rc.boolean("batching_enabled", input.BatchingEnabled, optBoolValue(api.BatchingEnabled), true),
		ConflationEnabled:       rc.boolean("conflation_enabled", input.ConflationEnabled, optBoolValue(api.ConflationEnabled), true),
	}

	if api.BatchingEnabled != nil && *api.BatchingEnabled {
		ns.BatchingInterval = rc.int64val("batching_interval", input.BatchingInterval, optIntValue(api.BatchingInterval), true)
	}

	if api.ConflationEnabled != nil && *api.ConflationEnabled {
		ns.ConflationInterval = rc.int64val("conflation_interval", input.ConflationInterval, optIntValue(api.ConflationInterval), true)
		ns.ConflationKey = rc.str("conflation_key", input.ConflationKey, optStringValue(api.ConflationKey), true)
	}

	return ns
}

// Create creates a new resource.
func (r ResourceNamespace) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
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
	namespaceValues := control.NamespacePost{
		ID:                      plan.ID.ValueString(),
		Authenticated:           plan.Authenticated.ValueBool(),
		Persisted:               plan.Persisted.ValueBool(),
		PersistLast:             plan.PersistLast.ValueBool(),
		PushEnabled:             plan.PushEnabled.ValueBool(),
		TLSOnly:                 plan.TlsOnly.ValueBool(),
		ExposeTimeserial:        plan.ExposeTimeserial.ValueBool(),
		MutableMessages:         plan.MutableMessages.ValueBool(),
		PopulateChannelRegistry: plan.PopulateChannelRegistry.ValueBool(),
	}

	if plan.BatchingEnabled.ValueBool() {
		namespaceValues.BatchingEnabled = ptr(true)
		if !plan.BatchingInterval.IsNull() && !plan.BatchingInterval.IsUnknown() {
			namespaceValues.BatchingInterval = ptr(int(plan.BatchingInterval.ValueInt64()))
		}
	}

	if plan.ConflationEnabled.ValueBool() {
		namespaceValues.ConflationEnabled = ptr(true)
		if !plan.ConflationInterval.IsNull() && !plan.ConflationInterval.IsUnknown() {
			namespaceValues.ConflationInterval = ptr(int(plan.ConflationInterval.ValueInt64()))
		}

		if !plan.ConflationKey.IsNull() {
			namespaceValues.ConflationKey = ptr(plan.ConflationKey.ValueString())
		}
	}

	// Creates a new Ably namespace by invoking the CreateNamespace function from the Client Library
	ablyNamespace, err := r.p.client.CreateNamespace(ctx, plan.AppID.ValueString(), namespaceValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ably_namespace",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	rc := newReconciler(&resp.Diagnostics)
	respNs := buildNamespaceState(rc, plan, ablyNamespace)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sets state for the new Ably namespace.
	diags = resp.State.Set(ctx, respNs)
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
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

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

	// Fetches all Ably Namespaces in the app. The function invokes the Client Library ListNamespaces() method.
	// NOTE: Control API & Client Lib do not currently support fetching single namespace given namespace id
	namespaces, err := r.p.client.ListNamespaces(ctx, appID)
	if err != nil {
		if is404(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading ably_namespace",
			"Could not read resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Loops through namespaces and if id matches, sets state.
	for _, v := range namespaces {
		if v.ID == namespaceID {
			rc := newReconciler(&resp.Diagnostics).forRead()
			respNs := buildNamespaceState(rc, state, v)
			if resp.Diagnostics.HasError() {
				return
			}

			// Sets state to namespace values.
			diags = resp.State.Set(ctx, &respNs)
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
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

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

	// Instantiates struct of type control.NamespacePatch and sets values to output of plan
	namespaceValues := control.NamespacePatch{
		Authenticated:           ptr(plan.Authenticated.ValueBool()),
		Persisted:               ptr(plan.Persisted.ValueBool()),
		PersistLast:             ptr(plan.PersistLast.ValueBool()),
		PushEnabled:             ptr(plan.PushEnabled.ValueBool()),
		TLSOnly:                 ptr(plan.TlsOnly.ValueBool()),
		ExposeTimeserial:        ptr(plan.ExposeTimeserial.ValueBool()),
		MutableMessages:         ptr(plan.MutableMessages.ValueBool()),
		PopulateChannelRegistry: ptr(plan.PopulateChannelRegistry.ValueBool()),
	}

	batchEnabled := plan.BatchingEnabled.ValueBool()
	namespaceValues.BatchingEnabled = &batchEnabled
	if batchEnabled {
		if !plan.BatchingInterval.IsNull() && !plan.BatchingInterval.IsUnknown() {
			namespaceValues.BatchingInterval = ptr(int(plan.BatchingInterval.ValueInt64()))
		}
	}

	conflationEnabled := plan.ConflationEnabled.ValueBool()
	namespaceValues.ConflationEnabled = &conflationEnabled
	if conflationEnabled {
		if !plan.ConflationInterval.IsNull() && !plan.ConflationInterval.IsUnknown() {
			namespaceValues.ConflationInterval = ptr(int(plan.ConflationInterval.ValueInt64()))
		}

		if !plan.ConflationKey.IsNull() {
			namespaceValues.ConflationKey = ptr(plan.ConflationKey.ValueString())
		}
	}

	// Updates an Ably Namespace. The function invokes the Client Library UpdateNamespace method.
	ablyNamespace, err := r.p.client.UpdateNamespace(ctx, appID, namespaceID, namespaceValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ably_namespace",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	rc := newReconciler(&resp.Diagnostics)
	respNs := buildNamespaceState(rc, plan, ablyNamespace)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sets state to new namespace.
	diags = resp.State.Set(ctx, respNs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource.
func (r ResourceNamespace) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

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

	err := r.p.client.DeleteNamespace(ctx, appID, namespaceID)
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				"Resource does not exist",
				"Resource does not exist, it may have already been deleted: "+err.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error deleting ably_namespace",
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
