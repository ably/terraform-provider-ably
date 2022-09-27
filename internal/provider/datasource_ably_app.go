package ably_control

import (
	"context"

	tfsdk_datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfsdk_provider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type datasourceAppType struct{}

// Get App DataSource schema
func (r datasourceAppType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Required:    true,
				Description: "The application ID.",
			},
			"account_id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The ID of your Ably account.",
			},
			"name": {
				Type:        types.StringType,
				Computed:    true,
				Description: "The application name.",
			},
			"status": {
				Type:     types.StringType,
				Computed: true,
				// TODO: Update this after Control API bug has been fixed.
				Description: "The application status. Disabled applications will not accept new connections and will return an error to all clients. When creating a new application, ensure that its status is set to enabled.",
			},
			"tls_only": {
				Type:        types.BoolType,
				Computed:    true,
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
		MarkdownDescription: "The `ably_app` datasource allows you to fetch information for specific Ably Apps ",
	}, nil
}

// New datasource instance
func (c datasourceAppType) NewDataSource(_ context.Context,
	p tfsdk_provider.Provider) (tfsdk_datasource.DataSource, diag.Diagnostics) {
	return datasourceApp{
		p: *(p.(*provider)),
	}, nil
}

type datasourceApp struct {
	p provider
}

type appDataSourceData struct {
	AccountID              types.String `tfsdk:"account_id"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Status                 types.String `tfsdk:"status"`
	TLSOnly                types.Bool   `tfsdk:"tls_only"`
	FcmKey                 types.String `tfsdk:"fcm_key"`
	ApnsCertificate        types.String `tfsdk:"apns_certificate"`
	ApnsPrivateKey         types.String `tfsdk:"apns_private_key"`
	ApnsUseSandboxEndpoint types.Bool   `tfsdk:"apns_use_sandbox_endpoint"`
}

func (d datasourceApp) Read(ctx context.Context, req tfsdk_datasource.ReadRequest, resp *tfsdk_datasource.ReadResponse) {
	var data appDataSourceData

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Gets the Ably App ID value for the resource
	app_id := data.ID.Value

	// Fetches all Ably Apps in the account. The function invokes the Client Library Apps() method.
	// NOTE: Control API & Client Lib do not currently support fetching single app given app id
	apps, err := d.p.client.Apps()
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
				ID:                     types.String{Value: v.ID},
				AccountID:              types.String{Value: v.AccountID},
				Name:                   types.String{Value: v.Name},
				Status:                 types.String{Value: v.Status},
				TLSOnly:                types.Bool{Value: v.TLSOnly},
				FcmKey:                 types.String{Value: v.FcmKey},
				ApnsCertificate:        types.String{Value: v.ApnsCertificate},
				ApnsPrivateKey:         types.String{Value: v.ApnsPrivateKey},
				ApnsUseSandboxEndpoint: types.Bool{Value: v.ApnsUseSandboxEndpoint},
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
