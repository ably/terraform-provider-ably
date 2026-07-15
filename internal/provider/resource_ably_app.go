// Package provider implements the Ably provider for Terraform
package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/ably/terraform-provider-ably/control"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ResourceApp{}
var _ resource.ResourceWithImportState = &ResourceApp{}

type ResourceApp struct {
	p *AblyProvider
}

// AblyAppState extends the base AblyApp model with computed timestamp fields.
type AblyAppState struct {
	AccountID                   types.String `tfsdk:"account_id"`
	ID                          types.String `tfsdk:"id"`
	Name                        types.String `tfsdk:"name"`
	Status                      types.String `tfsdk:"status"`
	TLSOnly                     types.Bool   `tfsdk:"tls_only"`
	FcmKey                      types.String `tfsdk:"fcm_key"`
	FcmServiceAccount           types.String `tfsdk:"fcm_service_account"`
	FcmProjectId                types.String `tfsdk:"fcm_project_id"`
	FcmServiceAccountConfigured types.Bool   `tfsdk:"fcm_service_account_configured"`
	ApnsCertificate             types.String `tfsdk:"apns_certificate"`
	ApnsPrivateKey              types.String `tfsdk:"apns_private_key"`
	ApnsUseSandboxEndpoint      types.Bool   `tfsdk:"apns_use_sandbox_endpoint"`
	ApnsAuthType                types.String `tfsdk:"apns_auth_type"`
	ApnsSigningKey              types.String `tfsdk:"apns_signing_key"`
	ApnsSigningKeyId            types.String `tfsdk:"apns_signing_key_id"`
	ApnsIssuerKey               types.String `tfsdk:"apns_issuer_key"`
	ApnsTopicHeader             types.String `tfsdk:"apns_topic_header"`
	ApnsCertificateConfigured   types.Bool   `tfsdk:"apns_certificate_configured"`
	ApnsSigningKeyConfigured    types.Bool   `tfsdk:"apns_signing_key_configured"`
	Created                     types.String `tfsdk:"created"`
	Modified                    types.String `tfsdk:"modified"`
}

// formatTimestamp converts a Unix timestamp in milliseconds to an RFC3339 string.
// Returns an empty string if the timestamp is zero.
func formatTimestamp(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format("2006-01-02T15:04:05.000Z07:00")
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
					DefaultBoolAttribute(types.BoolValue(true)),
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
			"apns_auth_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The APNS authentication type. Can be 'certificate' or 'token'.",
				Validators: []validator.String{
					stringvalidator.OneOf("certificate", "token"),
				},
			},
			"apns_signing_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The APNS signing key used for token-based authentication.",
			},
			"apns_signing_key_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The APNS signing key ID used for token-based authentication.",
			},
			"apns_issuer_key": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The APNS issuer key (Team ID) used for token-based authentication.",
			},
			"apns_topic_header": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The APNS topic header, typically the app bundle ID.",
			},
			"apns_certificate_configured": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether an APNS certificate has been configured.",
			},
			"apns_signing_key_configured": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether an APNS signing key has been configured.",
			},
			"fcm_service_account_configured": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether a Firebase Cloud Messaging service account has been configured.",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the app was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"modified": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the app was last modified.",
			},
		},
		MarkdownDescription: "The `ably_app` resource allows you to create and manage Ably Apps " +
			"and configure Ably Push notifications. Read more about Ably Push Notifications in Ably documentation: https://ably.com/docs/general/push",
	}
}

func (r ResourceApp) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "ably_app"
}

