// Package provider implements the Ably provider for Terraform
package provider

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

// Schema defines the schema for the resource.
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
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					DefaultBoolAttribute(types.BoolValue(false)),
				},
			},
			"fcm_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The Firebase Cloud Messaging key.",
			},
			"fcm_service_account": schema.StringAttribute{
				Optional:    true,
				Description: "Used to specify the Firebase Cloud Messaging(FCM) service account credentials used for authentication and enabling communication with FCM to send push notifications to devices.",
				Sensitive:   true,
			},
			"fcm_project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The unique identifier for the Firebase Cloud Messaging(FCM) project. This ID is used to specify the Firebase project when configuring FCM or other Firebase services.",
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

// Create creates a new resource.
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
	appValues := control.NewApp{
		ID:                     plan.ID.ValueString(),
		Name:                   plan.Name.ValueString(),
		Status:                 plan.Status.ValueString(),
		TLSOnly:                plan.TLSOnly.ValueBool(),
		FcmKey:                 plan.FcmKey.ValueString(),
		FcmServiceAccount:      plan.FcmServiceAccount.ValueString(),
		FcmProjectId:           plan.FcmProjectId.ValueString(),
		ApnsCertificate:        plan.ApnsCertificate.ValueString(),
		ApnsPrivateKey:         plan.ApnsPrivateKey.ValueString(),
		ApnsUseSandboxEndpoint: plan.ApnsUseSandboxEndpoint.ValueBool(),
	}

	// Creates a new Ably App by invoking the CreateApp function from the Client Library
	var ablyApp control.App
	err := retryWithBackoff(ctx, "CreateApp", func() error {
		var createErr error
		ablyApp, createErr = r.p.client.CreateApp(&appValues)
		return createErr
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resource",
			"Could not create resource, unexpected error: "+err.Error(),
		)
		return
	}

	// Maps response body to resource schema attributes.
	respApps := AblyApp{
		AccountID:              types.StringValue(ablyApp.AccountID),
		ID:                     types.StringValue(ablyApp.ID),
		Name:                   types.StringValue(ablyApp.Name),
		Status:                 types.StringValue(ablyApp.Status),
		TLSOnly:                types.BoolValue(ablyApp.TLSOnly),
		FcmKey:                 plan.FcmKey,
		FcmServiceAccount:      plan.FcmServiceAccount,
		FcmProjectId:           plan.FcmProjectId,
		ApnsCertificate:        plan.ApnsCertificate,
		ApnsPrivateKey:         plan.ApnsPrivateKey,
		ApnsUseSandboxEndpoint: types.BoolValue(ablyApp.ApnsUseSandboxEndpoint),
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

// Read reads the resource.
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
	appID := state.ID.ValueString()

	// Fetches all Ably Apps in the account. The function invokes the Client Library Apps() method.
	// NOTE: Control API & Client Lib do not currently support fetching single app given app id
	var apps []control.App
	err := retryWithBackoff(ctx, "Apps", func() error {
		var readErr error
		apps, readErr = r.p.client.Apps()
		return readErr
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Resource",
			"Could not read resource, unexpected error: "+err.Error(),
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
				FcmServiceAccount:      state.FcmServiceAccount,
				FcmProjectId:           state.FcmProjectId,
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

// Update updates an existing resource.
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
	appID := plan.ID.ValueString()
	if plan.ID.IsUnknown() {
		appID = state.ID.ValueString()
	}

	// Instantiates struct of type control.App and sets values to output of plan
	appValues := control.NewApp{
		Name:                   plan.Name.ValueString(),
		Status:                 plan.Status.ValueString(),
		TLSOnly:                plan.TLSOnly.ValueBool(),
		FcmKey:                 plan.FcmKey.ValueString(),
		FcmServiceAccount:      plan.FcmServiceAccount.ValueString(),
		FcmProjectId:           plan.FcmProjectId.ValueString(),
		ApnsCertificate:        plan.ApnsCertificate.ValueString(),
		ApnsPrivateKey:         plan.ApnsPrivateKey.ValueString(),
		ApnsUseSandboxEndpoint: plan.ApnsUseSandboxEndpoint.ValueBool(),
	}

	// Updates an Ably App. The function invokes the Client Library UpdateApp method.
	var ablyApp control.App
	err := retryWithBackoff(ctx, "UpdateApp", func() error {
		var updateErr error
		ablyApp, updateErr = r.p.client.UpdateApp(appID, &appValues)
		return updateErr
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Resource",
			"Could not update resource, unexpected error: "+err.Error(),
		)
		return
	}

	respApps := AblyApp{
		ID:                     types.StringValue(ablyApp.ID),
		AccountID:              types.StringValue(ablyApp.AccountID),
		Name:                   types.StringValue(ablyApp.Name),
		Status:                 types.StringValue(ablyApp.Status),
		TLSOnly:                types.BoolValue(ablyApp.TLSOnly),
		FcmKey:                 plan.FcmKey,
		FcmServiceAccount:      plan.FcmServiceAccount,
		FcmProjectId:           plan.FcmProjectId,
		ApnsCertificate:        plan.ApnsCertificate,
		ApnsPrivateKey:         plan.ApnsPrivateKey,
		ApnsUseSandboxEndpoint: types.BoolValue(ablyApp.ApnsUseSandboxEndpoint),
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

// Delete deletes the resource.
func (r ResourceApp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state AblyApp
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	appID := state.ID.ValueString()

	err := retryWithBackoff(ctx, "DeleteApp", func() error {
		return r.p.client.DeleteApp(appID)
	})
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
func (r ResourceApp) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Save the import identifier in the id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
