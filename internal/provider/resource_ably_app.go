package ably_control

import (
	"context"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	tfsdk_resource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type resourceAppType struct{}

// Get App Resource schema
func (r resourceAppType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The application ID.",
			},
			"account_id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The ID of your Ably account.",
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
					DefaultAttribute(types.String{Value: "enabled"}),
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
				Description: "The Apple Push Notification service sandbox endpoint.",
				PlanModifiers: []tfsdk.AttributePlanModifier{
					DefaultAttribute(types.Bool{Value: false}),
				},
			},
		},
		MarkdownDescription: "The `ably_app` resource allows you to create and manage Ably Apps.",
	}, nil
}

// New resource instance
func (r resourceAppType) NewResource(_ context.Context, p tfsdk_provider.Provider) (tfsdk_resource.Resource, diag.Diagnostics) {
	return resourceApp{
		p: *(p.(*provider)),
	}, nil
}

type resourceApp struct {
	p provider
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
	app_values := ably_control_go.App{
		ID:                     plan.ID.Value,
		AccountID:              plan.AccountID.Value,
		Name:                   plan.Name.Value,
		Status:                 plan.Status.Value,
		TLSOnly:                plan.TLSOnly.Value,
		FcmKey:                 plan.FcmKey.Value,
		ApnsCertificate:        plan.ApnsCertificate.Value,
		ApnsPrivateKey:         plan.ApnsPrivateKey.Value,
		ApnsUseSandboxEndpoint: plan.ApnsUseSandboxEndpoint.Value,
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
		AccountID:              types.String{Value: ably_app.AccountID},
		ID:                     types.String{Value: ably_app.ID},
		Name:                   types.String{Value: ably_app.Name},
		Status:                 types.String{Value: ably_app.Status},
		TLSOnly:                types.Bool{Value: ably_app.TLSOnly},
		FcmKey:                 plan.FcmKey,
		ApnsCertificate:        plan.ApnsCertificate,
		ApnsPrivateKey:         plan.ApnsPrivateKey,
		ApnsUseSandboxEndpoint: types.Bool{Value: ably_app.ApnsUseSandboxEndpoint},
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
func (r resourceApp) Read(ctx context.Context, req tfsdk_resource.ReadRequest, resp *tfsdk_resource.ReadResponse) {
	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyApp
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID value for the resource
	app_id := state.ID.Value

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
				AccountID:              types.String{Value: v.AccountID},
				ID:                     types.String{Value: v.ID},
				Name:                   types.String{Value: v.Name},
				Status:                 types.String{Value: v.Status},
				TLSOnly:                types.Bool{Value: v.TLSOnly},
				FcmKey:                 state.FcmKey,
				ApnsCertificate:        state.ApnsCertificate,
				ApnsPrivateKey:         state.ApnsPrivateKey,
				ApnsUseSandboxEndpoint: types.Bool{Value: v.ApnsUseSandboxEndpoint},
			}
			emptyStringToNull(&resp_apps.FcmKey)
			emptyStringToNull(&resp_apps.ApnsCertificate)
			emptyStringToNull(&resp_apps.ApnsPrivateKey)

			// Sets state to app values.
			diags = resp.State.Set(ctx, &resp_apps)

			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
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

	// Get current state
	var state AblyApp
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the app ID
	app_id := state.ID.Value

	// Instantiates struct of type ably_control_go.App and sets values to output of plan
	app_values := ably_control_go.App{
		ID:                     plan.ID.Value,
		AccountID:              plan.AccountID.Value,
		Name:                   plan.Name.Value,
		Status:                 plan.Status.Value,
		TLSOnly:                plan.TLSOnly.Value,
		FcmKey:                 plan.FcmKey.Value,
		ApnsCertificate:        plan.ApnsCertificate.Value,
		ApnsPrivateKey:         plan.ApnsPrivateKey.Value,
		ApnsUseSandboxEndpoint: plan.ApnsUseSandboxEndpoint.Value,
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
		ID:                     types.String{Value: ably_app.ID},
		AccountID:              types.String{Value: ably_app.AccountID},
		Name:                   types.String{Value: ably_app.Name},
		Status:                 types.String{Value: ably_app.Status},
		TLSOnly:                types.Bool{Value: ably_app.TLSOnly},
		FcmKey:                 plan.FcmKey,
		ApnsCertificate:        plan.ApnsCertificate,
		ApnsPrivateKey:         plan.ApnsPrivateKey,
		ApnsUseSandboxEndpoint: types.Bool{Value: ably_app.ApnsUseSandboxEndpoint},
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
func (r resourceApp) Delete(ctx context.Context, req tfsdk_resource.DeleteRequest, resp *tfsdk_resource.DeleteResponse) {
	// Get current state
	var state AblyApp
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	app_id := state.ID.Value

	err := r.p.client.DeleteApp(app_id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Resource",
			"Could not delete resource, unexpected error: "+err.Error(),
		)
		return
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