// buildAppState reconciles plan/state input with an API response into an AblyAppState.
// For Create/Update, pass the plan as input. For Read, pass the prior state.
func buildAppState(rc *reconciler, input AblyAppState, api control.AppResponse) AblyAppState {
	return AblyAppState{
		AccountID:                   rc.str("account_id", input.AccountID, types.StringValue(api.AccountID), true),
		ID:                          rc.str("id", input.ID, types.StringValue(api.ID), true),
		Name:                        rc.str("name", input.Name, types.StringValue(api.Name), false),
		Status:                      rc.str("status", input.Status, types.StringValue(api.Status), true),
		TLSOnly:                     rc.boolean("tls_only", input.TLSOnly, optBoolValue(api.TLSOnly), true),
		FcmKey:                      rc.str("fcm_key", input.FcmKey, types.StringNull(), false),
		FcmServiceAccount:           rc.str("fcm_service_account", input.FcmServiceAccount, types.StringNull(), false),
		FcmProjectId:                rc.str("fcm_project_id", input.FcmProjectId, optStringValue(api.FCMProjectID), false),
		FcmServiceAccountConfigured: rc.boolean("fcm_service_account_configured", input.FcmServiceAccountConfigured, optBoolValue(api.FCMServiceAccountConfigured), true),
		ApnsCertificate:             rc.str("apns_certificate", input.ApnsCertificate, types.StringNull(), false),
		ApnsPrivateKey:              rc.str("apns_private_key", input.ApnsPrivateKey, types.StringNull(), false),
		ApnsUseSandboxEndpoint:      rc.boolean("apns_use_sandbox_endpoint", input.ApnsUseSandboxEndpoint, optBoolValue(api.APNSUseSandboxEndpoint), true),
		ApnsAuthType:                rc.str("apns_auth_type", input.ApnsAuthType, optStringValue(api.APNSAuthType), true),
		ApnsSigningKey:              rc.str("apns_signing_key", input.ApnsSigningKey, types.StringNull(), false),
		ApnsSigningKeyId:            rc.str("apns_signing_key_id", input.ApnsSigningKeyId, optStringValue(api.APNSSigningKeyID), true),
		ApnsIssuerKey:               rc.str("apns_issuer_key", input.ApnsIssuerKey, optStringValue(api.APNSIssuerKey), true),
		ApnsTopicHeader:             rc.str("apns_topic_header", input.ApnsTopicHeader, optStringValue(api.APNSTopicHeader), true),
		ApnsCertificateConfigured:   rc.boolean("apns_certificate_configured", input.ApnsCertificateConfigured, optBoolValue(api.APNSCertificateConfigured), true),
		ApnsSigningKeyConfigured:    rc.boolean("apns_signing_key_configured", input.ApnsSigningKeyConfigured, optBoolValue(api.APNSSigningKeyConfigured), true),
		Created:                     rc.str("created", input.Created, types.StringValue(formatTimestamp(api.Created)), true),
		Modified:                    rc.str("modified", input.Modified, types.StringValue(formatTimestamp(api.Modified)), true),
	}
}

