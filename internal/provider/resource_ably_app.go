package ably_control

import (
	"context"

	ably_control_go "github.com/ably/ably-control-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
			"status": {
				Type:     types.StringType,
				Optional: true,
				// TODO: Update this after Control API bug has been fixed.
				Description: "The application status. Disabled applications will not accept new connections and will return an error to all clients. When creating a new application, ensure that its status is set to enabled.",
			},
			"tls_only": {
				Type:        types.BoolType,
				Optional:    true,
				Description: "Enforce TLS for all connections. This setting overrides any channel setting.",
			},
		},
	}, nil
}

// New resource instance
func (r resourceAppType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceApp{
		p: *(p.(*provider)),
	}, nil
}

type resourceApp struct {
	p provider
}

// Create a new resource
func (r resourceApp) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
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
		ID:        plan.ID.Value,
		AccountID: plan.AccountID.Value,
		Name:      plan.Name.Value,
		Status:    plan.Status.Value,
		TLSOnly:   plan.TLSOnly.Value,
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
		ID:        types.String{Value: ably_app.ID},
		AccountID: types.String{Value: ably_app.AccountID},
		Name:      types.String{Value: ably_app.Name},
		Status:    types.String{Value: ably_app.Status},
		TLSOnly:   types.Bool{Value: ably_app.TLSOnly},
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, resp_apps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read resource
func (r resourceApp) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
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
	apps, _ := r.p.client.Apps()

	// Loops through apps and if account id matches, sets state.
	for _, v := range apps {
		if v.ID == app_id {
			resp_apps := AblyApp{
				ID:        types.String{Value: v.ID},
				AccountID: types.String{Value: v.AccountID},
				Name:      types.String{Value: v.Name},
				Status:    types.String{Value: v.Status},
				TLSOnly:   types.Bool{Value: v.TLSOnly},
			}
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
func (r resourceApp) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
}

// Delete resource
func (r resourceApp) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
}
