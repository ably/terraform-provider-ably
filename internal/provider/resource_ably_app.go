package provider

import (
	"context"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceApp struct {
	p *provider
}

// Get App Resource schema
func (r resourceApp) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The application ID.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.UseStateForUnknown(),
				},
			},
			"account_id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The ID of your Ably account.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk_resource.UseStateForUnknown(),
				},
			},
			"name": {
				Type:        types.StringType,
				Required:    true,
				Description: "The application name.",
			},
			// TODO: Update this after Control API bug has been fixed.
			"status": {
				Type:        types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "The application status. Disabled applications will not accept new connections and will return an error to all clients. When creating a new application, ensure that its status is set to enabled.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.StringValue("enabled")),
				},
			},
			"tls_only": {
				Type:        types.BoolType,
				Optional:    true,
				Description: "Enforce TLS for all connections. This setting overrides any channel setting.",
			},
			"fcm_key": {
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
				Description: "The Firebase Cloud Messaging key.",
			},
			"apns_certificate": {
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
				Description: "The Apple Push Notification service certificate.",
			},
			"apns_private_key": {
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
				Description: "The Apple Push Notification service private key.",
			},
			"apns_use_sandbox_endpoint": {
				Type:        types.BoolType,
				Optional:    true,
				Computed:    true,
				Description: "Use the Apple Push Notification service sandbox endpoint.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.BoolValue(false)),
				},
			},
		},
		MarkdownDescription: "The `ably_app` resource allows you to create and manage Ably Apps " +
			"and configure Ably Push notifications. Read more about Ably Push Notifications in Ably documentation: https://ably.com/docs/general/push",
	}, nil
}

func (r resourceApp) Metadata(ctx context.Context, req tfsdk_resource.MetadataRequest, resp *tfsdk_resource.MetadataResponse) {
	resp.TypeName = "ably_app"
}

// Create a new resource
func (r resourceApp) Create(ctx context.Context, req tfsdk_resource.CreateRequest, resp *tfsdk_resource.CreateResponse) {
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
	appValues := ably_control_go.NewApp{
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
	app, err := r.p.client.CreateApp(&appValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	respApps := AblyApp{
		AccountID:              types.StringValue(app.AccountID),
		ID:                     types.StringValue(app.ID),
		Name:                   types.StringValue(app.Name),
		Status:                 types.StringValue(app.Status),
		TLSOnly:                types.BoolValue(app.TLSOnly),
		FcmKey:                 plan.FcmKey,
		ApnsCertificate:        plan.ApnsCertificate,
		ApnsPrivateKey:         plan.ApnsPrivateKey,
		ApnsUseSandboxEndpoint: types.BoolValue(app.ApnsUseSandboxEndpoint),
	}
	emptyStringToNull(&respApps.FcmKey)
	emptyStringToNull(&respApps.ApnsCertificate)
	emptyStringToNull(&respApps.ApnsPrivateKey)

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, respApps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource
func (r resourceApp) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyApp
	found := false
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID value for the resource
	appID := state.ID.ValueString()

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
		if v.ID == appID {
			respApps := AblyApp{
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
			emptyStringToNull(&respApps.FcmKey)
			emptyStringToNull(&respApps.ApnsCertificate)
			emptyStringToNull(&respApps.ApnsPrivateKey)
			found = true

			// Sets state to app values.
			diags = resp.State.Set(ctx, &respApps)

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
func (r resourceApp) Update(ctx context.Context, req tfsdk_resource.UpdateRequest, resp *tfsdk_resource.UpdateResponse) {
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
	appID := plan.ID.ValueString()
	if plan.ID.IsUnknown() {
		appID = state.ID.ValueString()
	}

	// Instantiates struct of type ably_control_go.App and sets values to output of plan
	appValues := ably_control_go.NewApp{
		Name:                   plan.Name.ValueString(),
		Status:                 plan.Status.ValueString(),
		TLSOnly:                plan.TLSOnly.ValueBool(),
		FcmKey:                 plan.FcmKey.ValueString(),
		ApnsCertificate:        plan.ApnsCertificate.ValueString(),
		ApnsPrivateKey:         plan.ApnsPrivateKey.ValueString(),
		ApnsUseSandboxEndpoint: plan.ApnsUseSandboxEndpoint.ValueBool(),
	}

	// Updates an Ably App. The function invokes the Client Library UpdateApp method.
	app, err := r.p.client.UpdateApp(appID, &appValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	respApps := AblyApp{
		ID:                     types.StringValue(app.ID),
		AccountID:              types.StringValue(app.AccountID),
		Name:                   types.StringValue(app.Name),
		Status:                 types.StringValue(app.Status),
		TLSOnly:                types.BoolValue(app.TLSOnly),
		FcmKey:                 plan.FcmKey,
		ApnsCertificate:        plan.ApnsCertificate,
		ApnsPrivateKey:         plan.ApnsPrivateKey,
		ApnsUseSandboxEndpoint: types.BoolValue(app.ApnsUseSandboxEndpoint),
	}
	emptyStringToNull(&respApps.FcmKey)
	emptyStringToNull(&respApps.ApnsCertificate)
	emptyStringToNull(&respApps.ApnsPrivateKey)

	// Sets state to new app.
	diags = resp.State.Set(ctx, respApps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete resource
func (r resourceApp) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	// Get current state
	var state AblyApp
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	appID := state.ID.ValueString()

	err := r.p.client.DeleteApp(appID)
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

// Import resource
func (r resourceApp) ImportState(ctx context.Context, req tfsdk_resource.ImportStateRequest, resp *tfsdk_resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	// Recent PR in TF Plugin Framework for paths but Hashicorp examples not updated - https://github.com/hashicorp/terraform-plugin-framework/pull/390
	tfsdk_resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