// Create creates a new resource.
func (r ResourceApp) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	// Gets plan values
	var plan AblyAppState
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generates an API request body from the plan values
	appValues := control.AppPost{
		Name:                   plan.Name.ValueString(),
		Status:                 plan.Status.ValueString(),
		TLSOnly:                optionalBoolPtr(plan.TLSOnly),
		FCMKey:                 optionalStringPtr(plan.FcmKey),
		FCMServiceAccount:      optionalStringPtr(plan.FcmServiceAccount),
		FCMProjectID:           optionalStringPtr(plan.FcmProjectId),
		APNSCertificate:        optionalStringPtr(plan.ApnsCertificate),
		APNSPrivateKey:         optionalStringPtr(plan.ApnsPrivateKey),
		APNSUseSandboxEndpoint: optionalBoolPtr(plan.ApnsUseSandboxEndpoint),
		APNSAuthType:           optionalStringPtr(plan.ApnsAuthType),
		APNSSigningKey:         optionalStringPtr(plan.ApnsSigningKey),
		APNSSigningKeyID:       optionalStringPtr(plan.ApnsSigningKeyId),
		APNSIssuerKey:          optionalStringPtr(plan.ApnsIssuerKey),
		APNSTopicHeader:        optionalStringPtr(plan.ApnsTopicHeader),
	}

	// Creates a new Ably App by invoking the CreateApp function from the Client Library
	ablyApp, err := r.p.client.CreateApp(ctx, r.p.accountID, appValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating ably_app",
			"Could not create ably_app, unexpected error: "+err.Error(),
		)
		return
	}

	// Read back the resource via GET to ensure computed fields like `modified`
	// reflect the settled server state (the POST response may return a value
	// that the server updates asynchronously).
	apps, err := r.p.client.ListApps(ctx, r.p.accountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading back ably_app after create",
			"Could not read back ably_app, unexpected error: "+err.Error(),
		)
		return
	}
	found := false
	for _, a := range apps {
		if a.ID == ablyApp.ID {
			ablyApp = a
			found = true
			break
		}
	}
	if !found {
		resp.Diagnostics.AddError(
			"Error reading back ably_app after create",
			fmt.Sprintf("Created app %s was not found in the app list read-back", ablyApp.ID),
		)
		return
	}

	// Maps response body to resource schema attributes.
	rc := newReconciler(&resp.Diagnostics)
	respApps := buildAppState(rc, plan, ablyApp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sets state for the new Ably App.
	diags = resp.State.Set(ctx, respApps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read reads the resource.
func (r ResourceApp) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	var state AblyAppState
	found := false
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID value for the resource
	appID := state.ID.ValueString()

	// Fetches all Ably Apps in the account. The function invokes the Client Library ListApps() method.
	// NOTE: Control API & Client Lib do not currently support fetching single app given app id
	apps, err := r.p.client.ListApps(ctx, r.p.accountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading ably_app",
			"Could not read ably_app, unexpected error: "+err.Error(),
		)
		return
	}

	// Loops through apps and if account id matches, sets state.
	for _, v := range apps {
		if v.ID == appID {
			rc := newReconciler(&resp.Diagnostics).forRead()
			respApps := buildAppState(rc, state, v)
			if resp.Diagnostics.HasError() {
				return
			}
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
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	// Get plan values
	var plan AblyAppState
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AblyAppState
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

	// Instantiates struct of type control.AppPatch and sets values to output of plan
	appValues := control.AppPatch{
		Name:                   plan.Name.ValueString(),
		Status:                 plan.Status.ValueString(),
		TLSOnly:                optionalBoolPtr(plan.TLSOnly),
		FCMKey:                 optionalStringPtr(plan.FcmKey),
		FCMServiceAccount:      optionalStringPtr(plan.FcmServiceAccount),
		FCMProjectID:           optionalStringPtr(plan.FcmProjectId),
		APNSCertificate:        optionalStringPtr(plan.ApnsCertificate),
		APNSPrivateKey:         optionalStringPtr(plan.ApnsPrivateKey),
		APNSUseSandboxEndpoint: optionalBoolPtr(plan.ApnsUseSandboxEndpoint),
		APNSAuthType:           optionalStringPtr(plan.ApnsAuthType),
		APNSSigningKey:         optionalStringPtr(plan.ApnsSigningKey),
		APNSSigningKeyID:       optionalStringPtr(plan.ApnsSigningKeyId),
		APNSIssuerKey:          optionalStringPtr(plan.ApnsIssuerKey),
		APNSTopicHeader:        optionalStringPtr(plan.ApnsTopicHeader),
	}

	// Updates an Ably App. The function invokes the Client Library UpdateApp method.
	ablyApp, err := r.p.client.UpdateApp(ctx, appID, appValues)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating ably_app",
			"Could not update ably_app, unexpected error: "+err.Error(),
		)
		return
	}

	// Read back via GET to get settled computed fields.
	apps, err := r.p.client.ListApps(ctx, r.p.accountID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading back ably_app after update",
			"Could not read back ably_app, unexpected error: "+err.Error(),
		)
		return
	}
	foundInList := false
	for _, a := range apps {
		if a.ID == ablyApp.ID {
			ablyApp = a
			foundInList = true
			break
		}
	}
	if !foundInList {
		resp.Diagnostics.AddError(
			"Error reading back ably_app after update",
			fmt.Sprintf("Updated app %s was not found in the app list read-back", ablyApp.ID),
		)
		return
	}

	rc := newReconciler(&resp.Diagnostics)
	respApps := buildAppState(rc, plan, ablyApp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Sets state to new app.
	diags = resp.State.Set(ctx, respApps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource.
func (r ResourceApp) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.p.ensureConfigured(&resp.Diagnostics) {
		return
	}

	// Get current state
	var state AblyAppState
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the current state. If it is unable to, the provider responds with an error.
	appID := state.ID.ValueString()

	err := r.p.client.DeleteApp(ctx, appID)
	if err != nil {
		if is404(err) {
			resp.Diagnostics.AddWarning(
				"Resource does not exist",
				"Resource does not exist, it may have already been deleted: "+err.Error(),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error deleting ably_app",
				"Could not delete ably_app, unexpected error: "+err.Error(),
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
