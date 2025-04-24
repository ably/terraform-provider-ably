package ably_control

import (
	"context"

	control "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ResourceApp{}
var _ resource.ResourceWithImportState = &ResourceApp{}

type ResourceApp struct {
	p *AblyProvider
}

// Get App Resource schema
func (r ResourceApp) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The application ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of your Ably account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The application name.",
			},
			// TODO: Update this after Control API bug has been fixed.
			"status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The application status. Disabled applications will not accept new connections and will return an error to all clients. When creating a new application, ensure that its status is set to enabled.",
				PlanModifiers: []planmodifier.String{
					DefaultStringAttribute(types.StringValue("enabled")),
				},
			},
			"tls_only": schema.BoolAttribute{
				Optional:    true,
				Description: "Enforce TLS for all connections. This setting overrides any channel setting.",
			},
			"fcm_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Firebase Cloud Messaging key.",
			},
			"apns_certificate": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Apple Push Notification service certificate.",
			},
			"apns_private_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Apple Push Notification service private key.",
			},
			"apns_use_sandbox_endpoint": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Use the Apple Push Notification service sandbox endpoint.",
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
		},
		MarkdownDescription: "The `ably_app` resource allows you to create and manage Ably Apps " +
			"and configure Ably Push notifications. Read more about Ably Push Notifications in Ably documentation: https://ably.com/docs/general/push",
	}
}

func (r ResourceApp) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_app"
}

// Create a new resource
func (r ResourceApp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Checks whether the provider and API Client are configured. If they are not, the provider responds with an error.
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply",
		)
		return
	}

	// Gets plan values
	var plan AblyApp
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generates an API request body from the plan values
	app_values := control.NewApp{
		ID:                     plan.ID.ValueString(),
		Name:                   plan.Name.ValueString(),
		Status:                 plan.Status.ValueString(),
		TLSOnly:                plan.TLSOnly.ValueBool(),
		FcmKey:                 plan.FcmKey.ValueString(),
		ApnsCertificate:        plan.ApnsCertificate.ValueString(),
		ApnsPrivateKey:         plan.ApnsPrivateKey.ValueString(),
		ApnsUseSandboxEndpoint: plan.ApnsUseSandboxEndpoint.ValueBool(),
	}

	// Creates a new Ably App by invoking the CreateApp function from the Client Library
	ably_app, err := r.p.client.CreateApp(&app_values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	resp_apps := AblyApp{
		AccountID:              types.StringValue(ably_app.AccountID),
		ID:                     types.StringValue(ably_app.ID),
		Name:                   types.StringValue(ably_app.Name),
		Status:                 types.StringValue(ably_app.Status),
		TLSOnly:                types.BoolValue(ably_app.TLSOnly),
		FcmKey:                 plan.FcmKey,
		ApnsCertificate:        plan.ApnsCertificate,
		ApnsPrivateKey:         plan.ApnsPrivateKey,
		ApnsUseSandboxEndpoint: types.BoolValue(ably_app.ApnsUseSandboxEndpoint),
	}
	emptyStringToNull(&resp_apps.FcmKey)
	emptyStringToNull(&resp_apps.ApnsCertificate)
	emptyStringToNull(&resp_apps.ApnsPrivateKey)

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, resp_apps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource
func (r ResourceApp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyApp
	found := false
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID value for the resource
	app_id := state.ID.ValueString()

	// Fetches all Ably Apps in the account. The function invokes the Client Library Apps() method.
	// NOTE: Control API & Client Lib do not currently support fetching single app given app id
	apps, err := r.p.client.Apps()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Loops through apps and if account id matches, sets state.
	for _, v := range apps {
		if v.ID == app_id {
			resp_apps := AblyApp{
				AccountID:              types.StringValue(v.AccountID),
				ID:                     types.StringValue(v.ID),
				Name:                   types.StringValue(v.Name),
				Status:                 types.StringValue(v.Status),
				TLSOnly:                types.BoolValue(v.TLSOnly),
				FcmKey:                 state.FcmKey,
				ApnsCertificate:        state.ApnsCertificate,
				ApnsPrivateKey:         state.ApnsPrivateKey,
				ApnsUseSandboxEndpoint: types.BoolValue(v.ApnsUseSandboxEndpoint),
			}
			emptyStringToNull(&resp_apps.FcmKey)
			emptyStringToNull(&resp_apps.ApnsCertificate)
			emptyStringToNull(&resp_apps.ApnsPrivateKey)
			found = true

			// Sets state to app values.
			diags = resp.State.Set(ctx, &resp_apps)

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
func (r ResourceApp) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan AblyApp
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AblyApp
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the app ID
	app_id := plan.ID.ValueString()
	if plan.ID.IsUnknown() {
		app_id = state.ID.ValueString()
	}

	// Instantiates struct of type control.App and sets values to output of plan
	app_values := control.NewApp{
		Name:                   plan.Name.ValueString(),
		Status:                 plan.Status.ValueString(),
		TLSOnly:                plan.TLSOnly.ValueBool(),
		FcmKey:                 plan.FcmKey.ValueString(),
		ApnsCertificate:        plan.ApnsCertificate.ValueString(),
		ApnsPrivateKey:         plan.ApnsPrivateKey.ValueString(),
		ApnsUseSandboxEndpoint: plan.ApnsUseSandboxEndpoint.ValueBool(),
	}

	// Updates an Ably App. The function invokes the Client Library UpdateApp method.
	ably_app, err := r.p.client.UpdateApp(app_id, &app_values)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	resp_apps := AblyApp{
		ID:                     types.StringValue(ably_app.ID),
		AccountID:              types.StringValue(ably_app.AccountID),
		Name:                   types.StringValue(ably_app.Name),
		Status:                 types.StringValue(ably_app.Status),
		TLSOnly:                types.BoolValue(ably_app.TLSOnly),
		FcmKey:                 plan.FcmKey,
		ApnsCertificate:        plan.ApnsCertificate,
		ApnsPrivateKey:         plan.ApnsPrivateKey,
		ApnsUseSandboxEndpoint: types.BoolValue(ably_app.ApnsUseSandboxEndpoint),
	}
	emptyStringToNull(&resp_apps.FcmKey)
	emptyStringToNull(&resp_apps.ApnsCertificate)
	emptyStringToNull(&resp_apps.ApnsPrivateKey)

	// Sets state to new app.
	diags = resp.State.Set(ctx, resp_apps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r ResourceApp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state AblyApp
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	app_id := state.ID.ValueString()

	err := r.p.client.DeleteApp(app_id)
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
func (r ResourceApp) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
